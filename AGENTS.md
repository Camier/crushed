# Repository Guidelines

## Project Structure & Module Organization
- Go CLI/TUI app. Entry point: `main.go`.
- Core packages under `internal/` (e.g., `app`, `tui/*`, `llm`, `config`, `shell`, `db`).
- Database migrations in `internal/db/migrations/`; SQLC‑generated code in `internal/db/*.sql.go`.
- Config schema lives in `schema.json`; project config is `.crush.json` or `crush.json` (user config: `$HOME/.config/crush/crush.json`).

## Build, Test, and Development Commands
- Prereq: install task: `go install github.com/go-task/task/v3/cmd/task@latest`.
- `task bootstrap` — install lint/format tooling (wraps `lint:install` and `fmt:install`).
- `task build` — build the `crush` binary into `./bin/crush`.
- `task clean` — remove `bin/` and `dist/` artifacts.
- `task dev` — run locally with profiling flags (`go run .`).
- `task test [ARGS]` — run tests (`go test ./...`).
- `task lint` / `task lint-fix` — run golangci‑lint (uses `.golangci.yml`) (install via `task lint:install`).
- `task fmt` — format with `gofumpt`.
- `task install` — install binary.
- `task build:release` — run `goreleaser build --snapshot` for local packaging.
- `task schema` — generate `schema.json`.
- CI: GitHub Actions runs build + lint on PRs; releases via GoReleaser.

## Coding Style & Naming Conventions
- Go 1.25. Use tabs; formatting enforced.
- Format with `gofumpt` and `goimports` (enforced via lint).
- Package and file names: short, lowercase (`internal/tui/exp/list/list.go`).
- Exported identifiers use `PascalCase` with doc comments; unexported use `camelCase`.

## Testing Guidelines
- Frameworks: `testing`, `testify`, golden (`github.com/charmbracelet/x/exp/golden`).
- Place tests alongside code as `*_test.go`.
- Run all: `task test`. Filter: `go test ./internal/tui -run TestList`.
- Keep tests deterministic; skip OS‑specific cases when needed (see `internal/shell/shell_test.go`).
- Update goldens: `go test ./... -update` or `task -d internal/tui/exp/diffview test:update`.

## Database Changes
- Migrations: add goose files in `internal/db/migrations/` named `YYYYMMDDHHMMSS_description.sql` with `-- +goose Up/Down`.
- Queries: edit `internal/db/sql/`; run `sqlc generate` (see `sqlc.yaml`).

## Commit & Pull Request Guidelines
- Commit style: Conventional Commits (e.g., `feat: …`, `fix: …`, `docs(readme): …`, `chore(deps): …`). Use imperative mood and concise scope.
- Link issues in PR descriptions (`Fixes #123`). Include screenshots or recordings for TUI changes.
- Sign the CLA (automated check). PRs must pass CI (lint + build).

## Security & Configuration Tips
- Never commit secrets or API keys. Prefer environment variables (e.g., `OPENAI_API_KEY`).
- Use `.crushignore` to exclude local files; `.gitignore` is respected.
- Logs write to `./.crush/logs/crush.log`; avoid committing logs.

## Agent‑Specific Notes
- Keep changes minimal and focused; update tests and docs when touching behavior.
- Do not add license headers or reformat unrelated files. Prefer `task fmt` and `task lint` before sending PRs.
