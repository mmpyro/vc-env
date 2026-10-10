# CLAUDE.md

This file guides Claude Code (claude.ai/code) when working in this repository.

## Overview

`vc-env` is a version manager for the `vcluster` CLI (loft-sh/vcluster), in the style of tfenv and pyenv.
- Versions are installed under `$VCENV_ROOT`. The README recommends `~/.vcenv`.
- A generated POSIX shim at `$VCENV_ROOT/shims/vcluster` picks the version on every call. Precedence: **shell** (`VCENV_VERSION`) > **local** (`.vcluster-version`, searched upward from the current directory) > **global** (`$VCENV_ROOT/version`).
- Supported version forms: exact versions plus the aliases `latest`, `0.21` and `~0.21.1`.
- Optional features: auto-install, air-gapped install (`--from-file` and `--sha256`), self-upgrade, shell completions, and a three-layer release cache.

Go 1.24.4, **standard library only**: no third-party dependencies and no `go.sum`. Keep it that way unless asked.

## Commands

```sh
make build            # build/vc-env  (ldflags: -X main.Version=$(VERSION))
make build-all        # linux/darwin × amd64/arm64 into build/
make test             # go test -v -race ./...
make test-docker      # BATS integration suite in Docker (needs Docker + network)
make install          # go install ./cmd/vc-env
make clean
make docs-build       # mkdocs build --strict into site/ (creates .venv-docs from docs/requirements.txt)
make docs-serve       # http://127.0.0.1:8000/vc-env/

go test -race -run 'TestName/subtest' -v ./internal/commands/   # single unit test
go vet ./... && golangci-lint run                               # lint (CI uses golangci-lint, default config)
```

## Architecture

**Dispatch is hand-rolled; there is no cobra.** `cmd/vc-env/main.go` runs a big `switch args[0]` that parses flags by hand (`--prerelease`, `-s/--silent`, `--from-file[=]`, `--sha256[=]`, `-h/--help`) and calls `commands.X(...)`.
- An error goes to stderr and the process exits 1.
- `main.Version` is copied into `commands.Version`.
- Special cases:
  - `autocompletion` is a hidden **deprecated alias** for `completion`. It takes the same arguments and prints a warning to stderr.
  - `__complete-versions` is an internal hook that the completion scripts call.

| Package | Role |
|---|---|
| `internal/commands` | One file per subcommand. Each public `X()` is a thin wrapper over `xWithClient(client *github.Client, ...)`. `install.go` has `InstallOptions{Version, Silent, FromFile, SHA256}`, `InstallWith` and SHA-256 verification. `ensure.go` is an advanced/internal thin wrapper for `resolve <spec> --install --silent`, hidden from help and the completion command lists. `doctor.go` runs checks with `--fix` and `--deep`; `--deep` compares each binary with its `checksums.txt` entry. `prune.go` and `disk.go` scan `.vcluster-version` files under `VCENV_PRUNE_SCAN_ROOTS`, with alias-aware matching. `duration.go` has `ParseDurationShorthand` (`30d`, `2w`). `versions_shared.go` has `getRemoteVersions` (cache, delta fetch, baseline fallback). `completion.go` holds the bash/zsh/fish/powershell scripts **embedded as string constants**. `upgrade.go` does an atomic, SHA-verified self-upgrade from `mmpyro/vc-env` |
| `internal/config` | `GetVCEnvRoot`, `RequireInit`, `ResolveVersion` (shell > local > global), `FindLocalVersionFrom`, `ReadGlobalVersion`, paths |
| `internal/semver` | `Parse`, `Less`, `IsAlias`, `MatchAlias`, `SortDescending` |
| `internal/github` | `Client{BaseURL, DownloadBaseURL, Token, HTTPClient}` with a 30s timeout. Lists releases (Link-header paging, delta `ListReleasesSince`), downloads with progress, and gives a friendly rate-limit message |
| `internal/cache` | JSON cache at `$VCENV_ROOT/cache/releases.json`. It writes atomically (temp file + rename) and uses a 1h TTL (`VCENV_CACHE_TTL`). `baseline.go` holds hard-coded known versions. See `docs/caching.md` |
| `internal/platform` | OS/arch detection. Downloads are **raw binaries** (not tarballs) from `loft-sh/vcluster/releases/download/v<ver>/vcluster-<os>-<arch>`. `selfurl.go` provides the self-upgrade URLs |
| `internal/shim` | `GenerateShimScript`: when the binary is missing, it falls back to `vc-env resolve --silent`, adding `--install` when `VCENV_AUTO_INSTALL=1`. `GenerateShellInit` adds to PATH and defines the `vc-env()` shell function that intercepts `shell <ver>` |

