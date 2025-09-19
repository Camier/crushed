#!/usr/bin/env bash
set -euo pipefail

DEFAULT_MODEL_ROOT="${CRUSH_LOCAL_MODELS:-$HOME/.local/share/crush/models}"
MODEL_FILE="Qwen2.5-7B-Instruct-Q4_K_M.gguf"
MODEL_PATH="${MODEL_PATH:-${DEFAULT_MODEL_ROOT}/llama.cpp/models/${MODEL_FILE}}"
MODEL_DIR="$(dirname "$MODEL_PATH")"
PORT="${PORT:-8081}"
THREADS="${THREADS:-$(nproc)}"
CONTAINER_NAME="${CONTAINER_NAME:-llama7b-local}"

if [ ! -f "$MODEL_PATH" ]; then
  echo "Model not found: $MODEL_PATH" >&2
  echo "Download it first, e.g.:" >&2
  echo "  mkdir -p \"$MODEL_DIR\"" >&2
  echo "  curl -L -C - -o \"$MODEL_PATH\" \\" >&2
  echo "    'https://huggingface.co/bartowski/Qwen2.5-7B-Instruct-GGUF/resolve/main/Qwen2.5-7B-Instruct-Q4_K_M.gguf?download=true'" >&2
  exit 1
fi

podman rm -f "$CONTAINER_NAME" >/dev/null 2>&1 || true
podman run -d --name "$CONTAINER_NAME" \
  -p "${PORT}:8080" \
  -v "$MODEL_DIR:/models:Z" \
  ghcr.io/ggerganov/llama.cpp:server \
  -m "/models/$(basename "$MODEL_PATH")" \
  -c 32768 -t "$THREADS" --host 0.0.0.0 --port 8080

echo "Started llama.cpp (Qwen 7B) on http://127.0.0.1:${PORT}/v1/"
