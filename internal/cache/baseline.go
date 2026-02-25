package cache

// baselineVersions is a hardcoded list of all historically known vcluster
// stable releases up to the time this version of vc-env was built.
//
// Purpose: provide a zero-network starting point so that:
//  1. The very first invocation (no disk cache yet) can return a useful list
//     without fetching all pages from GitHub.
//  2. If the network is unavailable, the command still returns something
//     meaningful rather than failing entirely.
//
// Maintenance: append new versions here when cutting a new vc-env release.
// The delta-fetch logic will automatically pick up anything newer than the
// last entry in this list, so the list does not need to be exhaustively
// up-to-date — it just needs to be a reasonable lower bound.
//
// Versions are stored newest-first (descending semver order) to match the
// output format of list-remote and to make it easy to find the newest entry.
var baselineVersions = []string{
	// v0.22.x
	"0.22.0",
	// v0.21.x
	"0.21.3",
	"0.21.2",
	"0.21.1",
	"0.21.0",
	// v0.20.x
	"0.20.0",
	// v0.19.x
	"0.19.7",
	"0.19.6",
	"0.19.5",
	"0.19.4",
	"0.19.3",
	"0.19.2",
	"0.19.1",
	"0.19.0",
	// v0.18.x
	"0.18.1",
	"0.18.0",
	// v0.17.x
	"0.17.1",
	"0.17.0",
	// v0.16.x
	"0.16.0",
	// v0.15.x
	"0.15.7",
	"0.15.6",
	"0.15.5",
	"0.15.4",
	"0.15.3",
	"0.15.2",
	"0.15.1",
	"0.15.0",
	// v0.14.x
	"0.14.2",
	"0.14.1",
	"0.14.0",
	// v0.13.x
	"0.13.0",
	// v0.12.x
	"0.12.2",
	"0.12.1",
	"0.12.0",
	// v0.11.x
	"0.11.4",
	"0.11.3",
	"0.11.2",
	"0.11.1",
	"0.11.0",
	// v0.10.x
	"0.10.5",
	"0.10.4",
	"0.10.3",
	"0.10.2",
	"0.10.1",
	"0.10.0",
	// v0.9.x
	"0.9.1",
	"0.9.0",
	// v0.8.x
	"0.8.0",
	// v0.7.x
	"0.7.0",
	// v0.6.x
	"0.6.0",
	// v0.5.x
	"0.5.0",
	// v0.4.x
	"0.4.0",
}

// baselinePrereleaseVersions is the same as baselineVersions but also includes
// known pre-release versions.  It is used when --prereleases is requested.
var baselinePrereleaseVersions = []string{
	// v0.22.x
	"0.22.0",
	"0.22.0-beta.0",
	"0.22.0-alpha.0",
	// v0.21.x
	"0.21.3",
	"0.21.2",
	"0.21.1",
	"0.21.0",
	"0.21.0-beta.0",
	"0.21.0-alpha.0",
	// v0.20.x
	"0.20.0",
	"0.20.0-beta.0",
	// v0.19.x
	"0.19.7",
	"0.19.6",
	"0.19.5",
	"0.19.4",
	"0.19.3",
	"0.19.2",
	"0.19.1",
	"0.19.0",
	// v0.18.x
	"0.18.1",
	"0.18.0",
	// v0.17.x
	"0.17.1",
	"0.17.0",
	// v0.16.x
	"0.16.0",
	// v0.15.x
	"0.15.7",
	"0.15.6",
	"0.15.5",
	"0.15.4",
	"0.15.3",
	"0.15.2",
	"0.15.1",
	"0.15.0",
	// v0.14.x
	"0.14.2",
	"0.14.1",
	"0.14.0",
	// v0.13.x
	"0.13.0",
	// v0.12.x
	"0.12.2",
	"0.12.1",
	"0.12.0",
	// v0.11.x
	"0.11.4",
	"0.11.3",
	"0.11.2",
	"0.11.1",
	"0.11.0",
	// v0.10.x
	"0.10.5",
	"0.10.4",
	"0.10.3",
	"0.10.2",
	"0.10.1",
	"0.10.0",
	// v0.9.x
	"0.9.1",
	"0.9.0",
	// v0.8.x
	"0.8.0",
	// v0.7.x
	"0.7.0",
	// v0.6.x
	"0.6.0",
	// v0.5.x
	"0.5.0",
	// v0.4.x
	"0.4.0",
}

// BaselineVersions returns a copy of the hardcoded stable version list.
// The caller receives a fresh slice and may modify it freely.
func BaselineVersions() []string {
	out := make([]string, len(baselineVersions))
	copy(out, baselineVersions)
	return out
}

// BaselinePrereleaseVersions returns a copy of the hardcoded version list
// that includes pre-release entries.
func BaselinePrereleaseVersions() []string {
	out := make([]string, len(baselinePrereleaseVersions))
	copy(out, baselinePrereleaseVersions)
	return out
}

// BaselineNewest returns the newest version present in the baseline list,
// or an empty string if the baseline is empty.  This is used as the anchor
// for delta-fetching: only releases newer than this version are fetched from
// the GitHub API.
func BaselineNewest() string {
	if len(baselineVersions) == 0 {
		return ""
	}
	// The list is stored newest-first, so index 0 is the newest.
	return baselineVersions[0]
}
