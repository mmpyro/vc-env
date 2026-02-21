package commands

import (
	"fmt"

	"github.com/user/vc-env/internal/github"
)

// ListRemoteHelp prints the help message for the list-remote command.
func ListRemoteHelp() {
	fmt.Println(`Usage: vc-env list-remote [flags]

List all available versions of vcluster cli from GitHub releases.
Versions are printed from newest to oldest.

Flags:
  --prerelease   Include pre-release versions (e.g. alpha, beta, rc)
  -h, --help      Show this help message`)
}

// ListRemote prints all available vcluster versions from GitHub releases.
// It does NOT require init — only queries GitHub.
func ListRemote(includePrerelease bool) error {
	client := github.NewClient()
	versions, err := client.ListReleases(includePrerelease)
	if err != nil {
		return fmt.Errorf("failed to fetch remote versions: %w", err)
	}

	for _, v := range versions {
		fmt.Println(v)
	}

	return nil
}
