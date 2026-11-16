# vc-env

vc-env is a version manager for the [vcluster](https://www.vcluster.com/) CLI. It works like [tfenv](https://github.com/tfutils/tfenv) and [pyenv](https://github.com/pyenv/pyenv).

Use it to install multiple versions of the vcluster CLI and to switch between them.

## Documentation

Hosted site: **<https://mmpyro.github.io/vc-env/>**. MkDocs Material builds the site from `docs/`, and
[`.github/workflows/docs.yml`](.github/workflows/docs.yml) deploys it.

Sources:

- [Docs index](docs/index.md)
- [Installation and configuration](docs/installation-and-configuration.md)
- [CLI reference](docs/cli-reference.md)
- [Caching strategy](docs/caching.md)

To preview the site locally, run `make docs-serve`. On the first run, it creates `.venv-docs/`.
It opens the site at http://127.0.0.1:8000/vc-env/.

## Downloads (latest release)

Latest release: **v1.2.0** (tag: `v1.2.0`).

| Platform | Asset | Download |
|---|---|---|
| macOS (Intel) | `vc-env-darwin-amd64` | [Download](https://github.com/mmpyro/vc-env/releases/download/v1.2.0/vc-env-darwin-amd64) |
| macOS (Apple Silicon) | `vc-env-darwin-arm64` | [Download](https://github.com/mmpyro/vc-env/releases/download/v1.2.0/vc-env-darwin-arm64) |
| Linux (x86_64) | `vc-env-linux-amd64` | [Download](https://github.com/mmpyro/vc-env/releases/download/v1.2.0/vc-env-linux-amd64) |
| Linux (ARM64) | `vc-env-linux-arm64` | [Download](https://github.com/mmpyro/vc-env/releases/download/v1.2.0/vc-env-linux-arm64) |

## Features

- Install and manage multiple vcluster CLI versions
- Automatic version switch based on the project directory (`.vcluster-version`)
- Shell-level, local (directory) and global version configuration
- Version priority: shell > local > global
- Automatic detection of the OS and architecture for binary downloads
- A shim that transparently proxies `vcluster` commands

## Installation

### From source

```sh
git clone https://github.com/mmpyro/vc-env.git
cd vc-env
make build
```

The binary is then at `build/vc-env`. Move it to a directory in your `PATH`:

```sh
sudo mv build/vc-env /usr/local/bin/
```

### Cross-compile for all platforms

```sh
make build-all
```

This command builds binaries for these platforms:
- `linux/amd64`
- `linux/arm64`
- `darwin/amd64`
- `darwin/arm64`

### CI tip

If you use `vc-env` in CI (for example GitHub Actions), export `GITHUB_TOKEN`
or the vc-env-specific `VCENV_GITHUB_TOKEN`. Do this before you run `vc-env install`,
`vc-env latest` or `vc-env list-remote`. Without a token, GitHub applies an
anonymous rate limit of 60 requests per hour per IP. Shared CI runners hit this
limit easily. Authenticated requests get a quota of 5000/h. For the full list of
supported environment variables, see
[docs/cli-reference.md](docs/cli-reference.md#environment-variables).

## Quick start

### 1. Set VCENV_ROOT

Add this line to your `~/.bashrc` or `~/.zshrc`:

```sh
export VCENV_ROOT="$HOME/.vcenv"
```

Reload your shell:

```sh
source ~/.bashrc  # or source ~/.zshrc
```

### 2. Initialize vc-env

Add this line to your `~/.bashrc` or `~/.zshrc`, after the `VCENV_ROOT` export:

```sh
eval "$(vc-env init)"
```

This line does two things:
- It defines a `vc-env` shell function, which `vc-env shell` needs.
- It prepends the vcluster shim to `PATH`.

### 3. Install a vcluster version

```sh
# Install a specific version
vc-env install 0.21.1

# Install the latest stable version
vc-env install

# Install using an alias (resolves to a concrete version)
vc-env install latest        # newest non-prerelease
vc-env install 0.21          # highest 0.21.x
vc-env install ~0.21.1       # highest 0.21.y with y >= 1
```

Tip: if you set `VCENV_AUTO_INSTALL=1` in your shell profile, the `vcluster`
shim installs missing versions automatically on first use. This is useful when
teammates clone a repo that contains a `.vcluster-version` file.

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

The shim resolves the correct version and uses it automatically.

## Command reference

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
| `vc-env doctor [--fix] [--deep]` | Diagnose the environment (PATH, shim, binaries, GitHub, cache). `--deep` verifies binaries against release checksums |
| `vc-env prune [--keep-last N] [--older-than DUR] [--yes]` | Remove installed versions that no `.vcluster-version`, global or shell setting references (dry run unless `--yes`) |
| `vc-env install [VERSION\|ALIAS]` | Install a specific version or alias (or latest). Supports `--from-file`, `--sha256` |
| `vc-env uninstall VERSION` | Uninstall a specific version |
| `vc-env exec [--auto\|--no-auto] VERSION\|ALIAS [--] ARGS...` | Run vcluster at a specific version. Passes through the exit code of the child process |
| `vc-env shell [VERSION\|ALIAS]` | Set/show shell version (`VCENV_VERSION`) |
| `vc-env local [VERSION\|ALIAS]` | Set/show local version (`.vcluster-version`) |
| `vc-env global [VERSION\|ALIAS]` | Set/show global version (`$VCENV_ROOT/version`) |
| `vc-env resolve [VERSION\|ALIAS] [--install] [-s]` | Print the concrete installed version (no argument: the active version) |
| `vc-env which` | Print path to active vcluster binary |
| `vc-env version` | Print vc-env version |
| `vc-env upgrade` | Fetch the latest vc-env release from GitHub, verify its checksum, and replace the current binary in place |
| `vc-env completion [SHELL] [--shell SHELL]` | Generate a shell completion script (`bash`, `zsh`, `fish`, `powershell`/`pwsh`). Without a shell argument, it detects the shell from `$SHELL` |

Every command accepts `-h`/`--help`.

## Environment variables

| Variable | Purpose |
|---|---|
| `VCENV_ROOT` | Required. The directory for versions, shims and config |
| `VCENV_VERSION` | Shell-level version (set by `vc-env shell`). Highest priority |
| `VCENV_AUTO_INSTALL` | `1`: install missing versions on demand (shim, `resolve`, `exec`) |
| `VCENV_GITHUB_TOKEN` | GitHub token. Falls back to `GITHUB_TOKEN` |
| `VCENV_GITHUB_API_URL` | GitHub API base URL (default `https://api.github.com`), e.g. for GitHub Enterprise |
| `VCENV_DOWNLOAD_MIRROR` | Download base URL (default `https://github.com`), e.g. for air-gapped mirrors |
| `VCENV_CACHE_TTL` | Release cache TTL as a Go duration (default `1h`) |
| `VCENV_PRUNE_SCAN_ROOTS` | Roots that `prune` scans for `.vcluster-version` files (default `$HOME`, an empty value disables scanning) |

## Version priority

When you run `vcluster`, the shim resolves the version in this order:

1. Shell: the `VCENV_VERSION` environment variable (set with `vc-env shell`)
2. Local: a `.vcluster-version` file in the current or parent directories (set with `vc-env local`)
3. Global: the `$VCENV_ROOT/version` file (set with `vc-env global`)

If no level has a version, the command fails with an informative error.

## Shell setup

`vc-env` has native completion scripts for Bash, Zsh, Fish, and PowerShell. Use the block for your shell.

### Bash

Add the following to your `~/.bashrc`:

```sh
# vc-env setup
export VCENV_ROOT="$HOME/.vcenv"
eval "$(vc-env init)"
source <(vc-env completion bash)
```

### Zsh

Add the following to your `~/.zshrc`:

```sh
# vc-env setup
export VCENV_ROOT="$HOME/.vcenv"
eval "$(vc-env init)"

# Zsh completion requires the completion system to be initialised once.
autoload -Uz compinit && compinit
source <(vc-env completion zsh)
```

### Fish

Add the following to your `~/.config/fish/config.fish`:

```fish
# vc-env setup
set -gx VCENV_ROOT "$HOME/.vcenv"
vc-env init | source
vc-env completion fish | source
```

You can also install the completion permanently. Then Fish loads it automatically:

```fish
vc-env completion fish > ~/.config/fish/completions/vc-env.fish
```

### PowerShell

Add the following to your `$PROFILE`:

```powershell
# vc-env setup
$env:VCENV_ROOT = "$HOME\.vcenv"
vc-env init | Out-String | Invoke-Expression
vc-env completion powershell | Out-String | Invoke-Expression
```

> `vc-env autocompletion` is a deprecated alias for `vc-env completion` and takes the same arguments. `vc-env` still accepts it, so existing shell rc files keep working. It prints a deprecation warning to stderr.

## Caching

`vc-env` uses a three-layer caching strategy to keep `list-remote` and `latest` fast and reliable. To learn how it works and how to configure it, see [docs/caching.md](docs/caching.md).

## Directory structure

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

### Run tests

```sh
make test
```

### Run Docker integration tests

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
