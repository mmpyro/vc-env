# CLI reference

This reference covers every `vc-env` command implemented by the project.

Conventions:

- Commands return exit code `0` on success.
- On errors, commands typically print an error message to stderr and exit with code `1`.
- Some commands print help and exit `0`.

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

Syntax:

```text
vc-env install [version]
```

Options/flags: none.

Environment variables:

- `VCENV_ROOT` (required)

Exit codes:

- `0` on success.
- `1` if not initialized, platform detection fails, download fails, or filesystem writes fail.

Example:

```sh
vc-env install 0.21.1
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

Important: To *set* the version for your current shell session, you must have shell integration enabled via `eval "$(vc-env init)"`. Otherwise, you will only see the printed `export ...` line but your current shell will not be updated.

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

