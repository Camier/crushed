# Repository Guidelines

Follow this guide when contributing features, fixes, or experiments to the Crush CLI/TUI.

## Project Structure & Module Organization
- `main.go` is the entry point; all reusable code lives under `internal/` (e.g., `app`, `config`, `tui/*`, `llm`, `shell`, `db`).
- Database migrations belong in `internal/db/migrations/` with goose headers; SQLC output stays in `internal/db/*.sql.go`.
- Tests sit beside their subjects as `*_test.go`; golden files live with the TUI experiments (see `internal/tui/exp/**/*`).
- Configuration schema is `schema.json`; user/system configs are `.crush.json`, `crush.json`, or `$HOME/.config/crush/crush.json`.

## Build, Test, and Development Commands
- `task bootstrap` installs formatters and linters; run it after cloning or when tooling drifts.
- `task build` compiles the CLI to `./bin/crush`; `task install` installs it into your Go bin.
- `task dev` runs `go run .` with profiling flags for iterative work.
- `task test [ARGS]` wraps `go test ./...`; pass `-run` filters or `-update` to refresh golden fixtures.
- `task lint` and `task lint-fix` execute `golangci-lint` per `.golangci.yml`; `task fmt` applies `gofumpt`/`goimports`.
- `task build:release` calls `goreleaser build --snapshot` for packaging checks.

## Coding Style & Naming Conventions
- Target Go 1.25 with tab indentation; avoid manual formatting—always run `task fmt`.
- Keep package and file names short and lowercase; exported identifiers use PascalCase with doc comments, internals use camelCase.
- Do not add license headers or reformat untouched files; respect `.crushignore` and `.gitignore`.

## Testing Guidelines
- Use the standard library `testing`, `testify`, and Charmbracelet golden helpers; keep cases deterministic and skip OS-specific ones when necessary.
- Name tests `TestThing` alongside the code; use `go test ./internal/tui -run TestList` for focused runs.
- Update golden data with `go test ./... -update` or `task -d internal/tui/exp/diffview test:update`.

## Commit & Pull Request Guidelines
- Follow Conventional Commits (e.g., `feat: add keybinding`, `fix: handle empty query`).
- Link issues in PR descriptions (`Fixes #123`), note behavioral changes, and attach screenshots or recordings for TUI updates.
- Ensure CI passes (build + lint); sign the CLA before requesting review.

## Security & Configuration Tips
- Never commit secrets; rely on environment variables such as `OPENAI_API_KEY`.
- Logs write to `./.crush/logs/crush.log`; keep them out of Git history.

## Agent-Specific Notes
- Keep diffs focused and minimal; prefer composing changes that pass `task fmt` and `task lint` before submission.
- Do not revert user edits you did not create; investigate unexpected workspace changes before proceeding.
- Use `crush doctor lsp` to spot missing language servers and follow the install hints (e.g., `go install golang.org/x/tools/gopls@latest`).
