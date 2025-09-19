#!/usr/bin/env bash
set -euo pipefail

# Start vLLM (OpenOrca 7B) on 127.0.0.1:8000 with Turing-safe defaults.

MODEL_PATH="/run/media/miko/AYA/ai-models/vllm/openorca-7b"
NAME="openorca-7b"
HOST="127.0.0.1"
PORT="8000"

export VLLM_USE_V1=1
exec python -m vllm.entrypoints.openai.api_server \
  --model "$MODEL_PATH" \
  --served-model-name "$NAME" \
  --host "$HOST" --port "$PORT" \
  --dtype float16 --gpu-memory-utilization 0.80 \
  --max-model-len 3072 --max-num-batched-tokens 1024 --max-num-seqs 4

