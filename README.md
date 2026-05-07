# devstrap

Bootstrap and update your development tools with a single command.

devstrap downloads official releases, verifies SHA-256 checksums, and keeps everything up-to-date.

## Quick Start

```sh
curl -sSfL https://raw.githubusercontent.com/alecerf/devstrap/trunk/install.sh | sh
devstrap index update
devstrap tool install --all
```

Run `devstrap tool install --all` any time to upgrade everything to the latest versions.

## Installation

```sh
curl -sSfL https://raw.githubusercontent.com/alecerf/devstrap/trunk/install.sh | sh
```

The binary is installed to `~/.local/bin`. Make sure it's in your `PATH`.

| Variable      | Description                           | Default    |
| ------------- | ------------------------------------- | ---------- |
| `INSTALL_DIR` | Base directory (`bin` goes inside it) | `~/.local` |
| `VERSION`     | Pin a specific devstrap version       | latest     |

```sh
curl -sSfL https://raw.githubusercontent.com/alecerf/devstrap/trunk/install.sh | INSTALL_DIR=/usr/local sudo sh
```

## Commands

```
devstrap
├── tool
│   ├── install      Install or upgrade development tools
│   ├── list         List tools and their installed versions
│   └── uninstall    Remove installed tools
├── index
│   ├── update       Fetch the latest tool index
│   ├── list         List all tools in the index
│   └── search       Search tools by name or description
├── env              Print PATH exports for your shell
├── version          Print the devstrap version
└── completion       Generate shell completions (zsh, bash, fish, powershell)
```

### `devstrap tool install`

```sh
devstrap tool install --all
devstrap tool install go node
devstrap tool install go@1.22.0
devstrap tool install --all --dry-run
```

| Flag        | Description                             |
| ----------- | --------------------------------------- |
| `--all`     | Install every tool defined in the index |
| `--dry-run` | Check for upgrades without installing   |

Pin a version with `@version` (e.g. `go@1.22.0`). `--all` and `@version` cannot be combined.

### `devstrap tool list`

Show all tools and their installed versions.

### `devstrap tool uninstall`

```sh
devstrap tool uninstall terraform go
```

### `devstrap index update`

Fetch the latest tool definitions from the [index repository](https://github.com/alecerf/devstrap-index). Required before first install.

### `devstrap index list` / `devstrap index search`

```sh
devstrap index list
devstrap index search lint
```

### `devstrap env` / `devstrap completion`

Add to your `~/.zshrc`:

```sh
source <(devstrap env)
source <(devstrap completion zsh)
```

### Global Flags

| Flag         | Short | Description                       | Default                   |
| ------------ | ----- | --------------------------------- | ------------------------- |
| `--data-dir` |       | Directory for tool installations  | `~/.local/share/devstrap` |
| `--bin-dir`  |       | Directory for standalone binaries | `~/.local/bin`            |
| `--verbose`  | `-v`  | Show detailed output              |                           |

## License

[MIT](LICENSE)
