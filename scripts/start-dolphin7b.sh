#!/usr/bin/env bash
set -euo pipefail

DEFAULT_MODEL_ROOT="${CRUSH_LOCAL_MODELS:-$HOME/.local/share/crush/models}"
MODEL_FILE="dolphin-2.9.3-mistral-7B-32k-Q4_K_M.gguf"
MODEL_PATH="${MODEL_PATH:-${DEFAULT_MODEL_ROOT}/llama.cpp/models/${MODEL_FILE}}"
PORT="${PORT:-8082}"
THREADS="${THREADS:-$(nproc)}"
CONTAINER_NAME="${CONTAINER_NAME:-llama7b-dolphin}"

if [ ! -f "$MODEL_PATH" ]; then
  echo "Model not found: $MODEL_PATH" >&2
  echo "Download it first, e.g.:" >&2
  echo "  mkdir -p \"$(dirname "$MODEL_PATH")\"" >&2
  echo "  curl -L -C - -o \"$MODEL_PATH\" \\" >&2
  echo "    'https://huggingface.co/bartowski/dolphin-2.9.3-mistral-7B-32k-GGUF/resolve/main/dolphin-2.9.3-mistral-7B-32k-Q4_K_M.gguf?download=true'" >&2
  exit 1
fi

podman rm -f "$CONTAINER_NAME" >/dev/null 2>&1 || true
podman run -d --name "$CONTAINER_NAME" \
  -p "${PORT}:8080" \
  -v "$(dirname "$MODEL_PATH"):/models:Z" \
  ghcr.io/ggerganov/llama.cpp:server \
  -m "/models/$(basename "$MODEL_PATH")" \
  -c 32768 -t "$THREADS" --host 0.0.0.0 --port 8080

echo "Started llama.cpp (Dolphin 7B) on http://127.0.0.1:${PORT}/v1/"