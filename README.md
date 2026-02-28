# vc-env

A version manager for [vcluster](https://www.vcluster.com/) CLI, similar to [tfenv](https://github.com/tfutils/tfenv) and [pyenv](https://github.com/pyenv/pyenv).

Manage multiple versions of the vcluster CLI and switch between them seamlessly.

## Documentation

- [Docs index](docs/index.md)
- [Installation and configuration](docs/installation-and-configuration.md)
- [CLI reference](docs/cli-reference.md)
- [Caching strategy](docs/caching.md)

## Downloads (latest release)

Latest release: **v0.1.1** (tag: `v0.1.1`).

| Platform | Asset | Download |
|---|---|---|
| macOS (Intel) | `vc-env-darwin-amd64` | [Download](https://github.com/mmpyro/vc-env/releases/download/v.0.1.1/vc-env-darwin-amd64) |
| macOS (Apple Silicon) | `vc-env-darwin-arm64` | [Download](https://github.com/mmpyro/vc-env/releases/download/v.0.1.1/vc-env-darwin-arm64) |
| Linux (x86_64) | `vc-env-linux-amd64` | [Download](https://github.com/mmpyro/vc-env/releases/download/v.0.1.1/vc-env-linux-amd64) |
| Linux (ARM64) | `vc-env-linux-arm64` | [Download](https://github.com/mmpyro/vc-env/releases/download/v.0.1.1/vc-env-linux-arm64) |

## Features

- Install and manage multiple vcluster CLI versions
- Automatic version switching based on project directory (`.vcluster-version`)
- Shell-level, local (directory), and global version configuration
- Version priority: shell > local > global
- Auto-detection of OS and architecture for binary downloads
- Shim-based transparent proxying of `vcluster` commands

## Installation

### From Source

```sh
git clone https://github.com/mmpyro/vc-env.git
cd vc-env
make build
```

The binary will be at `build/vc-env`. Move it to a directory in your `PATH`:

```sh
sudo mv build/vc-env /usr/local/bin/
```

### Cross-compile for All Platforms

```sh
make build-all
```

This produces binaries for:
- `linux/amd64`
- `linux/arm64`
- `darwin/amd64`
- `darwin/arm64`

## Quick Start

### 1. Set up VCENV_ROOT

Add to your `~/.bashrc` or `~/.zshrc`:

```sh
export VCENV_ROOT="$HOME/.vcenv"
```

Reload your shell:

```sh
source ~/.bashrc  # or source ~/.zshrc
```

### 2. Initialize vc-env

Add the following to your `~/.bashrc` or `~/.zshrc` (after the `VCENV_ROOT` export):

```sh
eval "$(vc-env init)"
```

This sets up:
- A `vc-env` shell function for `vc-env shell` support
- PATH prepend for the vcluster shim

### 3. Install a vcluster version

```sh
# Install a specific version
vc-env install 0.21.1

# Install the latest stable version
vc-env install
```

### 4. Set a version

```sh
# Set global default
vc-env global 0.21.1

# Set for current directory (creates .vcluster-version)
vc-env local 0.21.1

# Set for current shell session
vc-env shell 0.21.1
```

### 5. Use vcluster

```sh
vcluster version
```

The shim automatically resolves and uses the correct version.

## Command Reference

Full reference: [docs/cli-reference.md](docs/cli-reference.md)

| Command | Description |
|---------|-------------|
| `vc-env help` | Display help and all available commands |
| `vc-env list` | List all installed versions |
| `vc-env list-remote` | List all available vcluster versions from GitHub |
| `vc-env list-remote --prerelease` | Include pre-release vcluster versions |
| `vc-env latest` | Print the latest available version of vcluster from GitHub |
| `vc-env init` | Initialize vc-env setup |
| `vc-env status` | Show current environment status |
| `vc-env install [VERSION]` | Install a specific version (or latest) |
| `vc-env uninstall VERSION` | Uninstall a specific version |
| `vc-env exec VERSION CMD` | Run a command using a specific vcluster version |
| `vc-env shell [VERSION]` | Set/show shell version (`VCENV_VERSION`) |
| `vc-env local [VERSION]` | Set/show local version (`.vcluster-version`) |
| `vc-env global [VERSION]` | Set/show global version (`$VCENV_ROOT/version`) |
| `vc-env which` | Print path to active vcluster binary |
| `vc-env version` | Print vc-env version |

## Version Priority

When `vcluster` is invoked, the version is resolved in this order:

1. **Shell** — `VCENV_VERSION` environment variable (set via `vc-env shell`)
2. **Local** — `.vcluster-version` file in the current or parent directories (set via `vc-env local`)
3. **Global** — `$VCENV_ROOT/version` file (set via `vc-env global`)

If no version is configured at any level, the command fails with an informative error.

## Shell Setup

### Bash

Add the following to your `~/.bashrc`:

```sh
# vc-env setup
export VCENV_ROOT="$HOME/.vcenv"
eval "$(vc-env init)"
source <(vc-env autocompletion)
```

### Zsh

Add the following to your `~/.zshrc`:

```sh
# vc-env setup
export VCENV_ROOT="$HOME/.vcenv"
eval "$(vc-env init)"
source <(vc-env autocompletion)
```

## Caching

`vc-env` uses a three-layer caching strategy to keep `list-remote` and `latest` fast and reliable. For more details on how it works and how to configure it, see [docs/caching.md](docs/caching.md).

## Directory Structure

```
$VCENV_ROOT/
├── versions/           # Installed vcluster versions
│   ├── 0.21.1/
│   │   └── vcluster    # vcluster binary
│   └── 0.22.0/
│       └── vcluster
├── shims/
│   └── vcluster        # Shim script (auto-generated)
└── version             # Global version file
```

## Development

### Run Tests

```sh
make test
```

### Run Docker Integration Tests

```sh
make test-docker
```

### Build

```sh
make build          # Current platform
make build-all      # All platforms
```

## License

See [LICENSE](LICENSE) for details.
