#!/usr/bin/env python3
"""Prefetch local copies of the vLLM models and keep them on disk.

Usage::

    # optional but recommended
    export HF_HOME=$HOME/.cache/huggingface
    export HF_TOKEN=...  # for gated repos

    pip install huggingface_hub
    python scripts/prefetch-vllm-models.py
"""

from __future__ import annotations

import os
from pathlib import Path
from typing import Iterable

from huggingface_hub import snapshot_download

VLLM_MODELS: Iterable[str] = (
    "meta-llama/Meta-Llama-3-8B-Instruct",
    "deepseek-ai/DeepSeek-Coder-V2-Lite-Instruct",
    "mistralai/Mixtral-8x7B-Instruct-v0.1",
)


def prefetch(repo_id: str) -> Path:
    cache_dir_env = os.environ.get("HF_HOME")
    kwargs = {
        "repo_id": repo_id,
        "resume_download": True,
        "local_files_only": False,
    }
    if cache_dir_env:
        kwargs["cache_dir"] = cache_dir_env
    if token := os.environ.get("HF_TOKEN"):
        kwargs["token"] = token

    print(f"→ Downloading {repo_id}")
    path = Path(snapshot_download(**kwargs))
    print(f"✓ Cached at {path}")
    return path


def main() -> None:
    for repo in VLLM_MODELS:
        prefetch(repo)


if __name__ == "__main__":
    main()
