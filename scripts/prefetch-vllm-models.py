#!/usr/bin/env python3
"""Prefetch local copies of the vLLM models and keep them on disk.

The script also creates convenient aliases so you can reference models from
stable filesystem locations in your provider startup command.

Usage::

    # optional but recommended
    export HF_HOME=$HOME/.cache/huggingface
    export HF_TOKEN=...  # for gated repos
    # override where aliases are created (defaults to ~/.local/share/crush/models)
    export CRUSH_LOCAL_MODELS=$HOME/.local/share/crush/models

    pip install huggingface_hub
    python scripts/prefetch-vllm-models.py
"""

from __future__ import annotations

import os
import sys
from dataclasses import dataclass
from pathlib import Path
from typing import Iterable

from huggingface_hub import snapshot_download


class PrefetchError(SystemExit):
    pass

@dataclass(frozen=True)
class Model:
    repo_id: str
    alias: str


VLLM_MODELS: Iterable[Model] = (
    Model("NousResearch/Nous-Hermes-2-Mistral-7B-DPO", "nous-hermes-7b"),
    Model("Open-Orca/Mistral-7B-OpenOrca", "openorca-7b"),
    Model("deepseek-ai/DeepSeek-Coder-V2-Lite-Instruct", "deepseek-coder"),
)

_LOCAL_MODELS_ENV = os.environ.get("CRUSH_LOCAL_MODELS")
if _LOCAL_MODELS_ENV:
    LOCAL_MODELS_ROOT = Path(os.path.expanduser(_LOCAL_MODELS_ENV))
else:
    LOCAL_MODELS_ROOT = Path.home() / ".local" / "share" / "crush" / "models"

CACHE_ROOT = LOCAL_MODELS_ROOT / "vllm"


def ensure_local_root() -> None:
    try:
        LOCAL_MODELS_ROOT.mkdir(parents=True, exist_ok=True)
    except OSError as err:
        raise PrefetchError(
            f"Unable to create model directory {LOCAL_MODELS_ROOT}: {err}. "
            "Check that the volume is mounted and writable, or set CRUSH_LOCAL_MODELS "
            "to a different location."
        )


def prefetch(model: Model) -> Path:
    cache_dir_env = os.environ.get("HF_HOME")
    kwargs = {
        "repo_id": model.repo_id,
        "resume_download": True,
        "local_files_only": False,
    }
    if cache_dir_env:
        kwargs["cache_dir"] = cache_dir_env
    if token := os.environ.get("HF_TOKEN"):
        kwargs["token"] = token

    print(f"→ Downloading {model.repo_id}")
    path = Path(snapshot_download(**kwargs))
    print(f"✓ Cached at {path}")
    return path


def ensure_alias(model: Model, target: Path) -> None:
    CACHE_ROOT.mkdir(parents=True, exist_ok=True)
    link_path = CACHE_ROOT / model.alias

    if link_path.exists() or link_path.is_symlink():
        if link_path.is_symlink():
            try:
                link_path.unlink()
            except OSError as err:
                print(f"! Failed to remove existing alias {link_path}: {err}", file=sys.stderr)
                print(f"  Use the cached snapshot directly: {target}", file=sys.stderr)
                return
        else:
            print(
                f"! Path {link_path} already exists and is not a symlink; "
                "skipping alias creation.",
                file=sys.stderr,
            )
            print(f"  Use the cached snapshot directly: {target}", file=sys.stderr)
            return

    try:
        if os.name == "nt":
            os.symlink(str(target), str(link_path), target_is_directory=True)
        else:
            link_path.symlink_to(target, target_is_directory=True)
    except OSError as err:
        print(f"! Failed to create alias {link_path}: {err}", file=sys.stderr)
        print(f"  Use the cached snapshot directly: {target}", file=sys.stderr)
        return

    print(f"✓ Alias {link_path} → {target}")


def main() -> None:
    ensure_local_root()
    created = []
    for model in VLLM_MODELS:
        snapshot_path = prefetch(model)
        ensure_alias(model, snapshot_path)
        created.append((model.alias, snapshot_path))

    if created:
        print("\nModel aliases ready:")
        for alias, _ in created:
            print(f"- {alias}: {CACHE_ROOT / alias}")
        sample = CACHE_ROOT / created[0][0]
        print(
            "\nUse these paths in your vLLM startup command, e.g.:\n"
            f"  --model \"{sample}\""
        )


if __name__ == "__main__":
    try:
        main()
    except PrefetchError as exc:
        print(f"! {exc}", file=sys.stderr)
        sys.exit(1)
