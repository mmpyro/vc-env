// Package commands implements all vc-env CLI commands.
package commands

import "fmt"

// Help prints the help message with all available commands.
func Help() {
	fmt.Println(`Usage: vc-env <command> [arguments]

Commands:
  help            Display this help message and all available commands
  list            List all installed versions of vcluster cli
  list-remote     List all available versions of vcluster cli from GitHub
  init            Initialize vc-env setup
  install         Install a specific version or alias (or latest if not specified).
                  Flags: -s/--silent, --from-file PATH, --sha256 HEX
  uninstall       Uninstall a specific version
  shell           Set or show the shell version of vcluster cli (accepts aliases)
  local           Set or show the local version of vcluster cli (accepts aliases)
  global          Set or show the global version of vcluster cli (accepts aliases)
  latest          Print the latest available version of vcluster cli from GitHub releases.
  which           Print the full path to the active vcluster binary
  exec            Run vcluster at a specific version: exec [--auto|--no-auto] <version> [--] <args...>
  resolve         Print the concrete installed version for a version, alias, or the
                  active version. Flags: --install, -s/--silent
  status          Show current vc-env environment status
  doctor          Diagnose the vc-env environment. Flags: --fix, --deep
  prune           Remove unreferenced installed versions (dry run unless --yes).
                  Flags: --keep-last N, --older-than DUR, --dry-run, --yes
  upgrade         Upgrade vc-env to the latest release
  completion      Generate a shell completion script (bash|zsh|fish|powershell)
  version         Print the version of vc-env

Run 'vc-env <command> --help' for details on a command.

Version aliases (accepted by install/global/local/shell/resolve/exec):
  latest, latest-stable    newest non-prerelease release
  latest-prerelease        newest release including prereleases
  MAJOR.MINOR  (0.21)      highest patch of that minor
  ~MAJOR.MINOR.PATCH       highest version >= floor, same minor

Environment:
  VCENV_ROOT               required; where versions, shims, and config live
  VCENV_VERSION            optional; forces a specific version (highest priority)
  VCENV_AUTO_INSTALL=1     install missing versions on demand (shim, resolve, exec)
  VCENV_GITHUB_TOKEN       GitHub token for API requests (falls back to GITHUB_TOKEN)
  VCENV_GITHUB_API_URL     GitHub API base URL (default https://api.github.com)
  VCENV_DOWNLOAD_MIRROR    download base URL (default https://github.com)
  VCENV_CACHE_TTL          release cache TTL as a Go duration (default 1h)
  VCENV_PRUNE_SCAN_ROOTS   roots scanned by 'prune' for .vcluster-version files
                           (path-list separated; default $HOME; empty disables)`)
}

// InstallHelp prints help for the install command.
func InstallHelp() {
	fmt.Println(`Usage: vc-env install [version-or-alias] [flags]

Install a vcluster version by downloading it from GitHub, or install a
pre-downloaded binary with --from-file (air-gapped setups).

Accepted versions:
  <empty>                  install the latest stable release
  0.21.1 / v0.21.1         concrete semver
  latest, latest-stable    resolved to newest non-prerelease
  latest-prerelease        resolved to newest including prereleases
  0.21                     resolved to the highest 0.21.x
  ~0.21.1                  resolved to the highest 0.21.y with y >= 1

Flags:
  -s, --silent             Do not display progress bar or checksum info
  --from-file PATH         Install the given local file as the vcluster binary.
                           Requires a concrete <version> argument and skips any
                           network calls. Use --sha256 to verify integrity.
  --sha256 HEX             Expected hex-encoded SHA-256 of the binary.
                           Overrides checksums.txt from the GitHub release.`)
}
