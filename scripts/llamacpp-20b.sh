#!/usr/bin/env bash
set -euo pipefail

# Start llama.cpp server for GPT-OSS 20B on 127.0.0.1:8088 with tuned flags.

MODEL_GGUF="/run/media/miko/AYA/ai-models/gguf/gpt-oss-20b-iq4_nl.gguf"
HOST="127.0.0.1"
PORT="8088"

exec /home/miko/LAB/dev/llama.cpp/build/bin/llama-server \
  -m "$MODEL_GGUF" \
  -c 4096 --ubatch-size 256 -ngl 35 -t -1 \
  --host "$HOST" --port "$PORT"

