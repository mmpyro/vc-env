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

### `VCENV_AUTO_INSTALL`

Optional. When set to `1`, the `vcluster` shim automatically installs missing
versions on demand instead of failing with an error. This is especially useful
for teammates cloning a repository with a `.vcluster-version` file.

Internally the shim calls `vc-env ensure <version>`, which also resolves
aliases (`latest`, `0.21`, `~0.21.1`, …) to a concrete version before
installing.

### `VCENV_GITHUB_TOKEN`

Optional. GitHub personal access token used to authenticate requests to the GitHub API (and asset downloads from private mirrors / GHES). When set it is sent as `Authorization: Bearer <token>` on every GitHub request issued by `vc-env`, raising the anonymous rate limit (60/h) to the authenticated limit (5000/h).

Takes precedence over `GITHUB_TOKEN` so that a shell-wide `GITHUB_TOKEN` cannot silently change `vc-env` behaviour.

### `GITHUB_TOKEN`

Optional. Fallback GitHub token used only when `VCENV_GITHUB_TOKEN` is unset. Convenient in CI environments where `GITHUB_TOKEN` is already injected by the runner (for example GitHub Actions).

### `VCENV_GITHUB_API_URL`

Optional. Overrides the GitHub API base URL (default `https://api.github.com`). Set this to point `vc-env` at a GitHub Enterprise Server instance, for example `https://ghe.example.com/api/v3`.

### `VCENV_DOWNLOAD_MIRROR`

Optional. Overrides the asset download base URL (default `https://github.com`). Set this to pull `vcluster` and `vc-env` release assets from an internal mirror (for air-gapped environments). `vc-env install`, `vc-env latest`, and `vc-env upgrade` all honour this value.

## Version aliases

Several commands (`install`, `global`, `local`, `shell`, `resolve`, `ensure`)
accept version aliases in addition to concrete versions. Aliases are stored
verbatim when written to `.vcluster-version`, `$VCENV_ROOT/version`, or
`VCENV_VERSION`, and are re-resolved on every shim call.

| Alias | Meaning |
|---|---|
| `latest`, `latest-stable` | Newest non-prerelease release |
| `latest-prerelease` | Newest release including prereleases |
| `MAJOR.MINOR` (e.g. `0.21`) | Highest patch of that minor |
| `~MAJOR.MINOR.PATCH` (e.g. `~0.21.1`) | Highest version `>=0.21.1 <0.22.0` |

Resolution tries installed versions first (fast, offline), then falls back to
the remote release list.

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

