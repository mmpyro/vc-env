package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/user/vc-env/internal/config"
	"github.com/user/vc-env/internal/semver"
)

// Status provides an overview of the current vc-env environment.
func Status() error {
	root, ok := config.GetVCEnvRoot()
	if !ok {
		fmt.Println("vc-env is not initialized (VCENV_ROOT is not set).")
		return nil
	}

	fmt.Printf("VCENV_ROOT: %s\n", root)

	version, source := getActiveVersionWithSource()
	if version == "" {
		fmt.Println("Active version: none")
	} else {
		fmt.Printf("Active version: %s (set by %s)\n", version, source)
		binaryPath, _ := config.GetBinaryPath(version)
		fmt.Printf("Binary path:    %s\n", binaryPath)
	}

	versionsDir := filepath.Join(root, "versions")
	entries, err := os.ReadDir(versionsDir)
	if err == nil {
		var installed []string
		for _, entry := range entries {
			if entry.IsDir() {
				installed = append(installed, entry.Name())
			}
		}
		installed = semver.SortDescending(installed)
		fmt.Printf("Installed versions (%d):\n", len(installed))
		for _, v := range installed {
			if v == version {
				fmt.Printf("  * %s\n", v)
			} else {
				fmt.Printf("    %s\n", v)
			}
		}
	} else {
		fmt.Println("Installed versions: none")
	}

	return nil
}

func getActiveVersionWithSource() (string, string) {
	// 1. Check VCENV_VERSION env var (shell version)
	if v := os.Getenv("VCENV_VERSION"); v != "" {
		return v, "VCENV_VERSION environment variable"
	}

	// 2. Walk up directories looking for .vcluster-version (local version)
	dir, err := os.Getwd()
	if err == nil {
		if v, err := config.FindLocalVersionFrom(dir); err == nil && v != "" {
			return v, ".vcluster-version file"
		}
	}

	// 3. Check global version file
	root, ok := config.GetVCEnvRoot()
	if ok {
		if v, err := config.ReadGlobalVersion(root); err == nil && v != "" {
			return v, "global version file"
		}
	}

	return "", ""
}
