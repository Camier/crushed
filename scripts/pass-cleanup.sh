#!/usr/bin/env bash
set -euo pipefail

# pass-cleanup.sh
# Cleans up legacy pass entries after consolidation and prunes empty dirs.
# It delegates migration/removal to pass-consolidate.sh with delete+reencrypt,
# then prunes now-empty legacy folders.
#
# Usage:
#   scripts/pass-cleanup.sh [--dry-run]
#
DRY_RUN=0
if [[ ${1:-} == "--dry-run" ]]; then DRY_RUN=1; fi

need() { command -v "$1" >/dev/null 2>&1 || { echo "ERROR: '$1' not found in PATH" >&2; exit 1; }; }
need pass
need gpg

ROOT_DIR=$(cd "$(dirname "$0")/.." && pwd)
STORE_DIR="${PASSWORD_STORE_DIR:-$HOME/.password-store}"
LEGACY_DIRS=(
  "$STORE_DIR/mcp/github"
  "$STORE_DIR/mcp/context7"
  "$STORE_DIR/mcp/openai"
  "$STORE_DIR/mcp/anthropic"
  "$STORE_DIR/mcp/openrouter"
  "$STORE_DIR/mcp/perplexity"
  "$STORE_DIR/mcp/groq"
  "$STORE_DIR/mcp/sourcery"
)

# 1) Migrate + delete legacy sources using consolidator
if [[ $DRY_RUN -eq 1 ]]; then
  echo "[dry-run] $ROOT_DIR/scripts/pass-consolidate.sh --delete-sources --reencrypt"
else
  "$ROOT_DIR/scripts/pass-consolidate.sh" --delete-sources --reencrypt
fi

# 2) Prune empty legacy directories
for d in "${LEGACY_DIRS[@]}"; do
  if [[ -d "$d" ]]; then
    # If directory contains only subdirs that are empty as well, prune recursively
    if [[ $DRY_RUN -eq 1 ]]; then
      # Show which dirs would be removed
      find "$d" -type d -empty -print | sed 's/^/[prune] /'
    else
      find "$d" -type d -empty -delete
    fi
  fi
done

echo "Cleanup complete."
