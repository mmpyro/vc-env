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
	"github.com/user/vc-env/internal/semver"
)

// InstallOptions controls Install behaviour.
type InstallOptions struct {
	// Version is the version to install. May be a concrete version
	// (e.g. "0.21.1"), empty (install latest stable), or an alias
	// understood by vc-env resolve (e.g. "latest", "0.21", "~0.21.1").
	// When FromFile is set, Version MUST be a concrete version.
	Version string

	// Silent suppresses progress bars and informational messages.
	Silent bool

	// FromFile, when non-empty, is the path to a local vcluster binary to
	// install instead of downloading from GitHub.  Checksum discovery is
	// skipped; use SHA256 to verify integrity.
	FromFile string

	// SHA256, when non-empty, is the expected hex-encoded SHA-256 digest of
	// the binary bytes.  If set, verification uses this value instead of
	// the one published alongside the GitHub release.  A mismatch aborts
	// the install.
	SHA256 string
}

// Install downloads and installs a specific vcluster version.
// If version is empty, it fetches the latest stable release.
func Install(version string, silent bool) error {
	return InstallWith(InstallOptions{Version: version, Silent: silent})
}

// InstallWith installs vcluster according to opts.
func InstallWith(opts InstallOptions) error {
	return installWithOptions(github.NewClient(), opts)
}

// installWithClient preserves the previous testing entry point.
func installWithClient(client *github.Client, version string, silent bool) error {
	return installWithOptions(client, InstallOptions{Version: version, Silent: silent})
}

func installWithOptions(client *github.Client, opts InstallOptions) error {
	if err := config.RequireInit(); err != nil {
		return err
	}

	version := opts.Version

	if opts.FromFile != "" {
		if version == "" {
			return fmt.Errorf("--from-file requires an explicit <version> argument")
		}
		if semver.IsAlias(version) {
			return fmt.Errorf("--from-file requires a concrete version, got alias %q", version)
		}
		return installFromFile(client, version, opts)
	}

	// If no version specified, fetch latest.
	if version == "" {
		latest, err := client.GetLatestRelease()
		if err != nil {
			return fmt.Errorf("failed to fetch latest version: %w", err)
		}
		version = latest
		if !opts.Silent {
			fmt.Printf("Latest version: %s\n", version)
		}
	} else if semver.IsAlias(version) {
		// Resolve alias to a concrete version using the shared resolver.
		concrete, err := resolveAlias(client, version)
		if err != nil {
			return err
		}
		if !opts.Silent {
			fmt.Printf("Alias %s resolves to %s\n", version, concrete)
		}
		version = concrete
	}

	// Check if already installed.
	installed, err := config.IsVersionInstalled(version)
	if err != nil {
		return err
	}
	if installed {
		if !opts.Silent {
			fmt.Printf("version %s already installed skipping\n", version)
		}
		return nil
	}

	// Detect platform.
	info, err := platform.Detect()
	if err != nil {
		return fmt.Errorf("failed to detect platform: %w", err)
	}

	// Construct download URL.
	url := client.DownloadURL(platform.DownloadPath(version, info))
	if !opts.Silent {
		fmt.Printf("Downloading vcluster %s for %s/%s...\n", version, info.OS, info.Arch)
	}

	// Download binary (with or without progress bar).
	var data []byte
	if opts.Silent {
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

	// Checksum validation.
	if opts.SHA256 != "" {
		if err := verifySHA256(data, opts.SHA256); err != nil {
			return err
		}
		if !opts.Silent {
			fmt.Println("Checksum verified successfully")
		}
	} else {
		checksumPath := platform.ChecksumPath(version)
		checksumURL := client.DownloadURL(checksumPath)
		checksumData, cerr := client.DownloadBinary(checksumURL)
		if cerr != nil {
			if !opts.Silent {
				fmt.Printf("Warning: could not download checksums for version %s: %v\n", version, cerr)
			}
		} else {
			expectedChecksum, ferr := findChecksum(string(checksumData), platform.BinaryName(info))
			if ferr != nil {
				if !opts.Silent {
					fmt.Printf("Warning: could not find checksum for %s in checksums.txt\n", platform.BinaryName(info))
				}
			} else {
				if err := verifySHA256(data, expectedChecksum); err != nil {
					return err
				}
				if !opts.Silent {
					fmt.Println("Checksum verified successfully")
				}
			}
		}
	}

	return writeBinary(version, data, opts.Silent)
}

// installFromFile installs a vcluster binary from a user-supplied local path.
// It never touches the network; integrity may optionally be checked via
// opts.SHA256.
func installFromFile(_ *github.Client, version string, opts InstallOptions) error {
	data, err := os.ReadFile(opts.FromFile)
	if err != nil {
		return fmt.Errorf("failed to read --from-file %s: %w", opts.FromFile, err)
	}

	if opts.SHA256 != "" {
		if err := verifySHA256(data, opts.SHA256); err != nil {
			return err
		}
		if !opts.Silent {
			fmt.Println("Checksum verified successfully")
		}
	} else if !opts.Silent {
		fmt.Println("Warning: no --sha256 provided; skipping integrity check")
	}

	// Skip the "already installed" fast-path for --from-file: the user is
	// explicitly asking us to (re)place the binary from a local file.
	if !opts.Silent {
		fmt.Printf("Installing vcluster %s from %s\n", version, opts.FromFile)
	}
	return writeBinary(version, data, opts.Silent)
}

// verifySHA256 compares the SHA-256 of data against the expected hex digest.
func verifySHA256(data []byte, expectedHex string) error {
	actual := sha256.Sum256(data)
	actualHex := hex.EncodeToString(actual[:])
	if !strings.EqualFold(actualHex, expectedHex) {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedHex, actualHex)
	}
	return nil
}

// writeBinary creates the version directory and writes the binary bytes.
func writeBinary(version string, data []byte, silent bool) error {
	versionDir, err := config.GetVersionDir(version)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(versionDir, 0o755); err != nil {
		return fmt.Errorf("failed to create version directory: %w", err)
	}

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
