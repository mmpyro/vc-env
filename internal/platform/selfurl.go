package platform

import (
	"fmt"
	"strings"
)

// SelfDownloadURL returns the download URL for a vc-env binary built for the
// given version and platform.
//
// baseURL is the hostname/path prefix that hosts the release assets (for
// example "https://github.com", or a mirror configured via
// VCENV_DOWNLOAD_MIRROR). Any trailing "/" is stripped so callers do not have
// to worry about producing double slashes.
//
// Asset naming convention: vc-env-{os}-{arch}
// Example: https://github.com/mmpyro/vc-env/releases/download/v1.0.0/vc-env-darwin-arm64
func SelfDownloadURL(baseURL, version string, info Info, ownerRepo string) string {
	base := strings.TrimRight(baseURL, "/")
	return fmt.Sprintf(
		"%s/%s/releases/download/v%s/vc-env-%s-%s",
		base, ownerRepo, version, info.OS, info.Arch,
	)
}
