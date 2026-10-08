# CLI reference

This reference covers every `vc-env` command implemented by the project.

Conventions:

- Commands return exit code `0` on success.
- On errors, commands typically print an error message to stderr and exit with code `1`.
- Some commands print help and exit `0`.

## Command summary

| Command | What it does |
|---|---|
| [`help`](#help) | Print usage and the list of commands |
| [`version`](#version) | Print the `vc-env` version |
| [`init`](#init) | Create the `VCENV_ROOT` layout and print shell integration code |
| [`list`](#list) | List installed `vcluster` versions |
| [`list-remote`](#list-remote) | List `vcluster` versions available on GitHub |
| [`latest`](#latest) | Print the latest `vcluster` release on GitHub |
| [`install`](#install) | Download, checksum-verify and install a `vcluster` version |
| [`uninstall`](#uninstall) | Remove an installed `vcluster` version |
| [`shell`](#shell) | Set or show the shell-level version (`VCENV_VERSION`) |
| [`local`](#local) | Set or show the directory-level version (`.vcluster-version`) |
| [`global`](#global) | Set or show the global default (`$VCENV_ROOT/version`) |
| [`which`](#which) | Print the full path to the active `vcluster` binary |
| [`upgrade`](#upgrade) | Replace the current `vc-env` binary with the latest release |
| [`exec`](#exec) | Run `vcluster` at a specific version for a single command |
| [`status`](#status) | Show `VCENV_ROOT`, resolved version, source and all installed versions |
| [`autocompletion`](#autocompletion) | Print a bash autocompletion script |

## Global usage

```text
vc-env <command> [arguments]
```

Top-level help is available via `vc-env help`, `vc-env --help`, or `vc-env -h`.

## Environment variables

### `VCENV_ROOT`

Required. Path where `vc-env` stores installed versions and shims.

Used by most commands and required for initialization.

### `VCENV_VERSION`

Optional. When set, it forces a particular `vcluster` version to be used (highest priority).

Typically set via `vc-env shell` after enabling shell integration with `eval "$(vc-env init)"`.

### `VCENV_GITHUB_TOKEN`

Optional. GitHub personal access token used to authenticate requests to the GitHub API (and asset downloads from private mirrors / GHES). When set it is sent as `Authorization: Bearer <token>` on every GitHub request issued by `vc-env`, raising the anonymous rate limit (60/h) to the authenticated limit (5000/h).

Takes precedence over `GITHUB_TOKEN` so that a shell-wide `GITHUB_TOKEN` cannot silently change `vc-env` behaviour.

### `GITHUB_TOKEN`

Optional. Fallback GitHub token used only when `VCENV_GITHUB_TOKEN` is unset. Convenient in CI environments where `GITHUB_TOKEN` is already injected by the runner (for example GitHub Actions).

### `VCENV_GITHUB_API_URL`

Optional. Overrides the GitHub API base URL (default `https://api.github.com`). Set this to point `vc-env` at a GitHub Enterprise Server instance, for example `https://ghe.example.com/api/v3`.

### `VCENV_DOWNLOAD_MIRROR`

Optional. Overrides the asset download base URL (default `https://github.com`). Set this to pull `vcluster` and `vc-env` release assets from an internal mirror (for air-gapped environments). `vc-env install`, `vc-env latest`, and `vc-env upgrade` all honour this value.

## Commands

### `help`

Purpose: Print usage and the list of available commands.

Syntax:

```text
vc-env help
vc-env --help
vc-env -h
```

Options/flags: none.

Environment variables: none.

Exit codes:

- `0` always.

Example:

```sh
vc-env help
```

---

### `version`

Purpose: Print the `vc-env` version.

Syntax:

```text
vc-env version
vc-env --version
vc-env -v
```

Options/flags: none.

Environment variables: none.

Exit codes:

- `0` on success.

Example:

```sh
vc-env version
```

---

### `init`

Purpose:

- Create required directories under `VCENV_ROOT`.
- Generate the `vcluster` shim under `$VCENV_ROOT/shims/vcluster`.
- Print shell initialization code to stdout (intended to be evaluated by your shell).

Syntax:

```text
vc-env init
```

Options/flags: none.

Environment variables:

- `VCENV_ROOT` (required)

Exit codes:

- `0` on success.
- `1` if `VCENV_ROOT` is not set or filesystem operations fail.

!!! note "Eval, not source"
    `vc-env init` emits shell code to stdout; `eval "$(vc-env init)"`
    is what makes the shim and the `vc-env` shell function available in
    the current shell.

Example:

```sh
export VCENV_ROOT="$HOME/.vcenv"
eval "$(vc-env init)"
```

---

### `list`

Purpose: List installed `vcluster` versions (newest to oldest).

Syntax:

```text
vc-env list
```

Options/flags: none.

Environment variables:

- `VCENV_ROOT` (required)

Exit codes:

- `0` on success.
- `1` if `vc-env` is not initialized.

Example:

```sh
vc-env list
```

---

### `list-remote`

Purpose: List all available `vcluster` versions from GitHub releases (newest to oldest).

This command does not require `VCENV_ROOT` or initialization. If `VCENV_ROOT` is set, results are persistently cached on disk (see [caching strategy](caching.md)).

Syntax:

```text
vc-env list-remote [flags]
```

Options/flags:

- `--prerelease`: include pre-release versions (alpha, beta, rc)
- `-h`, `--help`: show command help and exit

Environment variables: none.

Exit codes:

- `0` on success, or when printing `--help`.
- `1` on GitHub/network errors (including rate limiting).

Example:

```sh
vc-env list-remote --prerelease
```

---

### `latest`

Purpose: Print the latest available `vcluster` version from GitHub releases.

This command does not require `VCENV_ROOT` or initialization. If `VCENV_ROOT` is set, results are persistently cached on disk (see [caching strategy](caching.md)).

Syntax:

```text
vc-env latest [flags]
```

Options/flags:

- `--prerelease`: include pre-release versions when selecting the latest
- `-h`, `--help`: show command help and exit

Environment variables: none.

Exit codes:

- `0` on success, or when printing `--help`.
- `1` if no versions are found, or on GitHub/network errors.

Example:

```sh
vc-env latest
```

---

### `install`

Purpose: Download and install a `vcluster` version into `$VCENV_ROOT/versions/<version>/vcluster`.

If `<version>` is omitted, `vc-env` installs the latest stable version.

The command displays a progress bar during the download and automatically verifies the integrity of the downloaded file using SHA256 checksums from the GitHub release.

Syntax:

```text
vc-env install [version] [flags]
```

Options/flags:

- `-s`, `--silent`: do not display the progress bar or checksum verification information
- `-h`, `--help`: show command help and exit

Environment variables:

- `VCENV_ROOT` (required)

Exit codes:

- `0` on success.
- `1` if not initialized, platform detection fails, download fails, checksum mismatch, or filesystem writes fail.

!!! warning "Rate limits and air-gapped environments"
    Without a token, GitHub applies a 60 requests/hour anonymous limit
    per IP, which is easy to exhaust on shared CI runners. Export
    `VCENV_GITHUB_TOKEN` (or `GITHUB_TOKEN` in CI) to raise it to
    5000/h. To pull assets from a mirror instead of `github.com`, set
    `VCENV_DOWNLOAD_MIRROR`.

Example:

```sh
vc-env install 0.21.1
vc-env install --silent
vc-env install
```

---

### `uninstall`

Purpose: Remove an installed `vcluster` version directory.

Syntax:

```text
vc-env uninstall <version>
```

Options/flags: none.

Environment variables:

- `VCENV_ROOT` (required)

Exit codes:

- `0` on success.
- `1` if `<version>` is missing, not initialized, the version is not installed, or filesystem removal fails.

Example:

```sh
vc-env uninstall 0.21.1
```

---

### `shell`

Purpose: Set or show the shell-level `vcluster` version.

!!! note "Requires shell integration"
    To *set* the version for your current shell session, you must have
    shell integration enabled via `eval "$(vc-env init)"`. Otherwise,
    you will only see the printed `export …` line but your current
    shell will not be updated.

Syntax:

```text
vc-env shell            # show
vc-env shell <version>  # set
```

Options/flags: none.

Environment variables:

- `VCENV_ROOT` (required)
- `VCENV_VERSION` (read on show; written when you eval shell integration)

Exit codes:

- `0` on success.
- `1` if not initialized, no shell version is configured (show), or the requested version is not installed (set).

Example:

```sh
eval "$(vc-env init)"
vc-env shell 0.21.1
vcluster version
```

---

### `local`

Purpose: Set or show the local (directory-level) `vcluster` version.

Setting writes a `.vcluster-version` file into the current directory.

Syntax:

```text
vc-env local            # show
vc-env local <version>  # set
```

Options/flags: none.

Environment variables:

- `VCENV_ROOT` (required)

Exit codes:

- `0` on success.
- `1` if not initialized, no local version is configured for this directory (show), the requested version is not installed (set), or writing `.vcluster-version` fails.

Example:

```sh
vc-env install 0.21.1
vc-env local 0.21.1
vcluster version
```

---

### `global`

Purpose: Set or show the global default `vcluster` version.

Setting writes `$VCENV_ROOT/version`.

Syntax:

```text
vc-env global            # show
vc-env global <version>  # set
```

Options/flags: none.

Environment variables:

- `VCENV_ROOT` (required)

Exit codes:

- `0` on success.
- `1` if not initialized, no global version is configured (show), the requested version is not installed (set), or writing fails.

Example:

```sh
vc-env install 0.21.1
vc-env global 0.21.1
vcluster version
```

---

### `which`

Purpose: Print the full path to the active `vcluster` binary that would be used based on version resolution.

Syntax:

```text
vc-env which
```

Options/flags: none.

Environment variables:

- `VCENV_ROOT` (required)
- `VCENV_VERSION` (optional; highest priority if set)

Exit codes:

- `0` on success.
- `1` if not initialized or no version is configured.

Example:

```sh
vc-env which
```

---

### `upgrade`

Purpose: Download the latest stable release of `vc-env` from GitHub and replace the current binary in-place.

Syntax:

```text
vc-env upgrade
```

Options/flags: none.

Environment variables: none (the binary path is auto-detected via `os.Executable()`).

Exit codes:

- `0` on success, or when already up to date.
- `1` on network/download errors, permission issues, or filesystem errors.

Notes:

- The command detects the current OS and CPU architecture automatically.
- The new binary is written atomically (temp file + rename) to avoid corruption.
- If the current version is a development build (`dev`), the upgrade always proceeds.

Example:

```sh
vc-env upgrade
```

---

### `exec`

Purpose: Run a specific version of `vcluster` for a single command without changing the active version (shell, local, or global).

!!! tip "One-off calls"
    `exec` is the right tool when you need a different `vcluster`
    version for a single command and don't want to touch any of the
    three persistent sources (`VCENV_VERSION`, `.vcluster-version`,
    `$VCENV_ROOT/version`).

Syntax:

```text
vc-env exec <version> <command> [args...]
```

Options/flags: none.

Environment variables:

- `VCENV_ROOT` (required)
- `VCENV_VERSION` (set for the subprocess to match the requested version)

Exit codes:

- Exit code of the executed command.
- `1` if the version is not installed or initialization fails.

Example:

```sh
vc-env exec 0.21.1 version
```

---

### `status`

Purpose: Display a comprehensive overview of the current `vc-env` environment.

Output includes:

- `VCENV_ROOT` path.
- Currently active version and the source it was resolved from.
- Full path to the active `vcluster` binary.
- List of all installed versions (active one marked with `*`).

Syntax:

```text
vc-env status
```

Options/flags: none.

Environment variables:

- `VCENV_ROOT` (required)
- `VCENV_VERSION` (read if set)

Exit codes:

- `0` on success.
- `1` on error.

Example:

```sh
vc-env status
```

---

### `autocompletion`

Purpose: Generate bash autocompletion script for `vc-env`.

The script provides completion for subcommands and suggests installed versions for commands that accept a version argument (`install`, `uninstall`, `shell`, `local`, `global`, `exec`).

Syntax:

```text
vc-env autocompletion
```

Options/flags:

- `-h`, `--help`: show command help and exit

Environment variables: none.

Exit codes:

- `0` on success.

Examples:

=== "bash"

    ```sh
    # Current session only
    source <(vc-env autocompletion)

    # Permanent
    echo 'source <(vc-env autocompletion)' >> ~/.bashrc
    ```

=== "zsh"

    ```sh
    # Current session only (zsh understands the bash completion spec via
    # bashcompinit).
    autoload -Uz +X compinit && compinit
    autoload -Uz +X bashcompinit && bashcompinit
    source <(vc-env autocompletion)

    # Permanent — add the three lines above to ~/.zshrc
    ```