If `<version>` is omitted, `vc-env` installs the latest stable version. An
alias (e.g. `latest`, `0.21`, `~0.21.1`) is resolved first and the concrete
version is then installed. See [Version aliases](#version-aliases).

The command displays a progress bar during the download and automatically verifies the integrity of the downloaded file using SHA256 checksums from the GitHub release.

Syntax:

```text
vc-env install [version-or-alias] [flags]
```

Options/flags:

- `-s`, `--silent`: do not display the progress bar or checksum verification information
- `--from-file <path>`: install the given local file as the vcluster binary
  instead of downloading. Requires a concrete `<version>` argument and skips
  all network calls (air-gapped setups). Combine with `--sha256` to verify
  integrity.
- `--sha256 <hex>`: expected hex-encoded SHA-256 of the binary. Overrides the
  checksums.txt from the GitHub release and applies to both remote and
  `--from-file` installs. A mismatch aborts installation.
- `-h`, `--help`: show command help and exit

Environment variables:

- `VCENV_ROOT` (required)

Exit codes:

- `0` on success.
- `1` if not initialized, platform detection fails, download fails, checksum mismatch, or filesystem writes fail.

Example:

```sh
vc-env install 0.21.1
vc-env install --silent
vc-env install
vc-env install 0.21                                  # highest 0.21.x
vc-env install ~0.21.1                               # >=0.21.1 <0.22.0
vc-env install latest
# Air-gapped install from a pre-downloaded binary:
vc-env install 0.21.1 --from-file ./vcluster --sha256 <hex>
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

The `<version>` argument may be a concrete version (which must be installed)
or an alias (`latest`, `0.21`, `~0.21.1`, …). Aliases are stored verbatim and
re-resolved by the shim on every `vcluster` call. See
[Version aliases](#version-aliases).

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

Setting writes a `.vcluster-version` file into the current directory. The
value may be a concrete version or an alias (`latest`, `0.21`, `~0.21.1`, …).
Aliases are stored verbatim and re-resolved by the shim on every `vcluster`
call. See [Version aliases](#version-aliases).

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

Setting writes `$VCENV_ROOT/version`. The value may be a concrete version or
an alias (`latest`, `0.21`, `~0.21.1`, …). Aliases are stored verbatim and
re-resolved by the shim on every `vcluster` call. See
[Version aliases](#version-aliases).

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

### `completion`

Purpose: Generate a native shell completion script for `vc-env`.

The script completes subcommands and dynamically suggests versions:

- Installed versions for `uninstall`, `shell`, `local`, `global`, and `exec`.
- Cached remote versions for `install` (never performs a network request; falls back to the compiled-in baseline when the cache is empty).
- Shell names (`bash`, `zsh`, `fish`, `powershell`) for `completion`.

Syntax:

```text
vc-env completion <bash|zsh|fish|powershell>
```

Options/flags:
- `-h`, `--help`: show command help and exit

Environment variables:
- `VCENV_ROOT` (read at completion time by the generated scripts, via `vc-env __complete-versions`).

Exit codes:
- `0` on success.
- `1` if the shell argument is missing or unsupported.

Install (Bash):

```sh
# Current session
source <(vc-env completion bash)

# Permanently
echo 'source <(vc-env completion bash)' >> ~/.bashrc
```

Install (Zsh):

```sh
# Ensure the compsys framework is loaded once in ~/.zshrc:
autoload -Uz compinit && compinit

# Current session
source <(vc-env completion zsh)

# Permanently
echo 'source <(vc-env completion zsh)' >> ~/.zshrc
```

Install (Fish):

```fish
# Current session
vc-env completion fish | source

# Permanently
vc-env completion fish > ~/.config/fish/completions/vc-env.fish
```

Install (PowerShell):

```powershell
# Current session
vc-env completion powershell | Out-String | Invoke-Expression

# Permanently (append to your $PROFILE)
Add-Content -Path $PROFILE -Value 'vc-env completion powershell | Out-String | Invoke-Expression'
```

Internal helper:

The generated scripts call `vc-env __complete-versions {installed|remote}` to list candidate versions. This command is not part of the public interface and may change at any time, but it is safe to invoke manually for debugging.

Back-compat:

`vc-env autocompletion` remains available as a deprecated alias for `vc-env completion bash`; it prints a deprecation notice to stderr. Existing `~/.bashrc` snippets continue to work unchanged.

---

### `resolve`

Purpose: Resolve a version alias to a concrete version and print it to stdout.
Does not install anything.

Primarily intended for scripts and for the `vcluster` shim as a fallback when
`VCENV_AUTO_INSTALL` is not set.

Syntax:

```text
vc-env resolve <version-or-alias>
```

See [Version aliases](#version-aliases) for the supported syntax.

Environment variables:

- `VCENV_ROOT` (optional; speeds up resolution by matching installed versions first)

Exit codes:

- `0` on success.
- `1` on resolution failure (no matching version in installed or remote lists).

Example:

```sh
vc-env resolve latest        # e.g. 0.31.0
vc-env resolve 0.21          # e.g. 0.21.3
vc-env resolve ~0.21.1       # e.g. 0.21.3
vc-env resolve 0.21.1        # 0.21.1 (echoed verbatim)
```

---

### `ensure`

Purpose: Resolve a version alias and install the concrete version if it is
not already installed. Prints the concrete version to stdout.

This is the subcommand the `vcluster` shim invokes when
`VCENV_AUTO_INSTALL=1` is set and the requested version isn't yet present.
Install runs silently (no progress bar) so the shim can capture the final
line as the resolved version.

Syntax:

```text
vc-env ensure <version-or-alias>
```

Environment variables:

- `VCENV_ROOT` (required)

Exit codes:

- `0` on success.
- `1` on resolution failure, download failure, or checksum mismatch.

Example:

```sh
export VCENV_AUTO_INSTALL=1
vc-env ensure latest
vc-env ensure 0.21
```
