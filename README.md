# devstrap

Bootstrap and update your development tools with a single command.

## Overview

devstrap is a CLI tool that installs and keeps development tools up-to-date by downloading official releases. It verifies SHA-256 checksums on every download and supports both macOS and Linux.

Tool definitions are maintained in a separate [index repository](https://github.com/alecerf/devstrap-index) so that new tools can be added without releasing a new devstrap binary.

## Supported Tools

Tools are defined in the [devstrap-index](https://github.com/alecerf/devstrap-index) repository.
Run `devstrap index list` to see all available tools after updating the index.

## Platforms

| OS    | Architecture    |
| ----- | --------------- |
| macOS | amd64, arm64    |
| Linux | amd64, arm64    |

## Installation

```sh
curl -sSfL https://raw.githubusercontent.com/alecerf/devstrap/trunk/install.sh | sh
```

By default the binary is installed to `~/.local/bin`. Override with environment variables:

```sh
# Install to /usr/local/bin (requires sudo)
curl -sSfL https://raw.githubusercontent.com/alecerf/devstrap/trunk/install.sh | INSTALL_DIR=/usr/local sudo sh

# Install a specific version
curl -sSfL https://raw.githubusercontent.com/alecerf/devstrap/trunk/install.sh | VERSION=0.1.0 sh
```

## Usage

```sh
# Fetch the tool index (required on first run)
devstrap index update

# Check for available upgrades
devstrap tool update

# Upgrade all tools to their latest versions
devstrap tool upgrade

# Upgrade specific tools only
devstrap tool upgrade go node

# List installed tools and their versions
devstrap tool list

# Print devstrap version
devstrap version
```

### Index Management

```sh
# Update the local index cache from the remote repository
devstrap index update

# List all tools available in the index
devstrap index list

# Search for tools by name or description
devstrap index search lint
```

Tools are installed into `~/Workspace/<tool>/` by default. Use the `--base` flag to change the base directory:

```sh
devstrap tool upgrade --base ~/dev
```

## Adding a New Tool

To add a new tool to devstrap, create a JSON definition file in the [devstrap-index](https://github.com/alecerf/devstrap-index) repository. See the existing definitions for examples.

## License

This project is licensed under the [MIT License](LICENSE).
