package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/user/vc-env/internal/config"
	"github.com/user/vc-env/internal/github"
	"github.com/user/vc-env/internal/semver"
)

// ResolveHelp prints the help message for the resolve command.
func ResolveHelp() {
	fmt.Println(`Usage: vc-env resolve <version-or-alias>

Resolve a version alias to a concrete installed or available vcluster version
and print it to stdout.  Does not install anything.

Supported aliases:
  latest, latest-stable      newest non-prerelease release
  latest-prerelease          newest release including prereleases
  MAJOR.MINOR (e.g. 0.21)    highest patch of that minor
  ~MAJOR.MINOR.PATCH         highest version >= floor, same minor

Resolution order: installed versions first, then remote releases.
A concrete version (e.g. 0.21.1) is echoed back verbatim.`)
}

// Resolve prints the concrete vcluster version an alias resolves to.
// It does NOT install anything.
func Resolve(aliasOrVersion string) error {
	return resolveWithClient(github.NewClient(), aliasOrVersion)
}

func resolveWithClient(client *github.Client, aliasOrVersion string) error {
	if aliasOrVersion == "" {
		return fmt.Errorf("version or alias not specified. Usage: vc-env resolve <version-or-alias>")
	}

	concrete, err := resolveAlias(client, aliasOrVersion)
	if err != nil {
		return err
	}
	fmt.Println(concrete)
	return nil
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
