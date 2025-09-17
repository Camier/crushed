#!/usr/bin/env bash
set -euo pipefail

# Start a vLLM OpenAI-compatible server for Meta-Llama 3 8B Instruct.
# Requires NVIDIA GPU drivers + Docker with GPU support, or adjust for CPU.

MODEL_ID="${MODEL_ID:-meta-llama/Meta-Llama-3-8B-Instruct}"
PORT="${PORT:-8000}"
CONTAINER_NAME="${CONTAINER_NAME:-vllm-llama3-8b}" 

# Optional: pass your HF token for gated models
# export HF_TOKEN=...  # or set as env before running this script

echo "Starting vLLM (model: ${MODEL_ID}) on http://127.0.0.1:${PORT}/v1 ..."

# Prefer NVIDIA GPUs if available
GPU_ARGS=()
if command -v nvidia-smi >/dev/null 2>&1; then
  GPU_ARGS=(--gpus all)
fi

# Mount huggingface cache to avoid re-downloading
HF_CACHE_HOST="${HF_CACHE_HOST:-$HOME/.cache/huggingface}"
mkdir -p "$HF_CACHE_HOST"

docker rm -f "$CONTAINER_NAME" >/dev/null 2>&1 || true
docker run -d --name "$CONTAINER_NAME" \
  "${GPU_ARGS[@]}" \
  -e HF_TOKEN="${HF_TOKEN:-}" \
  -p "${PORT}:8000" \
  -v "${HF_CACHE_HOST}:/root/.cache/huggingface" \
  vllm/vllm-openai:latest \
  --model "$MODEL_ID" \
  --port 8000 --host 0.0.0.0

echo "vLLM is up. Test: curl http://127.0.0.1:${PORT}/v1/models"

