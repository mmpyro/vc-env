package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/user/vc-env/internal/config"
)

// List prints all installed vcluster versions.
func List() error {
	if err := config.RequireInit(); err != nil {
		return err
	}

	root, _ := config.GetVCEnvRoot()
	versionsDir := filepath.Join(root, "versions")

	entries, err := os.ReadDir(versionsDir)
	if err != nil {
		return fmt.Errorf("failed to read versions directory: %w", err)
	}

	var versions []string
	for _, entry := range entries {
		if entry.IsDir() {
			versions = append(versions, entry.Name())
		}
	}

	sort.Strings(versions)

	for _, v := range versions {
		fmt.Println(v)
	}

	return nil
}
