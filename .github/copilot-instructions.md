# Copilot Instructions for devstrap

## Build, Test & Lint

```sh
go build ./...
go test ./...
golangci-lint run          # linter config: .golangci.yml (v2, all linters enabled minus a short deny-list)
```

Version is injected at build time:

```sh
go build -ldflags "-X github.com/alecerf/devstrap/internal/cli.Version=1.0.0"
```

**After every implementation change**, run `golangci-lint run --fix` and fix all reported issues. Never add `//nolint` directives unless the user explicitly asks for it.

**Always format Go files** with `gofmt` or `goimports` after every code change.

**Format Markdown on every edit.** Whenever a `.md` file is created or modified, run `prettier --write <file>` to format it before committing.

**Keep this file up to date.** Whenever you add, rename, or remove packages, change the CLI command tree, modify public interfaces, or alter conventions, update the relevant sections of this file in the same commit. These instructions are the primary onboarding reference — they must always reflect the current codebase.

## Architecture

devstrap is a CLI that bootstraps and updates development tools by downloading official releases. Tool definitions are stored in an external JSON index ([devstrap-index](https://github.com/alecerf/devstrap-index)) and cached locally. It has five internal packages:

- **`internal/registry`** — Core package. Defines the `Tool` interface (`Name`, `FetchLatest`, `FetchVersion`, `CurrentVersion`, `Install`) and the `Installer` type that implements it from JSON `Definition`s. Provides orchestration functions `Run` (install/upgrade), `RunVersion` (exact version), and `Check` (dry-run). Handles version discovery (`json_api`, `github_release` source types), Go template rendering for URLs, platform/arch mapping, custom JSON path navigation, and index fetching/caching.
- **`internal/downloader`** — HTTP download, SHA-256 checksum verification, tar.gz extraction, and binary installation. Generic `FetchJSON[T]` is used by the registry package to query release APIs.
- **`internal/shell`** — Generates shell configuration (PATH export snippets) for devstrap-managed tool directories.
- **`internal/cli`** — Cobra root command, `Execute()` entry point, `Version` variable (build-time injected, default `"dev"`), and persistent flags (`--data-dir`, `--bin-dir`).
- **`internal/cli/tool`** — `tool install` and `tool list` subcommands. Resolves tool names from the index, validates flag combinations, and delegates to `registry.Run`/`RunVersion`.
- **`internal/cli/index`** — `index update`, `index list`, and `index search` subcommands.
- **`internal/cli/ui`** — Terminal output helpers: `Printer` (padded, colored output), `Spinner` (Braille animation), TTY detection, and color functions.
- **`internal/updater`** — Self-update logic for the devstrap binary. `FetchLatest` queries the GitHub API; `Update` downloads, verifies the SHA-256 checksum, and atomically replaces the running binary.

### CLI command tree

```
devstrap (root)
├── tool
│   ├── install <tools[@version]... | --all>  [--dry-run]
│   └── list
├── index
│   ├── update
│   ├── list
│   └── search <query>
├── env
├── update
├── version
└── completion
```

## Adding a New Tool

Add a JSON definition file to the [devstrap-index](https://github.com/alecerf/devstrap-index) repository. Each definition describes: version discovery (API URL + JSON path), download URL template, checksum strategy, installation mode (directory or binary), and version detection command. No Go code changes needed for tools that follow the supported patterns (`json_api` or `github_release` source types).

## Conventions

- All packages are under `internal/` — nothing is exported outside the module.
- Errors are defined as package-level sentinel `var`s and wrapped with `fmt.Errorf` + `%w`.
- Versions are plain semver strings without a `v` prefix throughout the codebase; the `v` is stripped at API boundaries.
- Downloads always verify SHA-256 checksums before extraction.
- Tools install into `<dataDir>/<tool>/` (default `~/.local/share/devstrap`). Standalone binaries go into `<binDir>/` (default `~/.local/bin`). Both respect XDG environment variables.
- The tool index is cached at `$XDG_CACHE_HOME/devstrap/` (default `~/.cache/devstrap`) and must be manually updated with `devstrap index update`.
- Test helpers in `export_test.go` expose private functions and sentinel errors for use in tests.
