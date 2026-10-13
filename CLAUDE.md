# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & dev commands

```bash
task build               # go build -o pyinit .
task lint                # golangci-lint run ./...
task test                # go test ./...
task check               # lint + test
task clean               # remove the pyinit binary
task install-golangci-lint  # brew install golangci-lint (first-time setup)
```

Go binary path: `/usr/local/go/bin/go` (not on PATH by default in this environment).

## Architecture

The tool runs in four sequential phases defined in `main.go`:

1. **`internal/preflight`** — checks that `uv` and `task` are on PATH; if not, offers a `huh` confirm prompt and installs via Homebrew (macOS) or official curl scripts as fallback.

2. **`internal/tui`** — renders a three-group `charmbracelet/huh` form that collects raw `config.Input` values. It only builds the form; field validation is delegated to `internal/config`.

3. **`internal/config`** — `config.New(Input)` validates the raw input and returns a `ProjectConfig`, or an error. It owns project-name and Python-version validation, and derives `PackageName` (hyphens → underscores), `PythonVersionTag` (e.g. `"312"`) and the resolved `OutputDir` (empty → home, `~` expanded, made absolute). It is the unit-test seam: `go test ./internal/config`.

4. **`internal/scaffold`** — creates the project on disk from a `config.ProjectConfig`. `scaffold.Run(cfg, runner)` takes an `internal/runner.Runner` (the process seam): `runner.Exec{}` in production, `runner.Fake` in tests (`go test ./internal/scaffold`). Sequence: `mkdir` → `uv init` → move `main.py` into `src/<package>/` → write `__init__.py` / `tests/` / `docs/` → render four Go templates (overwriting uv's minimal `pyproject.toml`) → `uv venv`.

## Templates

All four templates live in `internal/scaffold/templates/` and are embedded into the binary at compile time via `//go:embed templates/*`. They use Go's `text/template` with `config.ProjectConfig` as the data object. **Editing a template requires a rebuild** — there are no runtime assets.

Key template fields: `.ProjectName`, `.PackageName`, `.Description`, `.AuthorName`, `.AuthorEmail`, `.PythonVersion`, `.PythonVersionTag`.

## Branch naming

Use `spp/` as the branch prefix (e.g. `spp/my-feature`).

## Linter config

golangci-lint v2 (`version: "2"` in `.golangci.yml`). Formatters (`gofmt`, `goimports`) are declared under `formatters:`, not `linters:` — this is a v2 requirement. `goimports` is configured with local prefix `github.com/scottp/pyinit`.

## Agent skills

### Issue tracker

Issues live in GitHub Issues for spowers42/pyinit (via the `gh` CLI). See `docs/agents/issue-tracker.md`.

### Triage labels

Default five-label vocabulary (`needs-triage`, `needs-info`, `ready-for-agent`, `ready-for-human`, `wontfix`). See `docs/agents/triage-labels.md`.

### Domain docs

Single-context: one `CONTEXT.md` + `docs/adr/` at the repo root. See `docs/agents/domain.md`.
