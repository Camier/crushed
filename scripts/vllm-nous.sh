#!/usr/bin/env bash
set -euo pipefail

# Start vLLM for Nous Hermes 7B with configurable defaults.

MODEL_ROOT="${CRUSH_LOCAL_MODELS:-$HOME/.local/share/crush/models}"
MODEL_PATH="${MODEL_PATH:-${MODEL_ROOT}/vllm/nous-hermes-7b}"
MODEL_NAME="${MODEL_NAME:-nous-hermes-7b}"
HOST="${HOST:-127.0.0.1}"
PORT="${PORT:-8001}"
VLLM_PYTHON="${VLLM_PYTHON:-python}"
DTYPE="${DTYPE:-float16}"
GPU_MEMORY_UTILIZATION="${GPU_MEMORY_UTILIZATION:-0.80}"
MAX_MODEL_LEN="${MAX_MODEL_LEN:-3072}"
MAX_NUM_BATCHED_TOKENS="${MAX_NUM_BATCHED_TOKENS:-1024}"
MAX_NUM_SEQS="${MAX_NUM_SEQS:-4}"

if [ ! -d "$MODEL_PATH" ]; then
  echo "Model directory not found: $MODEL_PATH" >&2
  exit 1
fi

export VLLM_USE_V1=1
exec "$VLLM_PYTHON" -m vllm.entrypoints.openai.api_server \
  --model "$MODEL_PATH" \
  --served-model-name "$MODEL_NAME" \
  --host "$HOST" --port "$PORT" \
  --dtype "$DTYPE" \
  --gpu-memory-utilization "$GPU_MEMORY_UTILIZATION" \
  --max-model-len "$MAX_MODEL_LEN" \
  --max-num-batched-tokens "$MAX_NUM_BATCHED_TOKENS" \
  --max-num-seqs "$MAX_NUM_SEQS"
