#!/usr/bin/env bash
set -euo pipefail

# Start vLLM (Nous Hermes 7B) on 127.0.0.1:8001 with Turing-safe defaults.

MODEL_PATH="/run/media/miko/AYA/ai-models/vllm/nous-hermes-7b"
NAME="nous-hermes-7b"
HOST="127.0.0.1"
PORT="8001"

export VLLM_USE_V1=1
exec python -m vllm.entrypoints.openai.api_server \
  --model "$MODEL_PATH" \
  --served-model-name "$NAME" \
  --host "$HOST" --port "$PORT" \
  --dtype float16 --gpu-memory-utilization 0.80 \
  --max-model-len 3072 --max-num-batched-tokens 1024 --max-num-seqs 4

