# Upgrade Command — Implementation Plan

## Overview

Add a `vc-env upgrade` command that replaces the running `vc-env` binary with the newest version downloaded from the project's own GitHub releases.

---

## 1. Add `upgrade` command to `./internal/commands`

### New file: `internal/commands/upgrade.go`

The `Upgrade()` function performs these steps in order:

1. **Resolve the running binary path** — use `os.Executable()` and `filepath.EvalSymlinks()` to get the real, absolute path of the currently running `vc-env` binary. This is the file that will be replaced.

2. **Fetch latest vc-env release from GitHub** — query the GitHub API for the latest *stable* (non-draft, non-prerelease) release of **the vc-env repository** (owner/repo to be defined as a constant, e.g. `mmpyro/vc-env`).  
   Reuse the existing `github.Client` by adding a new method (see step below) or by parameterising the repo passed to `GetLatestRelease`.

3. **Compare versions** — parse the remote tag (e.g. `v0.2.0` → `0.2.0`) and the compiled-in `commands.Version` with `semver.Parse` / `semver.Less`.  
   - If they are equal, print `"vc-env is already up to date (version X)"` and return `nil`.
   - If the remote is *older*, print a warning and return (a downgrade is not expected from `upgrade`).

4. **Detect current OS/architecture** — call `platform.Detect()` (already available) to get `{OS, Arch}`.

5. **Build the asset download URL** — the release pipeline publishes assets with the naming pattern:

   ```
   vc-env-{os}-{arch}          (e.g. vc-env-darwin-arm64)
   ```

   So the URL is:

   ```
   https://github.com/<owner>/<repo>/releases/download/v<version>/vc-env-<os>-<arch>
   ```

6. **Download the new binary** — call `github.Client.DownloadBinary(url)`. This already handles long timeouts, 404 checks, and User-Agent headers.

7. **Atomic replace** — to avoid corrupting the binary mid-write:
   - Write the downloaded bytes to a temporary file **in the same directory** as the target (ensures same filesystem → rename is atomic).
   - `os.Chmod` the temp file to `0o755`.
   - `os.Rename(tempFile, targetPath)` — atomic on POSIX.

8. **Print result** — `"Upgraded vc-env from <old> to <new>"`.

#### Error handling

| Condition | Behaviour |
|---|---|
| Network / GitHub errors | Return wrapped error |
| Current version is `"dev"` | Print warning that dev builds cannot determine whether an upgrade is needed, but proceed with download (always upgrade). |
| `os.Rename` fails (e.g. cross-device) | Fall back to `copy + remove` |
| Permission denied | Suggest running with `sudo` or fixing permissions |

### Changes to `internal/github/client.go`

Add a new method or extend `GetLatestRelease` to accept an arbitrary `owner/repo` string:

```go
// GetLatestReleaseFor fetches the latest stable release for the given
// GitHub owner/repo (e.g. "mmpyro/vc-env").
func (c *Client) GetLatestReleaseFor(ownerRepo string) (string, error)
```

This keeps the existing `GetLatestRelease()` (hardcoded to `loft-sh/vcluster`) backward-compatible.

The vc-env repo to query is `mmpyro/vc-env`.

### New helper: `internal/platform/selfurl.go` (optional)

```go
// SelfDownloadURL returns the download URL for a vc-env binary.
func SelfDownloadURL(version string, info Info, ownerRepo string) string
```

This is a small helper that constructs the URL for the vc-env binary itself (as opposed to the vcluster binary returned by `DownloadURL`).

### New file: `internal/commands/upgrade_test.go`

Tests should cover:
- Printing "already up to date" when versions match.
- Detecting that the remote version is newer and proceeding.
- Skipping upgrade when remote version is older.
- Handling `Version = "dev"` (always upgrade).

---

## 2. Add `upgrade` to `./cmd/vc-env/main.go`

In the existing `switch args[0]` block, add a new case:

```go
case "upgrade":
    err = commands.Upgrade()
```

No flags or arguments are required for the initial implementation.

Also update the help text in `internal/commands/help.go` to include the `upgrade` command:

```
  upgrade       Upgrade vc-env to the latest version
```

---

## 3. Describe in `./docs/cli-reference.md`

Append a new section following the existing format:

```markdown
---

### `upgrade`

Purpose: Download the latest stable release of `vc-env` from GitHub and replace the current binary in-place.

Syntax:

\```text
vc-env upgrade
\```

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

\```sh
vc-env upgrade
\```
```

---

## How vc-env downloads its own binary

```
                          ┌────────────────────┐
                          │  vc-env upgrade     │
                          └────────┬───────────┘
                                   │
                    ┌──────────────▼──────────────────┐
                    │ 1. os.Executable() → binary path│
                    └──────────────┬──────────────────┘
                                   │
              ┌────────────────────▼─────────────────────┐
              │ 2. GET /repos/<owner>/<repo>/releases/   │
              │    latest → tag_name (e.g. "v0.3.0")     │
              └────────────────────┬─────────────────────┘
                                   │
               ┌───────────────────▼──────────────────┐
               │ 3. semver compare Version vs remote  │
               │    already up-to-date? → exit 0      │
               └───────────────────┬──────────────────┘
                                   │
              ┌────────────────────▼──────────────────┐
              │ 4. platform.Detect() → {OS, Arch}     │
              └────────────────────┬──────────────────┘
                                   │
    ┌──────────────────────────────▼───────────────────────────────┐
    │ 5. URL = github.com/<owner>/<repo>/releases/download/       │
    │          v<version>/vc-env-<os>-<arch>                      │
    └──────────────────────────────┬───────────────────────────────┘
                                   │
              ┌────────────────────▼──────────────────┐
              │ 6. DownloadBinary(url) → []byte       │
              └────────────────────┬──────────────────┘
                                   │
          ┌────────────────────────▼──────────────────────┐
          │ 7. Write temp file → chmod 0755 → os.Rename  │
          │    (atomic replace of running binary)         │
          └──────────────────────────────────────────────-┘
```

### Key design decisions

| Decision | Rationale |
|---|---|
| **Atomic rename** instead of overwriting | Prevents corruption if the process is killed mid-write or the system loses power. |
| **Same-directory temp file** | `os.Rename` is only atomic when source and target are on the same filesystem. |
| **Reuse `platform.Detect()`** | The vc-env release assets use the same `{os}-{arch}` naming as vcluster releases, so the existing detection logic works directly. |
| **No `VCENV_ROOT` required** | The upgrade replaces `os.Executable()`, not anything inside `VCENV_ROOT`. This means `upgrade` works even before `vc-env init`. |
| **`dev` version always upgrades** | During development the compiled version is `"dev"`, which cannot be compared meaningfully. Always downloading the latest is the safest default. |

---

## Resolved questions

1. **GitHub repository** — `mmpyro/vc-env`.
2. **Draft releases** — Releases are manually un-drafted after testing, so the `/releases/latest` endpoint will work correctly.
