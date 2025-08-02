# Repository Guidelines

## Project Structure & Module Organization
- `main.go` wires the CLI entrypoint through dependency injection in `internal/di`.
- `cmd/` contains Cobra command implementations and CLI-facing tests; `internal/` holds services, infra (Gist), TUI, validation, and clipboard helpers.
- `integration/` houses tagged integration tests; `test_prompts/` stores sample prompt fixtures; `scripts/test-tui.sh` runs the interactive TUI check.

## Build, Test, and Development Commands
- `make build` (or `go build -o pv .`) produces the `pv` binary; `make run` builds then runs it.
- `make test` runs unit tests with race detection when `gcc` is available; `make test-quick` skips race and coverage.
- `make test-integration` runs integration tests (`-tags=integration`, 60s timeout); `make test-all` aggregates unit, delete-focused, TUI, and integration suites and merges coverage.
- `make fmt` applies `gofmt`/`goimports`; `make lint` runs `go vet` and `golangci-lint` when installed.
- `make generate` regenerates Wire DI code; `make release` builds multi-platform artifacts under `build/`; `make clean` prunes build and coverage outputs.

## Coding Style & Naming Conventions
- Go defaults: tabs for indentation, camel-case identifiers, exported symbols commented when needed. Keep packages lower_snake and focused by domain (`service`, `infra`, `tui`).
- Run `make fmt` before PRs; keep imports organized (`goimports`) and prefer `context.Context` plumbing where applicable.
- CLI commands follow Cobra patterns in `cmd/` with short/long descriptions; ensure flags use meaningful names.

## Testing Guidelines
- Place unit tests adjacent to code (`*_test.go`); mirror command/service scenarios seen in `cmd/*_test.go` and `internal/*`.
- Use table-driven cases and prefer deterministic fixtures from `test_prompts/`.
- Integration tests require the `integration` tag and may touch the mocked store; they should remain network-free unless explicitly isolating Gist flows.
- For TUI regressions, run `scripts/test-tui.sh -c` for a clean start. Capture coverage via `make test` or `make test-all`.

## Commit & Pull Request Guidelines
- Commit history favors concise, present-tense summaries (often in Chinese, e.g., “增强 list 命令显示 prompt 导出状态”); keep scope focused per commit.
- PRs should include: purpose/impact summary, key commands run (`make test`, `make fmt/lint`, relevant integration suites), and notes on user-facing changes or new flags.
- Link related issues or TODOs when available; add screenshots only if a TUI change is visually relevant.

## Security & Configuration Tips
- Development auth uses a GitHub token with `gist` scope (`pv auth login`); never commit tokens or local state files.
- Keep tests and fixtures free of real Gist URLs or secrets; prefer generated IDs/URLs like those under `integration/`.
