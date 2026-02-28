package commands

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	"github.com/user/vc-env/internal/config"
	"github.com/user/vc-env/internal/github"
	"github.com/user/vc-env/internal/platform"
)

// Install downloads and installs a specific vcluster version.
// If version is empty, it fetches the latest stable release.
func Install(version string, silent bool) error {
	client := github.NewClient()
	return installWithClient(client, version, silent)
}

func installWithClient(client *github.Client, version string, silent bool) error {
	if err := config.RequireInit(); err != nil {
		return err
	}

	// If no version specified, fetch latest
	if version == "" {
		latest, err := client.GetLatestRelease()
		if err != nil {
			return fmt.Errorf("failed to fetch latest version: %w", err)
		}
		version = latest
		if !silent {
			fmt.Printf("Latest version: %s\n", version)
		}
	}

	// Check if already installed
	installed, err := config.IsVersionInstalled(version)
	if err != nil {
		return err
	}
	if installed {
		if !silent {
			fmt.Printf("version %s already installed skipping\n", version)
		}
		return nil
	}

	// Detect platform
	info, err := platform.Detect()
	if err != nil {
		return fmt.Errorf("failed to detect platform: %w", err)
	}

	// Construct download URL
	url := client.DownloadURL(platform.DownloadPath(version, info))
	if !silent {
		fmt.Printf("Downloading vcluster %s for %s/%s...\n", version, info.OS, info.Arch)
	}

	// Download binary with progress
	var data []byte
	if silent {
		data, err = client.DownloadBinary(url)
	} else {
		data, err = client.DownloadWithProgress(url, func(total, current int64) {
			if total <= 0 {
				fmt.Printf("\rDownloaded: %d bytes", current)
				return
			}
			percent := float64(current) / float64(total) * 100
			blocks := int(percent / 2) // 50 blocks
			bar := strings.Repeat("#", blocks) + strings.Repeat(" ", 50-blocks)
			fmt.Printf("\r[%s] %.0f%%", bar, percent)
			if current == total {
				fmt.Println()
			}
		})
	}
	if err != nil {
		return fmt.Errorf("failed to download vcluster %s: %w", version, err)
	}

	// Checksum validation
	checksumPath := platform.ChecksumPath(version)
	checksumUrl := client.DownloadURL(checksumPath)
	checksumData, err := client.DownloadBinary(checksumUrl)
	if err != nil {
		if !silent {
			fmt.Printf("Warning: could not download checksums for version %s: %v\n", version, err)
		}
	} else {
		expectedChecksum, err := findChecksum(string(checksumData), platform.BinaryName(info))
		if err != nil {
			if !silent {
				fmt.Printf("Warning: could not find checksum for %s in checksums.txt\n", platform.BinaryName(info))
			}
		} else {
			actualChecksum := sha256.Sum256(data)
			actualChecksumStr := hex.EncodeToString(actualChecksum[:])
			if actualChecksumStr != expectedChecksum {
				return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedChecksum, actualChecksumStr)
			}
			if !silent {
				fmt.Println("Checksum verified successfully")
			}
		}
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

	if !silent {
		fmt.Printf("Installed vcluster %s\n", version)
	}
	return nil
}

func findChecksum(checksums, filename string) (string, error) {
	lines := strings.Split(checksums, "\n")
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) >= 2 && parts[1] == filename {
			return parts[0], nil
		}
	}
	return "", fmt.Errorf("checksum not found for %s", filename)
}
