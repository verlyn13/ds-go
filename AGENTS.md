# Repository Guidelines

## Project Structure & Modules
- `cmd/ds/main.go` — CLI entry (Cobra).
- `internal/` — core packages:
  - `config` (XDG config, init), `scan` (discovery, indexing, organize/clone), `git` (repo ops), `ui` (table/JSON).
- `scripts/` — helper scripts (e.g., `validate-compliance.sh`).
- Root: `Makefile`, `.mise.toml` task runner, `.golangci.yml` linters. Built binary: `./ds`.

## Build, Test, and Development
- Build: `make build` (or `mise run build`) → creates `./ds`.
- Install: `make install` (moves to `/usr/local/bin`).
- Run dev: `make run ARGS="status -d"` or `go run ./cmd/ds status`.
- Test: `make test` (=`go test -v -race -cover ./...`).
- Lint/Format: `make lint`, `make fmt` (gofmt + gofumpt).
- CI locally: `mise run ci` (lint, test, build). List tasks: `mise tasks`.

## Coding Style & Naming
- Go 1.25; gofmt/goimports enforced; tabs; max line length ~120 (see `lll`).
- Use `golangci-lint` (errcheck, gosec, revive, staticcheck, etc.).
- Package names: short, lowercase. Exported identifiers need doc comments starting with the name.
- Errors: wrap with `fmt.Errorf("context: %w", err)`; avoid naked returns; prefer explicit contexts.
- Keep functions small; avoid magic numbers (see `gomnd`).

## Testing Guidelines
- Standard library `testing`; table-driven tests preferred.
- Place tests alongside code as `*_test.go`. Names: `TestXxx`, `BenchmarkXxx`.
- Run with coverage and race: `go test -race -cover ./...`.
- Avoid network/FS side-effects; use temp dirs and small fixtures.

## Commit & Pull Request Guidelines
- Commits: imperative mood, concise subject (≤72 chars), optional body. Example: `Add organize command for repo folders`.
- Reference issues: `Fixes #123` when applicable.
- PRs: clear description, rationale, test plan (sample CLI output), and linked issues. Must pass `mise run ci`.

## Security & Configuration Tips
- Do not commit secrets. Use sample values only.
- Config path: `~/.config/ds/config.yaml` (created by `ds init`). SSH hosts should match account/org entries in `~/.ssh/config`.

## Architecture Overview
- Cobra-based CLI. Scanner uses concurrency (`errgroup`, `semaphore`) to walk `BaseDir` and build an index.
- Persists `.ds-index.json` and `.ds-fetch-cache.json` in the base directory; UI supports table and `--json` output.

