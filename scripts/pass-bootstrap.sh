#!/usr/bin/env bash
set -euo pipefail

# pass-bootstrap.sh
# Helps populate pass entries Crush is configured to use.
# - Does NOT print secrets.
# - Uses `pass insert -m` so you paste tokens securely.
#
# Usage:
#   scripts/pass-bootstrap.sh [--force]
#
# Options:
#   --force   Re-insert even if an entry already exists
#

FORCE=0
if [[ ${1:-} == "--force" ]]; then
  FORCE=1
fi

need() { command -v "$1" >/dev/null 2>&1 || { echo "ERROR: '$1' not found in PATH" >&2; exit 1; }; }

need pass
need gpg

# Ensure gpg-agent uses current tty (avoids pinentry issues)
export GPG_TTY="${GPG_TTY:-$(tty || true)}"

STORE_DIR="${PASSWORD_STORE_DIR:-$HOME/.password-store}"
if [[ ! -f "$STORE_DIR/.gpg-id" ]]; then
  echo "ERROR: pass store not initialized (missing $STORE_DIR/.gpg-id)" >&2
  echo "Run: pass init YOUR_GPG_FINGERPRINT" >&2
  exit 1
fi

# List of entries to populate (path | friendly label)
ENTRIES=(
  "mcp/github/pat|GitHub MCP token (PAT)"
  "services/mcp/github-token|GitHub MCP token (services)"
  "services/mcp/context7/token|Context7 MCP token"
  "services/mcp/openai-token|OpenAI API key"
  "services/mcp/anthropic-token|Anthropic API key"
  "services/mcp/openrouter-token|OpenRouter API key"
  "services/mcp/perplexity-token|Perplexity API key"
  "services/mcp/groq-token|Groq API key"
)

maybe_insert() {
  local path="$1"; shift
  local label="$1"; shift
  local exists=0
  if pass show "$path" >/dev/null 2>&1; then
    exists=1
  fi
  if [[ $exists -eq 1 && $FORCE -eq 0 ]]; then
    echo "[skip] $label — already present at: $path"
    return 0
  fi
  echo "[insert] $label"
  echo "Paste the secret for $path, then press Ctrl+D"
  pass insert -m ${FORCE:+-f} "$path"
}

cat <<HDR
Crush pass bootstrap
- Store: $STORE_DIR
- GPG_TTY: ${GPG_TTY:-unset}

This will prompt you to paste tokens for the entries Crush is configured to read
(from .crush.json). You can re-run with --force to overwrite existing entries.
HDR

for item in "${ENTRIES[@]}"; do
  IFS='|' read -r path label <<<"$item"
  while true; do
    read -r -p "Populate $label? [Y/n] " ans
    ans=${ans:-Y}
    case "$ans" in
      [Yy]*) maybe_insert "$path" "$label"; break;;
      [Nn]*) echo "[skip] $label"; break;;
      *) echo "Please answer Y or n.";;
    esac
  done
  echo
done

echo "Done. You can now run: task build && ./bin/crush"
