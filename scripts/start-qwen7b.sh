#!/usr/bin/env bash
set -euo pipefail

MODEL_PATH="${MODEL_PATH:-$HOME/.local/share/llama.cpp/models/Qwen2.5-7B-Instruct-Q4_K_M.gguf}"
PORT="${PORT:-8081}"
THREADS="${THREADS:-$(nproc)}"
CONTAINER_NAME="${CONTAINER_NAME:-llama7b-local}"

if [ ! -f "$MODEL_PATH" ]; then
  echo "Model not found: $MODEL_PATH" >&2
  echo "Download it first, e.g.:" >&2
  echo "  curl -L -C - -o $HOME/.local/share/llama.cpp/models/Qwen2.5-7B-Instruct-Q4_K_M.gguf \\"
  echo "    'https://huggingface.co/bartowski/Qwen2.5-7B-Instruct-GGUF/resolve/main/Qwen2.5-7B-Instruct-Q4_K_M.gguf?download=true'" >&2
  exit 1
fi

podman rm -f "$CONTAINER_NAME" >/dev/null 2>&1 || true
podman run -d --name "$CONTAINER_NAME" \
  -p "${PORT}:8080" \
  -v "$(dirname "$MODEL_PATH"):/models:Z" \
  ghcr.io/ggerganov/llama.cpp:server \
  -m "/models/$(basename "$MODEL_PATH")" \
  -c 32768 -t "$THREADS" --host 0.0.0.0 --port 8080

echo "Started llama.cpp (Qwen 7B) on http://127.0.0.1:${PORT}/v1/"