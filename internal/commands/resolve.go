package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/user/vc-env/internal/config"
	"github.com/user/vc-env/internal/github"
	"github.com/user/vc-env/internal/semver"
)

// ResolveOptions controls Resolve behaviour.
type ResolveOptions struct {
	// Spec is the version or alias to resolve.  When empty, the active
	// version (shell > local > global) is used.
	Spec string

	// Install installs the best remote match when no installed version
	// satisfies Spec.  Implied by VCENV_AUTO_INSTALL=1.
	Install bool

	// Silent suppresses install progress output.
	Silent bool
}

// ResolveHelp prints the help message for the resolve command.
func ResolveHelp() {
	fmt.Println(`Usage: vc-env resolve [<version-or-alias>] [--install] [-s|--silent]

Resolve a version or alias to a concrete installed vcluster version and print
it to stdout.  Without an argument, the active version (shell > local > global)
is resolved.

Supported aliases:
  latest, latest-stable      newest non-prerelease release
  latest-prerelease          newest release including prereleases
  MAJOR.MINOR (e.g. 0.21)    highest patch of that minor
  ~MAJOR.MINOR.PATCH         highest version >= floor, same minor

Installed versions are matched first.  When nothing installed matches, the
command fails unless --install (or VCENV_AUTO_INSTALL=1) is given, in which
case the best remote match is installed.  Install progress goes to stderr so
stdout only ever contains the bare version.

Flags:
  --install      Install the best remote match when nothing installed matches
  -s, --silent   Suppress install progress output
  -h, --help     Show this help message

Environment:
  VCENV_AUTO_INSTALL=1   implies --install`)
}

// Resolve prints the concrete vcluster version that opts.Spec (or the
// active version) resolves to.
func Resolve(opts ResolveOptions) error {
	return resolveWithClient(github.NewClient(), opts)
}

func resolveWithClient(client *github.Client, opts ResolveOptions) error {
	if err := config.RequireInit(); err != nil {
		return err
	}

	spec := opts.Spec
	if spec == "" {
		active, _ := getActiveVersionWithSource()
		if active == "" {
			return fmt.Errorf("no vcluster version configured. Set one with 'vc-env global <version>' or pass a version")
		}
		spec = active
	}

	install := opts.Install || autoInstallEnabled()

	// Install progress must never pollute stdout: callers such as the shim
	// capture it as the resolved version.
	restore := func() {}
	if install {
		restore = redirectStdoutToStderr()
	}
	concrete, err := resolveSpec(client, spec, install, opts.Silent)
	restore()
	if err != nil {
		return err
	}

	fmt.Println(concrete)
	return nil
}

// autoInstallEnabled reports whether VCENV_AUTO_INSTALL=1 is set.  The shim
// uses the same exact-match check.
func autoInstallEnabled() bool {
	return os.Getenv("VCENV_AUTO_INSTALL") == "1"
}

// redirectStdoutToStderr points os.Stdout at os.Stderr and returns a function
// that restores the original.
func redirectStdoutToStderr() func() {
	old := os.Stdout
	os.Stdout = os.Stderr
	return func() { os.Stdout = old }
}

// resolveSpec resolves spec to a concrete version, preferring installed
// versions.  When nothing installed matches and install is true, the best
// remote match is installed; otherwise an error is returned.
func resolveSpec(client *github.Client, spec string, install, silent bool) (string, error) {
	if match, ok := matchInstalled(spec); ok {
		return match, nil
	}

	if !install {
		if semver.IsAlias(spec) {
			return "", fmt.Errorf("no installed version matches %q. Run 'vc-env install %s' or pass --install", spec, spec)
		}
		return "", fmt.Errorf("version %s is not installed. Run 'vc-env install %s' or pass --install", spec, spec)
	}

	concrete, err := resolveAlias(client, spec)
	if err != nil {
		return "", err
	}
	if err := installWithOptions(client, InstallOptions{Version: concrete, Silent: silent}); err != nil {
		return "", fmt.Errorf("auto-install of %s failed: %w", concrete, err)
	}
	return concrete, nil
}

// matchInstalled returns the installed version that spec selects, without
// touching the network.  Named aliases pick the newest installed version
// (stable only, except for latest-prerelease).
func matchInstalled(spec string) (string, bool) {
	if !semver.IsAlias(spec) {
		ok, _ := config.IsVersionInstalled(spec)
		return spec, ok
	}

	all, _ := listInstalledVersions()
	var installed []string
	for _, v := range all {
		if ok, _ := config.IsVersionInstalled(v); ok {
			installed = append(installed, v)
		}
	}

	switch spec {
	case "latest", "latest-stable":
		for _, v := range semver.SortDescending(installed) {
			if semver.Parse(v).PreRelease == "" {
				return v, true
			}
		}
		return "", false
	case "latest-prerelease":
		sorted := semver.SortDescending(installed)
		if len(sorted) == 0 {
			return "", false
		}
		return sorted[0], true
	}

	return semver.MatchAlias(spec, installed)
}

// resolveAlias returns the concrete vcluster version string that the given
// input resolves to.  If the input is already a concrete version it is
// returned unchanged.
//
// Resolution strategy for structural aliases (MAJOR.MINOR, ~X.Y.Z):
//  1. Try installed versions first (fast, offline).
//  2. Fall back to the remote cache / GitHub release list.
//
// For named aliases (latest*, latest-stable, latest-prerelease) the remote
// list is always consulted.
func resolveAlias(client *github.Client, input string) (string, error) {
	if !semver.IsAlias(input) {
		// Concrete version: nothing to do.
		return input, nil
	}

	// Named aliases need the remote list.
	switch input {
	case "latest", "latest-stable":
		stable, _, err := getRemoteVersions(client)
		if err != nil {
			return "", fmt.Errorf("failed to resolve %s: %w", input, err)
		}
		if len(stable) == 0 {
			return "", fmt.Errorf("no stable versions available to resolve %s", input)
		}
		return stable[0], nil

	case "latest-prerelease":
		_, pre, err := getRemoteVersions(client)
		if err != nil {
			return "", fmt.Errorf("failed to resolve %s: %w", input, err)
		}
		if len(pre) == 0 {
			return "", fmt.Errorf("no prerelease versions available to resolve %s", input)
		}
		return pre[0], nil
	}

	// Structural alias: try installed first.
	installed, _ := listInstalledVersions()
	if match, ok := semver.MatchAlias(input, installed); ok {
		return match, nil
	}

	// Fall back to remote stable list.
	stable, _, err := getRemoteVersions(client)
	if err != nil {
		return "", fmt.Errorf("failed to resolve %s: %w", input, err)
	}
	if match, ok := semver.MatchAlias(input, stable); ok {
		return match, nil
	}

	return "", fmt.Errorf("no version matches alias %q", input)
}

// listInstalledVersions returns the names of version directories under
// $VCENV_ROOT/versions.  Returns an empty slice when the root is not set or
// the directory does not exist (both are non-fatal here).
func listInstalledVersions() ([]string, error) {
	root, ok := config.GetVCEnvRoot()
	if !ok {
		return nil, nil
	}
	versionsDir := filepath.Join(root, "versions")
	entries, err := os.ReadDir(versionsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var versions []string
	for _, e := range entries {
		if e.IsDir() {
			versions = append(versions, e.Name())
		}
	}
	return versions, nil
}
