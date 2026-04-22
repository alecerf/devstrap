# Copilot Instructions for devstrap

## Build & Lint

```sh
go build ./...
golangci-lint run          # linter config: .golangci.yml (v2, all linters enabled minus a short deny-list)
```

Version is injected at build time:

```sh
go build -ldflags "-X github.com/alecerf/devstrap/internal/cli.Version=1.0.0"
```

There are no tests yet.

**After every implementation change**, run `golangci-lint run --fix` and fix all reported issues. Never add `//nolint` directives unless the user explicitly asks for it.

## Architecture

devstrap is a CLI that bootstraps and updates development tools (Go, Node.js, golangci-lint) by downloading official releases. It has three internal packages:

- **`internal/tool`** — Defines the `Tool` interface (`Name`, `FetchLatest`, `CurrentVersion`, `Install`) and the orchestration functions `Run` (install/upgrade) and `Check` (dry-run). Every installer implements this interface.
- **`internal/tool/{golang,node,github}`** — Concrete installers. `golang` and `node` are standalone packages. `github` is a reusable struct for any GitHub-release binary; `golangci.go` configures it for golangci-lint via `NewGolangCI`.
- **`internal/downloader`** — HTTP download, SHA-256 checksum verification, tar.gz extraction, and binary installation helpers. Generic `FetchJSON[T]` is used by all installers to query release APIs.
- **`internal/cli`** — Cobra commands (`update`, `upgrade`, `list`, `version`), the tool registry, and terminal UI (spinner, colored output).

## Adding a New Tool

1. Create a package under `internal/tool/` (or reuse `github.Tool` for GitHub-release binaries).
2. Implement `tool.Tool` — the factory signature is `func New(baseDir string, plat tool.Platform) tool.Tool`.
3. Register it in `internal/cli/registry.go` by appending to the `registry` slice.

## Conventions

- All packages are under `internal/` — nothing is exported outside the module.
- Tool constructors return `tool.Tool` (interface) and use `//nolint:ireturn` to suppress the linter.
- Errors are defined as package-level sentinel `var`s and wrapped with `fmt.Errorf` + `%w`.
- Versions are plain semver strings without a `v` prefix throughout the codebase; the `v` is stripped at API boundaries.
- Downloads always verify SHA-256 checksums before extraction.
- Tools install into `<baseDir>/<tool>/` (default `~/Workspace`). Standalone binaries go into `<baseDir>/bin/`.
