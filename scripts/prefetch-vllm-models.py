#!/usr/bin/env python3
"""Prefetch the Hugging Face models used by the vLLM preset.

This downloads model snapshots into the local Hugging Face cache so the vLLM
server can start without re-fetching large weights every time.

Usage:
  export HF_TOKEN=...  # if the repos are gated
  export HF_HOME=$HOME/.cache/huggingface  # optional custom cache path
  pip install huggingface_hub
  python scripts/prefetch-vllm-models.py
"""

from __future__ import annotations

import os
from typing import Iterable

from huggingface_hub import snapshot_download

VLLM_MODELS: Iterable[str] = (
    "meta-llama/Meta-Llama-3-8B-Instruct",
    "deepseek-ai/DeepSeek-Coder-V2-Lite-Instruct",
    "mistralai/Mixtral-8x7B-Instruct-v0.1",
)


def prefetch(repo_id: str) -> None:
    cache_dir = os.environ.get("HF_HOME")
    kwargs = {
        "repo_id": repo_id,
        "resume_download": True,
        "local_files_only": False,
    }
    if cache_dir:
        kwargs["cache_dir"] = cache_dir
    token = os.environ.get("HF_TOKEN")
    if token:
        kwargs["token"] = token

    print(f"→ Downloading {repo_id} ...", flush=True)
    snapshot_download(**kwargs)
    print(f"✓ {repo_id} cached", flush=True)


def main() -> None:
    for repo in VLLM_MODELS:
        prefetch(repo)


if __name__ == "__main__":
    main()
