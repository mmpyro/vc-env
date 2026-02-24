package commands

import (
	"fmt"
	"os"

	"github.com/user/vc-env/internal/cache"
	"github.com/user/vc-env/internal/github"
)

// ListRemoteHelp prints the help message for the list-remote command.
func ListRemoteHelp() {
	fmt.Println(`Usage: vc-env list-remote [flags]

List all available versions of vcluster cli from GitHub releases.
Versions are printed from newest to oldest.

Results are cached on disk (at $VCENV_ROOT/cache/releases.json) for one hour
by default to avoid redundant network requests.  Set VCENV_CACHE_TTL to a Go
duration string (e.g. "30m", "24h") to override the TTL.

Flags:
  --prerelease   Include pre-release versions (e.g. alpha, beta, rc)
  -h, --help      Show this help message`)
}

// ListRemote prints all available vcluster versions from GitHub releases.
// It does NOT require init — only queries GitHub (or the local cache).
//
// Caching strategy (three layers, applied in order):
//
//  1. Disk cache ($VCENV_ROOT/cache/releases.json) — if the cache file exists
//     and is younger than the TTL (default 1 h, overridable via
//     VCENV_CACHE_TTL), the cached list is returned immediately with no
//     network call.
//
//  2. Delta fetch — if the cache is stale or missing, only the releases newer
//     than the most recent known version are fetched from GitHub.  The delta
//     is merged with the existing cache (or the hardcoded baseline) and the
//     result is written back to disk.
//
//  3. Hardcoded baseline (cache.BaselineVersions) — if the network is
//     unavailable and no disk cache exists, the hardcoded list is returned
//     with a warning printed to stderr.  This ensures the command never
//     returns an empty list due to a transient network failure.
func ListRemote(includePrerelease bool) error {
	return listRemoteWithClient(github.NewClient(), includePrerelease)
}

// listRemoteWithClient is the testable core of ListRemote.  It accepts an
// injected GitHub client so tests can point it at a mock HTTP server.
func listRemoteWithClient(client *github.Client, includePrerelease bool) error {
	c := newCacheForRoot()

	// ── Layer 1: serve from disk cache if it is still fresh ──────────────
	if stableVersions, prereleaseVersions, ok := c.Load(); ok {
		versions := stableVersions
		if includePrerelease {
			versions = prereleaseVersions
		}
		for _, v := range versions {
			fmt.Println(v)
		}
		return nil
	}

	// ── Layer 2: delta fetch ──────────────────────────────────────────────
	// Read the stale cache (ignoring TTL) so we can use it as the merge base
	// and as the anchor for the delta fetch.  If no stale cache exists we fall
	// back to the hardcoded baseline.
	staleStable, stalePre, hasStale := loadStaleCache(c)

	stableAnchor := cache.BaselineNewest()
	prereleaseAnchor := cache.BaselineNewest()
	if hasStale {
		stableAnchor = cache.NewestVersion(staleStable)
		prereleaseAnchor = cache.NewestVersion(stalePre)
	}

	// Fetch only the releases newer than our anchor.
	deltaStable, errStable := client.ListReleasesSince(stableAnchor, false)
	deltaPre, errPre := client.ListReleasesSince(prereleaseAnchor, true)

	if errStable != nil || errPre != nil {
		// ── Layer 3: network unavailable — fall back to baseline / stale cache ──
		fmt.Fprintln(os.Stderr, "warning: failed to fetch remote versions; showing cached/baseline data")

		// Try to return a stale cache first (better than nothing).
		if hasStale {
			versions := staleStable
			if includePrerelease {
				versions = stalePre
			}
			for _, v := range versions {
				fmt.Println(v)
			}
			return nil
		}

		// Last resort: hardcoded baseline.
		versions := cache.BaselineVersions()
		if includePrerelease {
			versions = cache.BaselinePrereleaseVersions()
		}
		for _, v := range versions {
			fmt.Println(v)
		}
		return nil
	}

	// Merge delta with the stale cache (preferred) or the hardcoded baseline.
	var mergedStable, mergedPre []string
	if hasStale {
		mergedStable = cache.MergeWithCached(staleStable, deltaStable)
		mergedPre = cache.MergeWithCached(stalePre, deltaPre)
	} else {
		mergedStable = cache.MergeWithBaseline(deltaStable, false)
		mergedPre = cache.MergeWithBaseline(deltaPre, true)
	}

	// Persist the merged result so the next invocation is served from cache.
	// A save failure is non-fatal: we still print the result.
	if err := c.Save(mergedStable, mergedPre); err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not write version cache: %v\n", err)
	}

	versions := mergedStable
	if includePrerelease {
		versions = mergedPre
	}
	for _, v := range versions {
		fmt.Println(v)
	}
	return nil
}

// newCacheForRoot creates a Cache rooted at $VCENV_ROOT/cache.
// If VCENV_ROOT is not set the cache operates in memory-only mode (no disk
// persistence), which means every invocation performs a delta fetch.
func newCacheForRoot() *cache.Cache {
	root := os.Getenv("VCENV_ROOT")
	if root == "" {
		return cache.New("")
	}
	return cache.New(root + "/cache")
}

// loadStaleCache reads the cache file ignoring the TTL.  It returns the
// version lists and true if the file exists and is parseable, regardless of
// age.
func loadStaleCache(c *cache.Cache) (versions []string, prereleaseVersions []string, ok bool) {
	// We exploit the fact that Cache.Load respects the TTL, so we create a
	// temporary cache instance with a very large TTL to bypass expiry.
	staleReader := cache.NewWithTTL(c.Dir(), 1<<62) // effectively infinite TTL
	return staleReader.Load()
}
