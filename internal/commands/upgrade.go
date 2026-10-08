package commands

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/user/vc-env/internal/github"
	"github.com/user/vc-env/internal/platform"
	"github.com/user/vc-env/internal/semver"
)

const vcenvRepo = "mmpyro/vc-env"

// Upgrade downloads the latest stable vc-env release from GitHub and replaces
// the current binary in-place using an atomic rename.
func Upgrade() error {
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to determine executable path: %w", err)
	}
	binaryPath, err := filepath.EvalSymlinks(execPath)
	if err != nil {
		return fmt.Errorf("failed to resolve symlinks for %s: %w", execPath, err)
	}

	client := github.NewClient()
	latestVersion, err := client.GetLatestReleaseFor(vcenvRepo)
	if err != nil {
		return fmt.Errorf("failed to fetch latest vc-env release: %w", err)
	}

	if Version != "dev" {
		current := semver.Parse(Version)
		remote := semver.Parse(latestVersion)

		if !semver.Less(current, remote) {
			if current.Original == remote.Original ||
				(current.Major == remote.Major && current.Minor == remote.Minor && current.Patch == remote.Patch && current.PreRelease == remote.PreRelease) {
				fmt.Printf("vc-env is already up to date (version %s)\n", Version)
				return nil
			}
			fmt.Printf("Current version %s is newer than latest release %s, skipping upgrade\n", Version, latestVersion)
			return nil
		}
	} else {
		fmt.Println("Running a dev build — proceeding with upgrade to latest release")
	}

	info, err := platform.Detect()
	if err != nil {
		return fmt.Errorf("failed to detect platform: %w", err)
	}

	return upgradeWithClient(client, binaryPath, info, latestVersion)
}

// upgradeWithClient performs the download, checksum verification and atomic
// replace steps of the upgrade flow. It is split from Upgrade so tests can
// drive the download through an httptest server.
func upgradeWithClient(client *github.Client, binaryPath string, info platform.Info, latestVersion string) error {
	url := client.DownloadURL(platform.SelfDownloadPath(latestVersion, info, vcenvRepo))
	fmt.Printf("Downloading vc-env %s for %s/%s...\n", latestVersion, info.OS, info.Arch)

	data, err := client.DownloadBinary(url)
	if err != nil {
		return fmt.Errorf("failed to download vc-env %s: %w", latestVersion, err)
	}

	checksumURL := client.DownloadURL(platform.SelfChecksumPath(latestVersion, vcenvRepo))
	checksumData, err := client.DownloadBinary(checksumURL)
	if err != nil {
		fmt.Printf("Warning: could not download checksums for vc-env %s: %v\n", latestVersion, err)
	} else {
		expectedChecksum, err := findChecksum(string(checksumData), platform.SelfBinaryName(info))
		if err != nil {
			fmt.Printf("Warning: could not find checksum for %s in checksums.txt\n", platform.SelfBinaryName(info))
		} else {
			actualChecksum := sha256.Sum256(data)
			actualChecksumStr := hex.EncodeToString(actualChecksum[:])
			if actualChecksumStr != expectedChecksum {
				return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedChecksum, actualChecksumStr)
			}
			fmt.Println("Checksum verified successfully")
		}
	}

	if err := atomicReplace(binaryPath, data); err != nil {
		return err
	}

	if Version != "dev" {
		fmt.Printf("Upgraded vc-env from %s to %s\n", Version, latestVersion)
	} else {
		fmt.Printf("Upgraded vc-env to %s\n", latestVersion)
	}
	return nil
}

// atomicReplace writes data to a temporary file in the same directory as
// targetPath, then atomically renames it over the target.
func atomicReplace(targetPath string, data []byte) error {
	dir := filepath.Dir(targetPath)

	tmpFile, err := os.CreateTemp(dir, ".vc-env-upgrade-*")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %w (do you have write permission to %s?)", err, dir)
	}
	tmpPath := tmpFile.Name()

	defer func() {
		if err != nil {
			os.Remove(tmpPath)
		}
	}()

	if _, err = tmpFile.Write(data); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to write temporary file: %w", err)
	}
	if err = tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temporary file: %w", err)
	}

	if err = os.Chmod(tmpPath, 0o755); err != nil {
		return fmt.Errorf("failed to set permissions on temporary file: %w", err)
	}

	if err = os.Rename(tmpPath, targetPath); err != nil {
		if fallbackErr := copyFile(tmpPath, targetPath); fallbackErr != nil {
			return fmt.Errorf("failed to replace binary (rename: %w, copy fallback: %v)", err, fallbackErr)
		}
		os.Remove(tmpPath)
	}

	return nil
}

// copyFile copies src to dst, preserving executable permissions.
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o755)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err = io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}
