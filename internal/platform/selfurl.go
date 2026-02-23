package platform

import "fmt"

// SelfDownloadURL returns the GitHub release download URL for a vc-env binary
// built for the given version and platform.
//
// Asset naming convention: vc-env-{os}-{arch}
// Example: https://github.com/mmpyro/vc-env/releases/download/v0.2.0/vc-env-darwin-arm64
func SelfDownloadURL(version string, info Info, ownerRepo string) string {
	return fmt.Sprintf(
		"https://github.com/%s/releases/download/v%s/vc-env-%s-%s",
		ownerRepo, version, info.OS, info.Arch,
	)
}
