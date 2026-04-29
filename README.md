# devstrap

Bootstrap and update your development tools with a single command.

devstrap downloads official releases, verifies SHA-256 checksums, and keeps everything up-to-date — so you don't have to.

## Quick Start

```sh
# 1. Install devstrap
curl -sSfL https://raw.githubusercontent.com/alecerf/devstrap/trunk/install.sh | sh

# 2. Fetch the tool index (required on first run)
devstrap index update

# 3. Install all available tools
devstrap tool install --all
```

That's it. Run `devstrap tool install --all` any time to upgrade everything to the latest versions.

## Installation

```sh
curl -sSfL https://raw.githubusercontent.com/alecerf/devstrap/trunk/install.sh | sh
```

The binary is installed to `~/.local/bin` by default. Make sure it's in your `PATH`:

```sh
export PATH="$HOME/.local/bin:$PATH"
```

### Installation Options

| Variable      | Description                           | Default       |
| ------------- | ------------------------------------- | ------------- |
| `INSTALL_DIR` | Base directory (`bin` goes inside it) | `~/.local`    |
| `VERSION`     | Pin a specific devstrap version       | latest        |

```sh
# Install to /usr/local/bin (requires sudo)
curl -sSfL https://raw.githubusercontent.com/alecerf/devstrap/trunk/install.sh | INSTALL_DIR=/usr/local sudo sh

# Install a specific version
curl -sSfL https://raw.githubusercontent.com/alecerf/devstrap/trunk/install.sh | VERSION=0.1.0 sh
```

## Commands

```
devstrap
├── tool
│   ├── install      Install or upgrade development tools
│   └── list         List tools and their installed versions
├── index
│   ├── update       Fetch the latest tool index
│   ├── list         List all tools in the index
│   └── search       Search tools by name or description
├── env              Print PATH exports for your shell
├── update           Update devstrap to the latest version
├── version          Print the devstrap version
└── completion       Generate shell completions (zsh, bash, fish, powershell)
```

### `devstrap tool install`

Install or upgrade one or more development tools.

```sh
# Install or upgrade all tools to their latest versions
devstrap tool install --all

# Install specific tools only
devstrap tool install go node

# Install a specific version of a tool
devstrap tool install go --version 1.22.0

# Dry run — see what would be upgraded without changing anything
devstrap tool install --all --dry-run
```

| Flag        | Description                                      |
| ----------- | ------------------------------------------------ |
| `--all`     | Install every tool defined in the index           |
| `--dry-run` | Check for upgrades without installing              |
| `--version` | Pin a specific version (requires exactly one tool) |

> **Note:** `--all` and `--version` cannot be used together. Same for `--dry-run` and `--version`.

### `devstrap tool list`

Show all available tools and their currently installed versions.

```sh
devstrap tool list
```

Tools that aren't installed yet are shown as "not installed".

### `devstrap index update`

Fetch the latest tool definitions from the remote [index repository](https://github.com/alecerf/devstrap-index).

```sh
devstrap index update
```

Run this before your first install and whenever you want to pick up newly added tools.

### `devstrap index list`

List every tool available in your local index cache.

```sh
devstrap index list
```

### `devstrap index search`

Search for tools by name or description.

```sh
devstrap index search lint
devstrap index search node
```

### `devstrap env`

Print a shell `export PATH` statement that includes all directories managed by devstrap.

```sh
devstrap env
```

Add this to your `~/.zshrc` so tools are always on your `PATH`:

```sh
source <(devstrap env)
```

### `devstrap update`

Update the devstrap CLI itself to the latest version.

```sh
devstrap update
```

Downloads the latest release from GitHub, verifies its checksum, and atomically replaces the running binary.

### `devstrap completion`

Generate shell completion scripts. Supports zsh, bash, fish, and powershell.

```sh
# Generate zsh completions (source in your ~/.zshrc)
source <(devstrap completion zsh)

# Generate bash completions
source <(devstrap completion bash)
```

### `devstrap version`

Print the installed devstrap version.

```sh
devstrap version
```

### Global Flags

These flags can be used with any `tool` subcommand:

| Flag         | Description                             | Default                          |
| ------------ | --------------------------------------- | -------------------------------- |
| `--data-dir` | Directory for tool installations         | `~/.local/share/devstrap`        |
| `--bin-dir`  | Directory for standalone binaries        | `~/.local/bin`                   |

```sh
devstrap tool install --all --data-dir ~/dev --bin-dir ~/dev/bin
```

## Shell Integration

devstrap installs tools into their own directories. To use them, add these lines to your `~/.zshrc`:

```sh
# PATH for devstrap-managed tools
source <(devstrap env)

# Shell completions (tab-complete commands and tool names)
source <(devstrap completion zsh)
```

`source <(...)` uses zsh process substitution — it's equivalent to `eval "$(…)"` but cleaner.

This exports `PATH` entries for `~/.local/bin` and each tool's bin directory, and enables tab completion for all devstrap commands.

## How It Works

1. **Index** — Tool definitions live in a separate [index repository](https://github.com/alecerf/devstrap-index). Run `devstrap index update` to cache them locally.
2. **Version discovery** — devstrap queries official APIs (JSON endpoints or GitHub releases) to find the latest version of each tool.
3. **Download & verify** — Archives are downloaded and verified against SHA-256 checksums before extraction.
4. **Install** — Tools are extracted to `~/.local/share/devstrap/<tool>/` (directory-mode) or copied to `~/.local/bin/` (standalone binaries).

devstrap respects XDG directories:

| Variable         | Used for            | Default                    |
| ---------------- | ------------------- | -------------------------- |
| `XDG_DATA_HOME`  | Tool installations  | `~/.local/share/devstrap`  |
| `XDG_CACHE_HOME` | Index cache         | `~/.cache/devstrap`        |

## Platforms

| OS    | Architecture |
| ----- | ------------ |
| macOS | amd64, arm64 |
| Linux | amd64, arm64 |

## Adding a New Tool

Tools are defined as JSON files in the [devstrap-index](https://github.com/alecerf/devstrap-index) repository. No code changes to devstrap are needed — just add a definition and run `devstrap index update` to pick it up.

See the existing definitions in the index repo for examples.

## License

[MIT](LICENSE)
