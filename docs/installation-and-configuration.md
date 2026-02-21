# Installation and configuration

This guide covers supported platforms, installation methods, required setup steps, how `vc-env` discovers configuration, and common troubleshooting.

## Supported platforms

Prebuilt binaries are currently produced for:

- `darwin/amd64`
- `darwin/arm64`
- `linux/amd64`
- `linux/arm64`

If your platform is not listed, install from source.

## Prerequisites

- A POSIX-like shell (e.g. `bash` or `zsh`).
- Permission to create and write files in your chosen `VCENV_ROOT` directory.
- Network access to GitHub:
  - `vc-env install`, `vc-env latest`, and `vc-env list-remote` fetch information from GitHub.

No existing `vcluster` installation is required; `vc-env` manages the `vcluster` binaries it installs.

## Install methods

### Option A: Install a prebuilt binary (recommended)

1. Download the binary for your platform from the project’s GitHub releases.
2. Make it executable and move it into a directory on your `PATH`.

Example (Linux x86_64):

```sh
curl -L -o vc-env https://github.com/mmpyro/vc-env/releases/download/v.0.1.0/vc-env-linux-amd64
chmod +x vc-env
sudo mv vc-env /usr/local/bin/vc-env
```

### Option B: Build from source

```sh
git clone https://github.com/mmpyro/vc-env.git
cd vc-env
make build
```

The binary will be available at `build/vc-env`.

### Option C: Cross-compile for all supported platforms

```sh
make build-all
```

## Initial setup

### 1) Choose and export `VCENV_ROOT`

`VCENV_ROOT` is required. It defines where `vc-env` stores installed versions, the shim, and the global version file.

Add this to your shell profile (e.g. `~/.bashrc` or `~/.zshrc`):

```sh
export VCENV_ROOT="$HOME/.vcenv"
```

Required permissions:

- `vc-env init` creates directories under `$VCENV_ROOT`.
- `vc-env install` writes `vcluster` binaries under `$VCENV_ROOT/versions/<version>/vcluster`.
- `vc-env global` writes `$VCENV_ROOT/version`.

### 2) Initialize shell integration

Add this after the `VCENV_ROOT` export:

```sh
eval "$(vc-env init)"
```

What this does:

- Prepends `$VCENV_ROOT/shims` to your `PATH` so `vcluster` resolves to the shim.
- Defines a `vc-env` shell function that enables `vc-env shell` to affect the current shell environment.

### 3) Install a `vcluster` version

```sh
vc-env install 0.21.1
```

Or install the latest stable version:

```sh
vc-env install
```

### 4) Configure which version to use

Pick one of:

- Global default (applies everywhere unless overridden):

  ```sh
  vc-env global 0.21.1
  ```

- Per-directory (creates `.vcluster-version` in the current directory):

  ```sh
  vc-env local 0.21.1
  ```

- Per-shell session (sets `VCENV_VERSION`; requires `eval "$(vc-env init)"`):

  ```sh
  vc-env shell 0.21.1
  ```

## How configuration is discovered/loaded

When the `vcluster` shim runs, it selects a version using this priority order:

1. `VCENV_VERSION` (shell version)
2. `.vcluster-version` in the current directory or any parent directory (local version)
3. `$VCENV_ROOT/version` (global version)

Notes:

- The `.vcluster-version` lookup walks upward until the filesystem root.
- All version values are treated as strings and trimmed for whitespace.

## Common troubleshooting

### `VCENV_ROOT not set`

Symptoms:

- `vc-env init` prints instructions and exits with an error.

Fix:

```sh
export VCENV_ROOT="$HOME/.vcenv"
```

Then ensure you also have:

```sh
eval "$(vc-env init)"
```

### `vc-env: no vcluster version configured`

This is emitted by the `vcluster` shim when none of the version sources are configured.

Fix (choose one):

```sh
vc-env global 0.21.1
vc-env local 0.21.1
vc-env shell 0.21.1
```

### `version <X> not installed`

`vc-env` validates that a version is installed before setting it via `global`, `local`, or `shell`, and the shim also verifies the installed binary is present.

Fix:

```sh
vc-env install <X>
```

### GitHub API rate limit exceeded

Some commands query GitHub releases. If GitHub returns `403`, `vc-env` reports a rate limit error.

Fixes:

- Wait and retry later.
- If running in CI or heavily automated use, consider reducing frequency of `list-remote` / `latest` calls.

