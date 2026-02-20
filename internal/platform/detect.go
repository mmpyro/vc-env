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

// DownloadURL constructs the GitHub release download URL for a given version
// and platform.
func DownloadURL(version string, info Info) string {
	// vcluster release naming convention:
	// https://github.com/loft-sh/vcluster/releases/download/v{version}/vcluster-{os}-{arch}
	return fmt.Sprintf(
		"https://github.com/loft-sh/vcluster/releases/download/v%s/vcluster-%s-%s",
		version, info.OS, info.Arch,
	)
}
