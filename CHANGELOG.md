# Changelog

All notable changes to this project will be documented in this file.

## v0.2.0 — 2025-09-25

Highlights
- LSP
  - New CLI commands: `crush lsp list|enable|disable|test`.
  - Header details: compact LSP summary (active/total) + per‑LSP status list (✓ found, ⚠ missing, off when disabled).
  - `doctor lsp` diagnostics with optional version check (`CRUSH_LSP_VERSION_CHECK=1`).
- MCP
  - New `crush doctor mcp` to verify stdio/http entries (command presence, URL, auth header presence, quick reachability).
- Editor
  - External editor launch now uses an injected runner for unit testing; robust quoting/splitting.
  - Tests for VISUAL/EDITOR parsing, defaults, and error paths.
- Providers & Tokens
  - `.crush.json` resolves tokens via pass (with env fallbacks) for: OpenAI, Anthropic, OpenRouter, Perplexity, Groq, GitHub MCP, Context7 MCP.
  - Added helper: `task pass:bootstrap` to populate pass entries securely.
- TUI & Tests
  - Broader chat/header/splash golden snapshots; width variants and focus/overlay states.
  - Stability improvements and targeted `-race` runs for stable packages.
- CI & Builds
  - Added `Lint` workflow (golangci-lint) for PRs.
  - PR snapshot artifacts and nightly snapshot workflow.
  - Minimal cross‑build script (`scripts/build-snapshot.sh`) with `VERSION` stamping.

Notes
- Windows packaging/targets were removed earlier; POSIX targets remain.
- See the release page for binaries and checksums.

