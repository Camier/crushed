#!/usr/bin/env bash
set -euo pipefail

# Start vLLM (Dolphin Llama3 ZH) on 127.0.0.1:8002 with Turing-safe defaults.

MODEL_PATH="/run/media/miko/AYA/ai-models/vllm/dolphin-llama3-zh"
NAME="dolphin-llama3-zh"
HOST="127.0.0.1"
PORT="8002"

export VLLM_USE_V1=1
exec python -m vllm.entrypoints.openai.api_server \
  --model "$MODEL_PATH" \
  --served-model-name "$NAME" \
  --host "$HOST" --port "$PORT" \
  --dtype float16 --gpu-memory-utilization 0.80 \
  --max-model-len 3072 --max-num-batched-tokens 1024 --max-num-seqs 4

