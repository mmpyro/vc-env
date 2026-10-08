package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/user/vc-env/internal/cache"
	"github.com/user/vc-env/internal/config"
	"github.com/user/vc-env/internal/semver"
)

// CompleteVersions is a hidden helper used by the generated shell completion
// scripts. It prints one version per line to stdout and never writes to
// stderr, so completion UIs stay silent on error.
//
// Supported sub-arguments:
//
//	installed   list versions installed under $VCENV_ROOT/versions
//	remote      list versions from the on-disk release cache (no network).
//	            Falls back to the hardcoded baseline when no cache exists so
//	            fresh installs still get useful suggestions. Supports an
//	            optional "--prerelease" flag to list prerelease versions.
//
// On any failure (missing root, unreadable cache, unknown subcommand) the
// function returns nil after printing nothing: a silent completion result is
// strictly better than an error popup in the user's shell.
func CompleteVersions(args []string) error {
	if len(args) == 0 {
		return nil
	}

	switch args[0] {
	case "installed":
		printInstalledVersions()
	case "remote":
		includePrerelease := false
		for _, a := range args[1:] {
			if a == "--prerelease" {
				includePrerelease = true
			}
		}
		printCachedRemoteVersions(includePrerelease)
	}

	return nil
}

// printInstalledVersions writes one installed version per line to stdout.
// Silently returns on any error.
func printInstalledVersions() {
	root, ok := config.GetVCEnvRoot()
	if !ok {
		return
	}

	entries, err := os.ReadDir(filepath.Join(root, "versions"))
	if err != nil {
		return
	}

	var versions []string
	for _, entry := range entries {
		if entry.IsDir() {
			versions = append(versions, entry.Name())
		}
	}

	for _, v := range semver.SortDescending(versions) {
		fmt.Println(v)
	}
}

// printCachedRemoteVersions writes one cached remote version per line.
// Never performs a network request. Falls back to the hardcoded baseline if
// no cache file has been written yet.
func printCachedRemoteVersions(includePrerelease bool) {
	stable, pre, ok := cachedRemoteVersions()
	if !ok {
		// No cache yet; fall back to the compiled-in baseline so the user
		// gets at least some suggestions on a fresh install.
		if includePrerelease {
			for _, v := range cache.BaselinePrereleaseVersions() {
				fmt.Println(v)
			}
			return
		}
		for _, v := range cache.BaselineVersions() {
			fmt.Println(v)
		}
		return
	}

	versions := stable
	if includePrerelease {
		versions = pre
	}
	for _, v := range versions {
		fmt.Println(v)
	}
}
