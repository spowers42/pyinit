# pyinit

A CLI tool to bootstrap modern Python projects — interactively. Run it, answer a few questions, and walk away with a fully configured project using [uv](https://docs.astral.sh/uv/), [ruff](https://docs.astral.sh/ruff/), and [Task](https://taskfile.dev).

![pyinit demo](demo/demo.gif)

---

## What you get

Running `pyinit` creates a ready-to-go Python project with:

```
my-awesome-lib/
├── src/
│   └── my_awesome_lib/
│       ├── __init__.py
│       └── main.py
├── tests/
│   └── __init__.py
├── docs/
├── pyproject.toml   # hatchling build, ruff config, pytest dev dep
├── Taskfile.yml     # common dev tasks
├── .python-version
├── .gitignore
└── README.md
```

A virtual environment is created automatically via `uv venv`.

---

## Requirements

- [Go](https://go.dev/dl/) 1.23+ (only to build `pyinit` from source)
- [uv](https://docs.astral.sh/uv/getting-started/installation/) — Python package manager
- [Task](https://taskfile.dev/installation/) — task runner

`pyinit` will detect if `uv` or `task` are missing at startup and offer to install them for you.

---

## Installation

### Prebuilt binary (recommended)

Download the archive for your platform from [GitHub Releases](https://github.com/spowers42/pyinit/releases/latest) — Linux and macOS (amd64, arm64) and Windows (amd64). Verify it against the release's `pyinit_<version>_checksums.txt`, then put `pyinit` on your PATH:

```bash
VERSION=0.2.0   # latest release, without the leading "v"
ARCHIVE=pyinit_${VERSION}_linux_amd64.tar.gz   # or darwin_arm64, darwin_amd64, linux_arm64
BASE=https://github.com/spowers42/pyinit/releases/download/v${VERSION}

SUMS=pyinit_${VERSION}_checksums.txt

curl -LO "$BASE/$ARCHIVE" -LO "$BASE/$SUMS"
sha256sum --check --ignore-missing "$SUMS"   # macOS: shasum -a 256 --check --ignore-missing "$SUMS"
tar -xzf "$ARCHIVE" pyinit
mv pyinit /usr/local/bin/pyinit
```

On Windows, download `pyinit_<version>_windows_amd64.zip` and check it with `Get-FileHash -Algorithm SHA256` against the checksums file.

### `go install`

```bash
go install github.com/spowers42/pyinit@latest
```

### Build from source

```bash
git clone https://github.com/spowers42/pyinit.git
cd pyinit
go build -o pyinit .

# Optionally move to somewhere on your PATH
mv pyinit /usr/local/bin/pyinit
```

---

## Usage

```bash
pyinit             # start the interactive form
pyinit --version   # print the version and exit
```

The interactive form will ask for:

| Field | Required | Default |
|---|---|---|
| Project name | yes | — |
| Description | no | "" |
| Author name | no | "" |
| Author email | no | "" |
| Python version | no | `3.12` |
| Output directory | yes | `~/` |

Once submitted, `pyinit` will:

1. Create the project directory and src-layout structure
2. Run `uv init` to initialise the project
3. Move `main.py` into `src/<package>/`
4. Render `pyproject.toml`, `Taskfile.yml`, `README.md`, and `.gitignore` from templates
5. Run `uv venv` to create a virtual environment

---

## Generated Taskfile tasks

The generated project includes these tasks out of the box:

| Task | Description |
|---|---|
| `task onboard` | Install tools, create venv, install deps (full setup from scratch) |
| `task env` | Create virtual environment (`uv venv`) |
| `task install` | Install dependencies (`uv sync`) |
| `task lint` | Run ruff linter |
| `task format` | Format code with ruff |
| `task test` | Run pytest |
| `task check` | Lint + test |

---

## Development

```bash
task build    # build the pyinit binary
task lint     # run golangci-lint
task test     # run go test
task check    # lint + test (also tests scripts/next-version.sh)
task snapshot # build all release archives into dist/ (publishes nothing)
```

---

## Releasing

Maintainers cut releases with one command from an up-to-date `main` (needs an authenticated [`gh`](https://cli.github.com)):

```bash
task release -- patch   # or minor, or major
```

It prints the next version — computed from the latest tag that has a GitHub Release — then dispatches the Release workflow, follows it, and prints the release URL. The workflow runs `task check` on `main`, tags the commit only if checks pass, and publishes the binaries and checksums with [GoReleaser](https://goreleaser.com) and auto-generated release notes.

A `major` release asks you to type the computed version (e.g. `v1.0.0`) back to confirm; the workflow refuses a major bump whose confirmation doesn't match.

`task snapshot` builds every release archive locally into `dist/` without publishing anything.

Pushing a `v*` tag by hand does nothing: the workflow is the only release path (see [ADR 0001](docs/adr/0001-releases-only-via-dispatch-workflow.md)).
