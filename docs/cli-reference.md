# CLI reference

This reference covers every `vc-env` command implemented by the project.

Conventions:

- Commands return exit code `0` on success.
- On errors, commands typically print an error message to stderr and exit with code `1`.
- Every command accepts `-h`/`--help`, prints its help and exits `0` without
  doing anything else (`upgrade --help` never upgrades).
- `exec` exits with the child process's exit code.

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
| [`resolve`](#resolve) | Print the concrete installed version for a spec (or the active version) |
| [`status`](#status) | Show `VCENV_ROOT`, resolved version, source and all installed versions |
| [`doctor`](#doctor) | Diagnose the environment and optionally repair it |
| [`prune`](#prune) | Remove installed versions nothing references |
| [`completion`](#completion) | Generate a native completion script for `bash`, `zsh`, `fish` or `powershell` |

`ensure` is an [advanced / internal](#advanced--internal) command.

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

It also makes `vc-env resolve` and `vc-env exec` install missing versions
(as if `--install` / `--auto` were passed; `exec --no-auto` opts out).
Internally the shim calls `vc-env resolve --install --silent`, which resolves
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

Optional. Overrides the asset download base URL (default `https://github.com`). Set this to pull `vcluster` and `vc-env` release assets from an internal mirror (for air-gapped environments). Every download honours it: `install`, auto-install (shim, `resolve --install`, `exec --auto`), `doctor --deep`, and `upgrade`.

### `VCENV_CACHE_TTL`

Optional. How long the on-disk release cache (`$VCENV_ROOT/cache/releases.json`)
is considered fresh, as a Go duration such as `30m` or `6h`. Default `1h`.
Invalid values fall back to the default. See [caching strategy](caching.md).

### `VCENV_PRUNE_SCAN_ROOTS`

Optional. Directories that [`prune`](#prune) scans for `.vcluster-version`
files, separated by the OS path-list separator (`:` on Linux and macOS).
Unset: scan `$HOME`. Set to an empty string: disable scanning (useful on CI
runners and servers).

## Version aliases

Several commands (`install`, `global`, `local`, `shell`, `resolve`, `exec`)
accept version aliases in addition to concrete versions. Aliases are stored
verbatim when written to `.vcluster-version`, `$VCENV_ROOT/version`, or
`VCENV_VERSION`, and are re-resolved on every shim call.

| Alias | Meaning |
|---|---|
| `latest`, `latest-stable` | Newest non-prerelease release |
| `latest-prerelease` | Newest release including prereleases |
| `MAJOR.MINOR` (e.g. `0.21`) | Highest patch of that minor |
| `~MAJOR.MINOR.PATCH` (e.g. `~0.21.1`) | Highest version `>=0.21.1 <0.22.0` |

`resolve`, `exec` and the shim match installed versions first (fast, offline;
`latest` means the newest *installed* stable version). They consult the remote
release list only when nothing installed matches and installing is allowed
(`--install`, `--auto` or `VCENV_AUTO_INSTALL=1`). `install` always resolves
aliases against the remote release list.

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

Options/flags:

- `-h`, `--help`: show command help and exit

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

Options/flags:

- `-h`, `--help`: show command help and exit

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

Options/flags:

- `-h`, `--help`: show command help and exit

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

Options/flags:

- `-h`, `--help`: show command help and exit

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

Options/flags:

- `-h`, `--help`: show command help and exit

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

Options/flags:

- `-h`, `--help`: show command help and exit

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

Options/flags:

- `-h`, `--help`: show command help and exit

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

Purpose: Fetch the latest `vc-env` release from GitHub (`mmpyro/vc-env`), verify
its SHA-256 against the release's `checksums.txt`, and replace the current
binary in place.

Syntax:

```text
vc-env upgrade
```

Options/flags:

- `-h`, `--help`: show command help and exit (never upgrades)

Environment variables:

- `VCENV_GITHUB_TOKEN` / `GITHUB_TOKEN`: GitHub token for the release lookup
- `VCENV_GITHUB_API_URL`: GitHub API base URL
- `VCENV_DOWNLOAD_MIRROR`: download base URL for the binary and checksums

The binary path is auto-detected via `os.Executable()`.

Exit codes:

- `0` on success, when already up to date, or when printing `--help`.
- `1` on network/download errors, checksum mismatch, permission issues, or filesystem errors.

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
vc-env exec [--auto|--no-auto] <version-or-alias> [--] <args...>
```

The version is resolved like [`resolve`](#resolve): installed versions first.
Everything after the version is passed to `vcluster`; a single leading `--` is
stripped, so `vc-env exec 0.21 -- --help` runs `vcluster --help`.

Options/flags (only before the version):

- `--auto`: install the version if it is missing
- `--no-auto`: never install, even when `VCENV_AUTO_INSTALL=1`
- `-h`, `--help`: show command help and exit

Environment variables:

- `VCENV_ROOT` (required)
- `VCENV_AUTO_INSTALL` (`1` implies `--auto`)
- `VCENV_VERSION` (set to the concrete version for the subprocess)

Exit codes:

- The exit code of `vcluster`, passed through unchanged.
- `1` if not initialized, or the version is not installed and cannot be auto-installed.

Example:

```sh
vc-env exec 0.21.1 version
vc-env exec --auto 0.21 -- version --output json
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

Options/flags:

- `-h`, `--help`: show command help and exit

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

### `doctor`

Purpose: Diagnose the `vc-env` environment and print one `[OK]`, `[WARN]`,
`[FAIL]` or `[INFO]` line per check, with a hint on how to fix problems.

Checks, in order:

1. `VCENV_ROOT` is set.
2. The root exists and is writable.
3. `$VCENV_ROOT/versions` exists.
4. The shim exists and matches what `vc-env init` would generate.
5. `$VCENV_ROOT/shims` is on `PATH`, ahead of any other `vcluster`.
6. A shell rc file (`~/.bashrc`, `~/.zshrc`, `~/.config/fish/config.fish`)
   contains `vc-env init` (informational only).
7. The active version is configured; aliases are resolved offline against
   installed versions.
8. The active binary exists and is executable.
9. Every installed binary is executable, and no version directory lacks a binary.
10. The GitHub API (`VCENV_GITHUB_API_URL`) is reachable, with a warning when
    the rate-limit headroom is low.
11. The release cache is valid JSON.
12. With `--deep`: each installed binary's SHA-256 matches its release's `checksums.txt`.

Syntax:

```text
vc-env doctor [--fix] [--deep]
```

Options/flags:

- `--fix`: regenerate the shim and make installed binaries executable, then
  re-run the checks. It never edits `PATH` or shell rc files.
- `--deep`: download each installed version's `checksums.txt` and compare the
  entry for this platform with the on-disk binary (one small download per version)
- `-h`, `--help`: show command help and exit

Environment variables:

- `VCENV_ROOT`, `VCENV_VERSION`, `PATH`, `HOME`
- `VCENV_GITHUB_TOKEN` / `GITHUB_TOKEN`, `VCENV_GITHUB_API_URL`, `VCENV_DOWNLOAD_MIRROR`
- `VCENV_AUTO_INSTALL` (`1` downgrades a missing active version from FAIL to WARN,
  since the shim installs it on first use)

Exit codes:

- `0` if no check reported `FAIL`.
- `1` if at least one check reported `FAIL`.

Example:

```sh
vc-env doctor
vc-env doctor --fix
vc-env doctor --deep
```

---

### `prune`

Purpose: Remove installed versions that nothing references. **Dry run by
default**: nothing is deleted without `--yes`.

A version is kept when it is:

- the global version (`$VCENV_ROOT/version`),
- the exported `VCENV_VERSION`,
- named in a `.vcluster-version` file under the scan roots
  ([`VCENV_PRUNE_SCAN_ROOTS`](#vcenv_prune_scan_roots), default `$HOME`),
- the currently active version (shell > local > global), or
- among the `--keep-last` newest installed versions.

Aliases are alias-aware: a reference to `0.21` or `latest` keeps the installed
version it currently resolves to. The scan skips noisy directories such as
`.git`, `node_modules`, `vendor` and `~/Library`, and never enters `VCENV_ROOT`.

Syntax:

```text
vc-env prune [--keep-last N] [--older-than DUR] [--dry-run] [--yes]
```

Options/flags:

- `--keep-last N`: also keep the N newest installed versions (default `0`)
- `--older-than DUR`: only remove versions whose directory is older than
  `DUR` (Go duration, plus `d` days and `w` weeks: `30m`, `24h`, `30d`, `12w`)
- `--dry-run`: report what would be removed (the default; wins over `--yes`)
- `--yes`: actually remove the selected versions
- `-h`, `--help`: show command help and exit

Environment variables:

- `VCENV_ROOT` (required)
- `VCENV_VERSION`
- `VCENV_PRUNE_SCAN_ROOTS`

Exit codes:

- `0` on success (including dry runs and "nothing to prune").
- `1` if not initialized, a flag value is invalid, or removal fails.

Example:

```sh
vc-env prune                          # dry run
vc-env prune --keep-last 2 --yes
VCENV_PRUNE_SCAN_ROOTS="$HOME/src" vc-env prune --older-than 30d --yes
```
---

### `completion`

Purpose: Generate a native shell completion script for `vc-env`.

The script completes subcommands and dynamically suggests versions:

- Installed versions for `uninstall`, `shell`, `local`, `global`, and `exec`.
- Cached remote versions for `install` and `resolve` (never performs a network request; falls back to the compiled-in baseline when the cache is empty).
- Shell names (`bash`, `zsh`, `fish`, `powershell`, `pwsh`) for `completion`.
- Flags for `doctor` and `prune`.

Syntax:

```text
vc-env completion [<shell>] [--shell <shell>]
```

`<shell>` is one of `bash`, `zsh`, `fish`, `powershell` (or its alias `pwsh`).
When no shell is given, it is detected from the basename of `$SHELL`; if that
is unset or unsupported, the bash script is printed and a note goes to stderr.

Options/flags:

- `--shell <shell>`: same as the positional argument
- `-h`, `--help`: show command help and exit

Environment variables:

- `SHELL` (read when no shell is given)
- `VCENV_ROOT` (read at completion time by the generated scripts, via `vc-env __complete-versions`).

Exit codes:

- `0` on success.
- `1` if the shell is unsupported.

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
Add-Content -Path $PROFILE -Value 'vc-env completion pwsh | Out-String | Invoke-Expression'
```

Internal helper:

The generated scripts call `vc-env __complete-versions {installed|remote}` to list candidate versions. This command is not part of the public interface and may change at any time, but it is safe to invoke manually for debugging.

Back-compat:

`vc-env autocompletion` is a deprecated alias for `vc-env completion` and takes
the same arguments. It prints `vc-env: 'autocompletion' is deprecated; use
'vc-env completion'` to stderr. With no argument it now follows `$SHELL`
detection; existing `source <(vc-env autocompletion)` lines in `~/.bashrc`
still get the bash script when `$SHELL` is bash.
---

### `resolve`

Purpose: Resolve a version or alias to a concrete **installed** version and
print it to stdout. Without an argument, the active version (shell > local >
global) is resolved.

Used by the `vcluster` shim and handy in scripts.

Syntax:

```text
vc-env resolve [<version-or-alias>] [--install] [-s|--silent]
```

See [Version aliases](#version-aliases) for the supported syntax. Installed
versions are matched first; `latest` picks the newest installed stable
version. When nothing installed matches, the command fails, unless
`--install` (or `VCENV_AUTO_INSTALL=1`) is given: then the best remote match
is installed. Install progress goes to stderr, so stdout only ever holds the
bare version.

Options/flags:

- `--install`: install the best remote match when nothing installed matches
- `-s`, `--silent`: suppress install progress output
- `-h`, `--help`: show command help and exit

Environment variables:

- `VCENV_ROOT` (required)
- `VCENV_VERSION` (read when no argument is given)
- `VCENV_AUTO_INSTALL` (`1` implies `--install`)

Exit codes:

- `0` on success.
- `1` if not initialized, no version is configured, nothing installed matches
  (without `--install`), or installing fails.

Example:

```sh
vc-env resolve               # active version, e.g. 0.21.3
vc-env resolve 0.21          # e.g. 0.21.3 (highest installed 0.21.x)
vc-env resolve ~0.21.1       # e.g. 0.21.3
vc-env resolve 0.22 --install
```
---

## Advanced / internal

These commands are kept for compatibility and scripting. They are hidden from
`vc-env help` and from the completion command lists.

### `ensure`

Purpose: Resolve a version or alias and install the concrete version if it is
missing. Prints the concrete version to stdout. Equivalent to
`vc-env resolve <version-or-alias> --install --silent`.

Older shims generated by previous `vc-env` releases call it when
`VCENV_AUTO_INSTALL=1` is set; run `vc-env init` (or `vc-env doctor --fix`) to
regenerate the shim, which now calls `resolve` directly.

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
vc-env ensure latest
vc-env ensure 0.21
```
