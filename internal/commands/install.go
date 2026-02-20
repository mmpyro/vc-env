package commands

import (
	"fmt"
	"os"

	"github.com/user/vc-env/internal/config"
	"github.com/user/vc-env/internal/github"
	"github.com/user/vc-env/internal/platform"
)

// Install downloads and installs a specific vcluster version.
// If version is empty, it fetches the latest stable release.
func Install(version string) error {
	if err := config.RequireInit(); err != nil {
		return err
	}

	client := github.NewClient()

	// If no version specified, fetch latest
	if version == "" {
		latest, err := client.GetLatestRelease()
		if err != nil {
			return fmt.Errorf("failed to fetch latest version: %w", err)
		}
		version = latest
		fmt.Printf("Latest version: %s\n", version)
	}

	// Check if already installed
	installed, err := config.IsVersionInstalled(version)
	if err != nil {
		return err
	}
	if installed {
		fmt.Printf("version %s already installed skipping\n", version)
		return nil
	}

	// Detect platform
	info, err := platform.Detect()
	if err != nil {
		return fmt.Errorf("failed to detect platform: %w", err)
	}

	// Construct download URL
	url := platform.DownloadURL(version, info)
	fmt.Printf("Downloading vcluster %s for %s/%s...\n", version, info.OS, info.Arch)

	// Download binary
	data, err := client.DownloadBinary(url)
	if err != nil {
		return fmt.Errorf("failed to download vcluster %s: %w", version, err)
	}

	// Create version directory
	versionDir, err := config.GetVersionDir(version)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(versionDir, 0o755); err != nil {
		return fmt.Errorf("failed to create version directory: %w", err)
	}

	// Write binary
	binaryPath, err := config.GetBinaryPath(version)
	if err != nil {
		return err
	}
	if err := os.WriteFile(binaryPath, data, 0o755); err != nil {
		return fmt.Errorf("failed to write binary: %w", err)
	}

	fmt.Printf("Installed vcluster %s\n", version)
	return nil
}
