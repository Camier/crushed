#!/usr/bin/env bash
set -euo pipefail

# Start llama.cpp server for GPT-OSS 20B with configurable defaults.

MODEL_ROOT="${CRUSH_LOCAL_MODELS:-$HOME/.local/share/crush/models}"
MODEL_GGUF="${MODEL_GGUF:-${MODEL_ROOT}/gguf/gpt-oss-20b-iq4_nl.gguf}"
HOST="${HOST:-127.0.0.1}"
PORT="${PORT:-8088}"
THREADS="${THREADS:-$(nproc)}"
UBATCH_SIZE="${UBATCH_SIZE:-256}"
GPU_LAYERS="${GPU_LAYERS:-35}"
LLAMACPP_BIN="${LLAMACPP_BIN:-$(command -v llama-server || true)}"

if [ -z "$LLAMACPP_BIN" ]; then
  echo "Set LLAMACPP_BIN to the llama.cpp server binary or add it to PATH." >&2
  exit 1
fi

if [ ! -f "$MODEL_GGUF" ]; then
  echo "Model not found: $MODEL_GGUF" >&2
  exit 1
fi

exec "$LLAMACPP_BIN" \
  -m "$MODEL_GGUF" \
  -c 4096 --ubatch-size "$UBATCH_SIZE" -ngl "$GPU_LAYERS" -t "$THREADS" \
  --host "$HOST" --port "$PORT"
