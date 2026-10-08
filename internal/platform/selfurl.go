package platform

import (
	"fmt"
	"strings"
)

// SelfBinaryName returns the vc-env binary name for the given platform.
//
// Asset naming convention: vc-env-{os}-{arch}
// Example: vc-env-darwin-arm64
func SelfBinaryName(info Info) string {
	name := fmt.Sprintf("vc-env-%s-%s", info.OS, info.Arch)
	if info.OS == "windows" {
		name += ".exe"
	}
	return name
}

// SelfDownloadPath returns the path part of the vc-env release download URL
// (everything after the hostname). Pair it with a *github.Client's
// DownloadURL to honour VCENV_DOWNLOAD_MIRROR and to allow test servers to
// intercept the request.
//
// Example: mmpyro/vc-env/releases/download/v1.0.0/vc-env-darwin-arm64
func SelfDownloadPath(version string, info Info, ownerRepo string) string {
	return fmt.Sprintf(
		"%s/releases/download/v%s/%s",
		ownerRepo, version, SelfBinaryName(info),
	)
}

// SelfChecksumPath returns the path part of the vc-env checksums.txt URL
// (everything after the hostname).
//
// Example: mmpyro/vc-env/releases/download/v1.0.0/checksums.txt
func SelfChecksumPath(version string, ownerRepo string) string {
	return fmt.Sprintf(
		"%s/releases/download/v%s/checksums.txt",
		ownerRepo, version,
	)
}

// SelfDownloadURL returns the download URL for a vc-env binary built for the
// given version and platform.
//
// baseURL is the hostname/path prefix that hosts the release assets (for
// example "https://github.com", or a mirror configured via
// VCENV_DOWNLOAD_MIRROR). Any trailing "/" is stripped so callers do not have
// to worry about producing double slashes.
//
// Example: https://github.com/mmpyro/vc-env/releases/download/v1.0.0/vc-env-darwin-arm64
func SelfDownloadURL(baseURL, version string, info Info, ownerRepo string) string {
	base := strings.TrimRight(baseURL, "/")
	return fmt.Sprintf("%s/%s", base, SelfDownloadPath(version, info, ownerRepo))
}
