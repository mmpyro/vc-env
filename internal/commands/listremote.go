package commands

import (
	"fmt"

	"github.com/user/vc-env/internal/github"
)

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