Runtime layout:
- `$VCENV_ROOT/versions/<ver>/vcluster`
- `$VCENV_ROOT/shims/vcluster` (generated; do not edit)
- `$VCENV_ROOT/version`
- `$VCENV_ROOT/cache/releases.json`

### Adding or changing a command

1. Add a `case` in `cmd/vc-env/main.go`.
2. Add `X()` and `XHelp()` in `internal/commands/<cmd>.go`. If it touches the network, also add an `xWithClient` variant and a `<cmd>_test.go`.
3. List it in `internal/commands/help.go`.
4. Update the completion scripts in `completion.go`, and `complete_versions.go` if it takes a version argument.
5. Update `docs/cli-reference.md`, which documents every command and env var, and the README.
6. Add BATS tests in `tests/integration.bats`.

## Testing

- **Unit tests** sit next to the code under `internal/` (`*_test.go`). `cmd/` has no tests.
  - Use the standard `testing` package with `t.Run` subtests and `net/http/httptest` servers standing in for GitHub. Do not add assertion libraries.
  - Isolate each test with `t.Setenv("VCENV_ROOT", t.TempDir())`. Also `t.Setenv` the token, API URL and mirror variables when they matter.
- **BATS integration tests** live in `tests/integration.bats` (55 tests, bats-support and bats-assert).
  - They run only in Docker through `Dockerfile.test`, with `scripts/run-bats.sh` as the entrypoint.
  - They hit real GitHub (list-remote, latest, and installing 0.21.1).

### Every user-facing change needs BATS coverage

This covers any new or changed command, flag, env var, or resolution, shim or completion behavior. Add or update `@test` cases in `tests/integration.bats`. Go unit tests alone are not enough for behavior the CLI shows to users.
- Follow the existing pattern. `setup()` gives a fresh `VCENV_ROOT=$BATS_TMPDIR/vc-env` with the shims on PATH, and `teardown()` removes it.
- Call `vc-env init`, then `run vc-env ...`, then `assert_success` and `assert_output --partial "..."` (or `--regexp`).
- Cover the error path too: `assert_failure`. Examples: a bad `--sha256`, a missing version, or an uninitialized root.

```bash
@test "vc-env <cmd> <does what>" {
	vc-env init
	run vc-env <cmd> <args>
	assert_success
	assert_output --partial "<expected>"
}
```

## Definition of done: final verification

Run these in order before calling any change complete:
1. `go vet ./...`, plus `golangci-lint run` if it is installed.
2. `make test`: unit tests with `-race`.
3. `make docs-build`, if you touched `docs/` or `mkdocs.yml`. Strict mode fails on broken links and anchors.
4. `make test-docker`: the **full BATS suite**. This is **always the last step**, and it needs Docker running and network access.

Report the real pass/fail counts from the output. If Docker is unavailable, say so plainly; do not claim success.

To iterate on one BATS test:
```sh
docker build -f Dockerfile.test -t vc-env-test .
docker run --rm -e TERM=xterm --entrypoint bats vc-env-test --filter "<test name regex>" /src/tests/integration.bats
```
Always finish with a full `make test-docker` afterwards.

