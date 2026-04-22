# devstrap

Bootstrap and update your development tools with a single command.

## Overview

devstrap is a CLI tool that installs and keeps development tools up-to-date by downloading official releases. It verifies SHA-256 checksums on every download and supports both macOS and Linux.

## Supported Tools

| Tool           | Source                          |
| -------------- | ------------------------------- |
| Go             | go.dev official releases        |
| Node.js        | nodejs.org official releases    |
| golangci-lint  | GitHub releases                 |

## Platforms

| OS    | Architecture    |
| ----- | --------------- |
| macOS | amd64, arm64    |
| Linux | amd64, arm64    |

## Installation

```sh
curl -sSfL https://raw.githubusercontent.com/alecerf/devstrap/trunk/install.sh | sh
```

By default the binary is installed to `/usr/local/bin`. Override with environment variables:

```sh
# Install to ~/.local/bin
curl -sSfL https://raw.githubusercontent.com/alecerf/devstrap/trunk/install.sh | INSTALL_DIR=~/.local sh

# Install a specific version
curl -sSfL https://raw.githubusercontent.com/alecerf/devstrap/trunk/install.sh | VERSION=0.1.0 sh
```

## Usage

```sh
# Check for available upgrades
devstrap update

# Upgrade all tools to their latest versions
devstrap upgrade

# Upgrade specific tools only
devstrap upgrade go node

# List installed tools and their versions
devstrap list

# Print devstrap version
devstrap version
```

Tools are installed into `~/Workspace/<tool>/` by default. Use the `--base` flag to change the base directory:

```sh
devstrap upgrade --base ~/dev
```

## License

This project is licensed under the [MIT License](LICENSE).
