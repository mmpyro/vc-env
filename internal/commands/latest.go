package commands

import (
	"fmt"

	"github.com/user/vc-env/internal/github"
)

// LatestHelp prints the help message for the latest command.
func LatestHelp() {
	fmt.Println(`Usage: vc-env latest [flags]

Print the latest available version of vcluster cli from GitHub releases.

Flags:
  --prerelease    Include pre-release versions (e.g. alpha, beta, rc)
  -h, --help      Show this help message`)
}

// Latest prints the latest available vcluster version from GitHub releases.
// If includePrerelease is false, only the latest stable version is returned.
// If includePrerelease is true, the latest version including prereleases is returned.
// It does NOT require init — only queries GitHub.
func Latest(includePrerelease bool) error {
	client := github.NewClient()

	var version string
	var err error

	if includePrerelease {
		versions, fetchErr := client.ListReleases(true)
		if fetchErr != nil {
			return fmt.Errorf("failed to fetch remote versions: %w", fetchErr)
		}
		if len(versions) == 0 {
			return fmt.Errorf("no versions found")
		}
		version = versions[0]
	} else {
		version, err = client.GetLatestRelease()
		if err != nil {
			return fmt.Errorf("failed to fetch latest version: %w", err)
		}
	}

	fmt.Println(version)
	return nil
}