## CLI contract (shared with helm-env and istioctl-env)

These three sibling repos expose the same command surface. Keep them in sync when you change one of them.

- `resolve [<spec>] [--install] [-s|--silent]`: requires init. With no spec, uses the active spec. Matches installed versions first, so `latest` means the newest installed stable version. It goes remote only with `--install` or `VCENV_AUTO_INSTALL=1`. stdout is only the bare version.
- `exec [--auto|--no-auto] <spec> [--] <cmd> [args...]`: strips one leading `--`. **Exits with the child's exit code.**
- `completion [<shell>] [--shell <shell>]`: supports bash, zsh, fish, powershell and `pwsh`, falling back to `$SHELL`.
- `-h/--help` short-circuits on **every** command through `exitOnHelp` in main.go, and must never run side effects; `upgrade -h` must not upgrade.
- `prune [--keep-last N] [--older-than DUR] [--dry-run] [--yes]`: a **dry run unless `--yes`**. The active version is always protected.
- `doctor [--fix] [--deep]`.

Note: if your shell exports `VCENV_AUTO_INSTALL=1`, tests that depend on it set or unset it explicitly. Do the same in new tests.

## Environment variables

| Variable | Purpose |
|---|---|
| `VCENV_ROOT` | Required. `init` refuses to run without it |
| `VCENV_VERSION` | Shell-level version, set by `vc-env shell` |
| `VCENV_AUTO_INSTALL` | `1`: the shim, `resolve` and `exec` install missing versions. `doctor` downgrades "active version missing" to WARN |
| `VCENV_GITHUB_TOKEN` | GitHub token; falls back to `GITHUB_TOKEN` |
| `VCENV_GITHUB_API_URL` | Overrides the API base (GitHub Enterprise) |
| `VCENV_DOWNLOAD_MIRROR` | Overrides the download base (air-gapped mirrors) |
| `VCENV_CACHE_TTL` | Overrides the 1h release-cache TTL |
| `VCENV_PRUNE_SCAN_ROOTS` | Path list `prune` scans for `.vcluster-version` files. Empty disables the scan; the default is `$HOME` |

Shell setup:
- bash/zsh: `eval "$(vc-env init)"`
- fish: `vc-env init | source`
- Completions: `vc-env completion bash|zsh|fish|powershell`

## Gotchas

- **The shim has its own fallback logic** in `internal/shim/shim.go`, which uses `resolve`. Shims generated by older releases still call `ensure`, so keep `ensure` working. If you change resolution or auto-install, check both the Go code and the shim script.
- The module path is the placeholder `github.com/user/vc-env`. The real repo is `github.com/mmpyro/vc-env`.
- The root directory name differs: the README says `~/.vcenv`, the Docker images use `.vc-env`, and BATS uses `$BATS_TMPDIR/vc-env`.
- `build/` may hold stale gitignored binaries, so don't trust them; rebuild with `make build`. `.venv-docs/` and `site/` are also gitignored.

## CI and release

- `.github/workflows/ci.yml` runs on push or PR to `main`:
  1. golangci-lint (latest).
  2. `gotestsum -- -race ./...`, with the JUnit report published by `dorny/test-reporter`.
  3. `make test-docker`.
- `.github/workflows/release.yml` runs on `v*.*.*` tags:
  1. `make build-all VERSION=<tag>`.
  2. BATS against the prebuilt linux amd64 and arm64 binaries using `Dockerfile.ci`. That image has no Go; the binary is mounted at `/src/vc-env` and run under QEMU.
  3. Writes `build/checksums.txt` with `sha256sum`.
  4. Creates a **draft** GitHub release with the 4 binaries and the checksums. It does not use goreleaser.
- `.github/workflows/docs.yml` runs `mkdocs build --strict`, then deploys to GitHub Pages from `main`. Docs versions are pinned in `docs/requirements.txt`.
- Dependabot updates gomod weekly.
