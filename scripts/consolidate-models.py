#!/usr/bin/env python3
"""
Consolidate local model storage under a shared root.

Creates canonical symlinks for:
- vLLM aliases (Transformers snapshots) under <ROOT>/vllm
- llama.cpp GGUF files under <ROOT>/gguf

Writes a manifest to <ROOT>/models-manifest.json listing discovered repos and links.

ROOT is resolved from CRUSH_LOCAL_MODELS or defaults to ~/.local/share/crush/models.

Idempotent and safe: re-links only when necessary.
"""
from __future__ import annotations

import json
import os
import sys
import time
from pathlib import Path
from typing import Optional


def get_root() -> Path:
    root = os.environ.get("CRUSH_LOCAL_MODELS")
    if root:
        return Path(os.path.expanduser(root))
    return Path.home() / ".local" / "share" / "crush" / "models"


def resolve_snapshot(repo_dir: Path) -> Optional[Path]:
    main_ref = repo_dir / "refs" / "main"
    if main_ref.exists():
        try:
            return main_ref.resolve()
        except OSError:
            return main_ref
    snaps_dir = repo_dir / "snapshots"
    if not snaps_dir.exists():
        return None
    snaps = [p for p in snaps_dir.iterdir() if p.is_dir()]
    if not snaps:
        return None
    snaps.sort(key=lambda p: p.stat().st_mtime, reverse=True)
    return snaps[0]


KNOWN_ALIASES = {
    ("Open-Orca", "Mistral-7B-OpenOrca"): "openorca-7b",
    ("NousResearch", "Nous-Hermes-2-Mistral-7B-DPO"): "nous-hermes-7b",
    ("deepseek-ai", "DeepSeek-Coder-V2-Lite-Instruct"): "deepseek-coder",
}


def alias_name(org: str, repo: str) -> str:
    return KNOWN_ALIASES.get((org, repo)) or f"hf-{org}-{repo}".lower()


def main() -> int:
    root = get_root()
    hf = root / "hf-home"
    vllm = root / "vllm"
    gguf = root / "gguf"
    vllm.mkdir(parents=True, exist_ok=True)
    gguf.mkdir(parents=True, exist_ok=True)

    manifest = {
        "generated_at": time.strftime("%Y-%m-%d %H:%M:%S"),
        "root": str(root),
        "hf_repos": [],
        "vllm_aliases": [],
        "gguf_files": [],
    }

    # vLLM aliases
    if hf.exists():
        for repo_dir in hf.glob("models--*--*"):
            if not repo_dir.is_dir():
                continue
            parts = repo_dir.name.split("--", 2)
            if len(parts) < 3:
                continue
            _, org, repo = parts
            snap = resolve_snapshot(repo_dir)
            manifest["hf_repos"].append(
                {
                    "org": org,
                    "repo": repo,
                    "path": str(repo_dir),
                    "snapshot": str(snap) if snap else None,
                }
            )
            if not snap:
                continue
            # Check if this looks like a Transformers repo
            has_cfg = (snap / "config.json").exists()
            has_tok = any((snap / f).exists() for f in ("tokenizer.json", "tokenizer.model", "tokenizer_config.json"))
            has_weights = any(snap.glob("*.safetensors")) or any(snap.glob("pytorch_model*.bin"))
            if has_cfg and has_tok and has_weights:
                name = alias_name(org, repo)
                link = vllm / name
                try:
                    if link.exists() or link.is_symlink():
                        try:
                            cur = link.resolve()
                        except OSError:
                            cur = None
                        if str(cur) != str(snap):
                            link.unlink()
                            link.symlink_to(snap, target_is_directory=True)
                    else:
                        link.symlink_to(snap, target_is_directory=True)
                    manifest["vllm_aliases"].append({"alias": name, "target": str(snap)})
                except OSError as e:
                    manifest["vllm_aliases"].append({"alias": name, "target": str(snap), "error": str(e)})

    # GGUF consolidation — scan likely roots and link into gguf/
    def add_root(candidate: Optional[Path]) -> None:
        if candidate and candidate.exists() and candidate not in search_roots:
            search_roots.append(candidate)

    search_roots: list[Path] = []
    add_root(hf)
    add_root(root)
    parent = root.parent if root.parent != root else None
    add_root(parent)
    add_root(Path.home() / "models")

    extra_roots = os.environ.get("CRUSH_MODEL_SEARCH_PATHS", "")
    if extra_roots:
        for part in extra_roots.split(os.pathsep):
            part = part.strip()
            if part:
                add_root(Path(os.path.expanduser(part)))

    seen: set[str] = set()
    for base in search_roots:
        if not base.exists():
            continue
        for p in base.rglob("*.gguf"):
            try:
                rp = p.resolve()
            except OSError:
                rp = p
            if str(rp) in seen:
                continue
            seen.add(str(rp))
            link = gguf / p.name.lower()
            i = 2
            while link.exists() and link.resolve() != rp:
                stem = link.stem
                ext = link.suffix
                link = gguf / f"{stem}-{i}{ext}"
                i += 1
            try:
                if link.exists() or link.is_symlink():
                    if link.resolve() != rp:
                        link.unlink()
                        link.symlink_to(rp)
                else:
                    link.symlink_to(rp)
                manifest["gguf_files"].append({"alias": link.name, "target": str(rp)})
            except OSError as e:
                manifest["gguf_files"].append({"alias": link.name, "target": str(rp), "error": str(e)})

    (root / "models-manifest.json").write_text(json.dumps(manifest, indent=2))
    print("Manifest written:", root / "models-manifest.json")
    print("vLLM aliases:", ", ".join(a["alias"] for a in manifest["vllm_aliases"]) or "(none)")
    print("GGUF files:", len(manifest["gguf_files"]))
    return 0


if __name__ == "__main__":
    sys.exit(main())

