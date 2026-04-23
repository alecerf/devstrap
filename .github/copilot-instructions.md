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

devstrap is a CLI that bootstraps and updates development tools by downloading official releases. Tool definitions are stored in an external JSON index ([devstrap-index](https://github.com/alecerf/devstrap-index)) and cached locally. It has four internal packages:

- **`internal/tool`** — Defines the `Tool` interface (`Name`, `FetchLatest`, `CurrentVersion`, `Install`) and the orchestration functions `Run` (install/upgrade) and `Check` (dry-run).
- **`internal/index`** — Declarative tool index system. Loads JSON tool definitions from the local cache, implements a generic `tool.Tool` driven by those definitions. Handles version discovery (`json_api`, `github_release`), Go template rendering for URLs, platform arch/OS mapping, and index fetching/caching.
- **`internal/downloader`** — HTTP download, SHA-256 checksum verification, tar.gz extraction, and binary installation helpers. Generic `FetchJSON[T]` is used by the index package to query release APIs.
- **`internal/cli`** — Cobra commands (`update`, `upgrade`, `list`, `version`, `index`), the tool registry (loads from index), and terminal UI (spinner, colored output).

## Adding a New Tool

Add a JSON definition file to the [devstrap-index](https://github.com/alecerf/devstrap-index) repository. Each definition describes: version discovery (API URL + JSON path), download URL template, checksum strategy, installation mode (directory or binary), and version detection command. No Go code changes needed for tools that follow the supported patterns (`json_api` or `github_release` source types).

## Conventions

- All packages are under `internal/` — nothing is exported outside the module.
- Errors are defined as package-level sentinel `var`s and wrapped with `fmt.Errorf` + `%w`.
- Versions are plain semver strings without a `v` prefix throughout the codebase; the `v` is stripped at API boundaries.
- Downloads always verify SHA-256 checksums before extraction.
- Tools install into `<baseDir>/<tool>/` (default `~/Workspace`). Standalone binaries go into `<baseDir>/bin/`.
- The tool index is cached at `~/.devstrap/index/` and must be manually updated with `devstrap index update`.
