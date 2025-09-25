#!/usr/bin/env bash
set -euo pipefail

# pass-consolidate.sh
# Consolidate/migrate Crush-related pass entries to canonical paths.
# - Never prints secret values
# - Supports dry-run and force modes
#
# Canonical destinations:
#   services/mcp/github-token
#   services/mcp/context7/token
#   services/mcp/openai-token
#   services/mcp/anthropic-token
#   services/mcp/openrouter-token
#   services/mcp/perplexity-token
#   services/mcp/groq-token
#   services/mcp/sourcery-token
#
# Usage:
#   scripts/pass-consolidate.sh [--dry-run] [--force] [--delete-sources] [--reencrypt]
#
# Options:
#   --dry-run         Print planned actions only
#   --force           Overwrite destination if it already exists
#   --delete-sources  Remove old source entries after successful copy
#   --reencrypt       Run `pass reencrypt -p services/mcp` at the end
#

need() { command -v "$1" >/dev/null 2>&1 || { echo "ERROR: '$1' not found in PATH" >&2; exit 1; }; }
need pass
need gpg

DRY_RUN=0
FORCE=0
DELETE_SOURCES=0
REENCRYPT=0

while [[ $# -gt 0 ]]; do
  case "$1" in
    --dry-run) DRY_RUN=1; shift ;;
    --force) FORCE=1; shift ;;
    --delete-sources) DELETE_SOURCES=1; shift ;;
    --reencrypt) REENCRYPT=1; shift ;;
    *) echo "Unknown option: $1" >&2; exit 1 ;;
  esac
done

export GPG_TTY="${GPG_TTY:-$(tty || true)}"
STORE_DIR="${PASSWORD_STORE_DIR:-$HOME/.password-store}"
if [[ ! -f "$STORE_DIR/.gpg-id" ]]; then
  echo "ERROR: pass store not initialized (missing $STORE_DIR/.gpg-id)" >&2
  echo "Run: pass init YOUR_GPG_FINGERPRINT" >&2
  exit 1
fi

# dest | src1 src2 src3 ...
MAPPINGS=(
  "services/mcp/github-token|mcp/github/pat mcp/github/token"
  "services/mcp/context7/token|mcp/context7/api-key"
  "services/mcp/openai-token|mcp/openai/api-key"
  "services/mcp/anthropic-token|mcp/anthropic/api-key"
  "services/mcp/openrouter-token|mcp/openrouter/api-key"
  "services/mcp/perplexity-token|mcp/perplexity/api-key"
  "services/mcp/groq-token|mcp/groq/api-key"
  "services/mcp/sourcery-token|mcp/sourcery/token"
)

has_entry() { pass show "$1" >/dev/null 2>&1; }

copy_entry() {
  local src="$1" dest="$2"
  if [[ $DRY_RUN -eq 1 ]]; then
    echo "[copy] $src -> $dest"
    return 0
  fi
  pass show "$src" | pass insert -m ${FORCE:+-f} "$dest"
}

remove_entry() {
  local path="$1"
  if [[ $DRY_RUN -eq 1 ]]; then
    echo "[rm]   $path"
    return 0
  fi
  pass rm -f "$path" >/dev/null
}

process_mapping() {
  local dest="$1"; shift
  local sources=("$@")

  local dest_exists=0
  has_entry "$dest" && dest_exists=1

  if [[ $dest_exists -eq 1 && $FORCE -eq 0 ]]; then
    echo "[keep] $dest (exists)"
    # Optionally clean dupes with identical content
    for src in "${sources[@]}"; do
      if has_entry "$src"; then
        if [[ $DRY_RUN -eq 1 ]]; then
          echo "[dupe?] $src (consider removing after verifying)"
        else
          # Compare by SHA-256 without printing secrets
          local sha_src sha_dst
          sha_src=$(pass show "$src" | sha256sum | awk '{print $1}')
          sha_dst=$(pass show "$dest" | sha256sum | awk '{print $1}')
          if [[ "$sha_src" == "$sha_dst" && $DELETE_SOURCES -eq 1 ]]; then
            remove_entry "$src"
          fi
        fi
      fi
    done
    return 0
  fi

  # Find first available source
  local picked=""
  for src in "${sources[@]}"; do
    if has_entry "$src"; then
      picked="$src"; break
    fi
  done

  if [[ -z "$picked" ]]; then
    echo "[skip] $dest (no sources found)"
    return 0
  fi

  copy_entry "$picked" "$dest"
  echo "[ok]   $picked -> $dest"

  if [[ $DELETE_SOURCES -eq 1 ]]; then
    for src in "${sources[@]}"; do
      [[ "$src" == "$picked" ]] || has_entry "$src" || continue
      remove_entry "$src"
    done
  fi
}

main() {
  echo "Consolidating Crush pass entries"
  echo "Store: $STORE_DIR"
  echo "Options: dry_run=$DRY_RUN force=$FORCE delete_sources=$DELETE_SOURCES reencrypt=$REENCRYPT"
  echo

  local dest srcs line
  for line in "${MAPPINGS[@]}"; do
    dest=${line%%|*}
    # shellcheck disable=SC2206
    srcs=(${line#*|})
    process_mapping "$dest" "${srcs[@]}"
  done

  if [[ $REENCRYPT -eq 1 ]]; then
    if [[ $DRY_RUN -eq 1 ]]; then
      echo "[reencrypt] pass reencrypt -p services/mcp"
    else
      pass reencrypt -p services/mcp
    fi
  fi

  echo "Done."
}

main "$@"
