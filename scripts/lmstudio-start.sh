#!/usr/bin/env bash
set -euo pipefail

# LM Studio startup helper
# - Tries to start an OpenAI-compatible API server if LM Studio CLI/app is present
# - Otherwise prints clear instructions
#
# Environment variables:
#   LM_STUDIO_BASE_URL   Default: http://127.0.0.1:1234/v1
#   LM_STUDIO_STARTUP_COMMAND  If set, this command will be executed directly
#
# Usage:
#   scripts/lmstudio-start.sh

BASE_URL="${LM_STUDIO_BASE_URL:-http://127.0.0.1:1234/v1}"

if [ -n "${LM_STUDIO_STARTUP_COMMAND:-}" ]; then
  echo "Executing LM_STUDIO_STARTUP_COMMAND…"
  exec bash -lc "$LM_STUDIO_STARTUP_COMMAND"
fi

# Try common launch patterns
if command -v lmstudio >/dev/null 2>&1; then
  echo "Found 'lmstudio' CLI. Please start the API server from LM Studio UI or specify LM_STUDIO_STARTUP_COMMAND."
  echo "Expected OpenAI endpoint: $BASE_URL"
  exit 0
fi

# macOS: try to open the app
if [[ "${OSTYPE:-}" == darwin* ]]; then
  if [ -d "/Applications/LM Studio.app" ]; then
    echo "Opening LM Studio.app…"
    open -a "LM Studio"
    echo "Once running, enable the local server in LM Studio and set base URL to: $BASE_URL"
    exit 0
  fi
fi

echo "LM Studio CLI/app not detected. Options:"
echo "- Install LM Studio and enable the local OpenAI server (then use $BASE_URL)"
echo "- Or set LM_STUDIO_STARTUP_COMMAND to a command that starts an OpenAI-compatible server"
