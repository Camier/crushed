#!/usr/bin/env bash
set -euo pipefail

# install-user-config.sh
# Installs a sample user config that resolves secrets via pass (with env fallbacks).
# - Installs to $XDG_CONFIG_HOME/crush/crush.json (or $HOME/.config/crush/crush.json)
# - Will not overwrite an existing file unless --force is supplied
#
# Usage:
#   scripts/install-user-config.sh [--force] [--dry-run]
#
FORCE=0
DRY_RUN=0
if [[ ${1:-} == "--force" || ${2:-} == "--force" ]]; then FORCE=1; fi
if [[ ${1:-} == "--dry-run" || ${2:-} == "--dry-run" ]]; then DRY_RUN=1; fi

ROOT_DIR=$(cd "$(dirname "$0")/.." && pwd)
SRC="$ROOT_DIR/docs/examples/user-config.pass.json"
CONF_DIR="${XDG_CONFIG_HOME:-$HOME/.config}/crush"
DEST="$CONF_DIR/crush.json"

if [[ ! -f "$SRC" ]]; then
  echo "ERROR: template not found: $SRC" >&2
  exit 1
fi

if [[ -f "$DEST" && $FORCE -eq 0 ]]; then
  echo "User config already exists: $DEST" >&2
  echo "Use --force to overwrite." >&2
  exit 0
fi

echo "Installing user config to: $DEST"
if [[ $DRY_RUN -eq 1 ]]; then
  echo "[dry-run] mkdir -p $CONF_DIR && cp $SRC $DEST"
  exit 0
fi
mkdir -p "$CONF_DIR"
cp -f "$SRC" "$DEST"
echo "Done. You can edit: $DEST"
