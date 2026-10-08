// Package semver provides semantic version parsing and sorting utilities.
package semver

import (
	"sort"
	"strconv"
	"strings"
)

// Version represents a parsed semantic version.
type Version struct {
	Major      int
	Minor      int
	Patch      int
	PreRelease string // e.g. "alpha", "alpha.1", "beta.2"
	Original   string // the original string as passed in
}

// Parse parses a semantic version string (with or without leading "v").
// Returns a zero Version with Original set if parsing fails.
func Parse(s string) Version {
	original := s
	s = strings.TrimPrefix(s, "v")

	// Split on "-" to separate pre-release
	parts := strings.SplitN(s, "-", 2)
	core := parts[0]
	preRelease := ""
	if len(parts) == 2 {
		preRelease = parts[1]
	}

	nums := strings.Split(core, ".")
	if len(nums) != 3 {
		return Version{Original: original}
	}

	major, err1 := strconv.Atoi(nums[0])
	minor, err2 := strconv.Atoi(nums[1])
	patch, err3 := strconv.Atoi(nums[2])
	if err1 != nil || err2 != nil || err3 != nil {
		return Version{Original: original}
	}

	return Version{
		Major:      major,
		Minor:      minor,
		Patch:      patch,
		PreRelease: preRelease,
		Original:   original,
	}
}

// Less reports whether v is less than w in semver precedence.
// Pre-release versions have lower precedence than the release version:
//
//	0.31.1-alpha < 0.31.1
func Less(v, w Version) bool {
	if v.Major != w.Major {
		return v.Major < w.Major
	}
	if v.Minor != w.Minor {
		return v.Minor < w.Minor
	}
	if v.Patch != w.Patch {
		return v.Patch < w.Patch
	}
	// Same major.minor.patch — compare pre-release.
	// No pre-release > any pre-release (e.g. 0.31.1 > 0.31.1-alpha).
	if v.PreRelease == "" && w.PreRelease != "" {
		return false // v is the release, w is pre-release → v > w
	}
	if v.PreRelease != "" && w.PreRelease == "" {
		return true // v is pre-release, w is the release → v < w
	}
	// Both have pre-release: compare lexicographically.
	return v.PreRelease < w.PreRelease
}

// IsAlias reports whether s is a version alias rather than a concrete
// semantic version.  Recognised aliases:
//
//   - "latest", "latest-stable", "latest-prerelease"
//   - "MAJOR.MINOR" (e.g. "0.21") — highest patch of that minor
//   - "~MAJOR.MINOR.PATCH" (e.g. "~0.21.1") — >=0.21.1, <0.22.0
//
// Concrete versions like "0.21.1" or "v0.21.1-alpha" return false.
func IsAlias(s string) bool {
	switch s {
	case "latest", "latest-stable", "latest-prerelease":
		return true
	}

	// Tilde range: "~MAJOR.MINOR.PATCH"
	if strings.HasPrefix(s, "~") {
		rest := strings.TrimPrefix(s, "~")
		rest = strings.TrimPrefix(rest, "v")
		parts := strings.Split(rest, ".")
		if len(parts) != 3 {
			return false
		}
		for _, p := range parts {
			if _, err := strconv.Atoi(p); err != nil {
				return false
			}
		}
		return true
	}

	// Partial version: "MAJOR.MINOR" (two components, both numeric).
	trimmed := strings.TrimPrefix(s, "v")
	parts := strings.Split(trimmed, ".")
	if len(parts) == 2 {
		for _, p := range parts {
			if _, err := strconv.Atoi(p); err != nil {
				return false
			}
		}
		return true
	}

	return false
}

// MatchAlias returns the highest-precedence candidate that satisfies the
// alias.  Only the structural aliases "MAJOR.MINOR" and "~MAJOR.MINOR.PATCH"
// are supported here; "latest*" aliases must be resolved by the caller
// against the appropriate (stable or prerelease) list.
//
// Pre-release candidates are excluded.  Returns ("", false) when no
// candidate matches or the alias is not structural.
func MatchAlias(alias string, candidates []string) (string, bool) {
	// Tilde range: match MAJOR.MINOR and version >= floor.
	if strings.HasPrefix(alias, "~") {
		rest := strings.TrimPrefix(alias, "~")
		trimmed := strings.TrimPrefix(rest, "v")
		if len(strings.Split(trimmed, ".")) != 3 {
			return "", false
		}
		floor := Parse(rest)
		// Reject non-numeric components (Parse returns zero Version with just
		// Original set when any component fails to parse).
		if floor.Major == 0 && floor.Minor == 0 && floor.Patch == 0 && floor.PreRelease == "" &&
			trimmed != "0.0.0" {
			return "", false
		}

		best := Version{}
		found := false
		for _, c := range candidates {
			v := Parse(c)
			if v.PreRelease != "" {
				continue
			}
			if v.Major != floor.Major || v.Minor != floor.Minor {
				continue
			}
			if Less(v, floor) {
				continue
			}
			if !found || Less(best, v) {
				best = v
				found = true
			}
		}
		if !found {
			return "", false
		}
		return best.Original, true
	}

	// Partial MAJOR.MINOR: highest patch of that minor.
	trimmed := strings.TrimPrefix(alias, "v")
	parts := strings.Split(trimmed, ".")
	if len(parts) == 2 {
		major, err1 := strconv.Atoi(parts[0])
		minor, err2 := strconv.Atoi(parts[1])
		if err1 != nil || err2 != nil {
			return "", false
		}
		best := Version{}
		found := false
		for _, c := range candidates {
			v := Parse(c)
			if v.PreRelease != "" {
				continue
			}
			if v.Major != major || v.Minor != minor {
				continue
			}
			if !found || Less(best, v) {
				best = v
				found = true
			}
		}
		if !found {
			return "", false
		}
		return best.Original, true
	}

	return "", false
}

// SortDescending sorts a slice of version strings from newest to oldest
// using semantic versioning rules.  Strings that cannot be parsed are placed
// at the end in their original order.
func SortDescending(versions []string) []string {
	parsed := make([]Version, len(versions))
	for i, v := range versions {
		parsed[i] = Parse(v)
	}

	sort.SliceStable(parsed, func(i, j int) bool {
		// Descending: newer (greater) versions come first.
		return Less(parsed[j], parsed[i])
	})

	result := make([]string, len(parsed))
	for i, v := range parsed {
		result[i] = v.Original
	}
	return result
}
