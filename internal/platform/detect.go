// Package platform provides OS and architecture detection for downloading
// the correct vcluster binary.
package platform

import (
	"fmt"
	"runtime"
)

// Info holds the detected platform information.
type Info struct {
	OS   string
	Arch string
}

// Detect returns the current platform's OS and architecture.
func Detect() (Info, error) {
	osName, err := mapOS(runtime.GOOS)
	if err != nil {
		return Info{}, err
	}
	archName, err := mapArch(runtime.GOARCH)
	if err != nil {
		return Info{}, err
	}
	return Info{OS: osName, Arch: archName}, nil
}

// mapOS maps Go's runtime.GOOS to the vcluster release naming convention.
func mapOS(goos string) (string, error) {
	switch goos {
	case "linux":
		return "linux", nil
	case "darwin":
		return "darwin", nil
	default:
		return "", fmt.Errorf("unsupported operating system: %s", goos)
	}
}

// mapArch maps Go's runtime.GOARCH to the vcluster release naming convention.
func mapArch(goarch string) (string, error) {
	switch goarch {
	case "amd64":
		return "amd64", nil
	case "arm64":
		return "arm64", nil
	default:
		return "", fmt.Errorf("unsupported architecture: %s", goarch)
	}
}

// DownloadPath returns the path part of the GitHub release download URL.
func DownloadPath(version string, info Info) string {
	return fmt.Sprintf(
		"loft-sh/vcluster/releases/download/v%s/%s",
		version, BinaryName(info),
	)
}

// BinaryName returns the vcluster binary name for the given platform.
func BinaryName(info Info) string {
	name := fmt.Sprintf("vcluster-%s-%s", info.OS, info.Arch)
	if info.OS == "windows" {
		name += ".exe"
	}
	return name
}

// ChecksumPath returns the path part of the URL for the checksums.txt file.
func ChecksumPath(version string) string {
	return fmt.Sprintf(
		"loft-sh/vcluster/releases/download/v%s/checksums.txt",
		version,
	)
}
