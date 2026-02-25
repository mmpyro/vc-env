// Package github provides a client for interacting with the GitHub API
// to fetch vcluster release information.
package github

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/user/vc-env/internal/semver"
)

// Release represents a GitHub release.
type Release struct {
	TagName    string `json:"tag_name"`
	Prerelease bool   `json:"prerelease"`
	Draft      bool   `json:"draft"`
}

// Client is a GitHub API client for fetching vcluster releases.
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewClient creates a new GitHub API client with default settings.
func NewClient() *Client {
	return &Client{
		BaseURL: "https://api.github.com",
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ListReleases fetches all vcluster releases from GitHub.
// If includePrerelease is false, pre-releases are filtered out.
func (c *Client) ListReleases(includePrerelease bool) ([]string, error) {
	var allReleases []Release
	page := 1

	for {
		url := fmt.Sprintf("%s/repos/loft-sh/vcluster/releases?per_page=100&page=%d", c.BaseURL, page)
		releases, nextPage, err := c.fetchReleasesPage(url)
		if err != nil {
			return nil, err
		}

		allReleases = append(allReleases, releases...)

		if nextPage == "" {
			break
		}
		page++
	}

	var versions []string
	for _, r := range allReleases {
		if r.Draft {
			continue
		}
		if !includePrerelease && r.Prerelease {
			continue
		}
		version := strings.TrimPrefix(r.TagName, "v")
		if version != "" {
			versions = append(versions, version)
		}
	}

	return semver.SortDescending(versions), nil
}

// ListReleasesSince fetches only the vcluster releases that are strictly newer
// than sinceVersion (e.g. "0.21.0").  It stops paginating as soon as it
// encounters a version that is ≤ sinceVersion, so it typically needs only one
// API page when few or no new releases exist.
//
// If sinceVersion is empty the behaviour is identical to ListReleases.
// If includePrerelease is false, pre-releases are filtered out of the result
// (but they are still used as stop-markers during pagination).
func (c *Client) ListReleasesSince(sinceVersion string, includePrerelease bool) ([]string, error) {
	// Fast path: no anchor version — fall back to a full fetch.
	if sinceVersion == "" {
		return c.ListReleases(includePrerelease)
	}

	anchor := semver.Parse(sinceVersion)

	var collected []string
	page := 1

	for {
		url := fmt.Sprintf("%s/repos/loft-sh/vcluster/releases?per_page=100&page=%d", c.BaseURL, page)
		releases, nextPage, err := c.fetchReleasesPage(url)
		if err != nil {
			return nil, err
		}

		done := false
		for _, r := range releases {
			if r.Draft {
				continue
			}
			v := strings.TrimPrefix(r.TagName, "v")
			if v == "" {
				continue
			}
			parsed := semver.Parse(v)
			// GitHub returns releases newest-first.  Stop as soon as we reach
			// a version that is not newer than the anchor.
			if !semver.Less(anchor, parsed) {
				done = true
				break
			}
			if !includePrerelease && r.Prerelease {
				continue
			}
			collected = append(collected, v)
		}

		if done || nextPage == "" {
			break
		}
		page++
	}

	return semver.SortDescending(collected), nil
}

// GetLatestRelease fetches the latest stable release version.
func (c *Client) GetLatestRelease() (string, error) {
	url := fmt.Sprintf("%s/repos/loft-sh/vcluster/releases/latest", c.BaseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "vc-env")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch latest release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden {
		return "", fmt.Errorf("GitHub API rate limit exceeded. Please try again later")
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", fmt.Errorf("failed to parse release: %w", err)
	}

	return strings.TrimPrefix(release.TagName, "v"), nil
}

// GetLatestReleaseFor fetches the latest stable release for the given
// GitHub owner/repo (e.g. "mmpyro/vc-env").
func (c *Client) GetLatestReleaseFor(ownerRepo string) (string, error) {
	url := fmt.Sprintf("%s/repos/%s/releases/latest", c.BaseURL, ownerRepo)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "vc-env")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch latest release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden {
		return "", fmt.Errorf("GitHub API rate limit exceeded. Please try again later")
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", fmt.Errorf("failed to parse release: %w", err)
	}

	return strings.TrimPrefix(release.TagName, "v"), nil
}

// fetchReleasesPage fetches a single page of releases and returns the next page URL.
func (c *Client) fetchReleasesPage(url string) ([]Release, string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "vc-env")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("failed to fetch releases: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden {
		return nil, "", fmt.Errorf("GitHub API rate limit exceeded. Please try again later")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read response: %w", err)
	}

	var releases []Release
	if err := json.Unmarshal(body, &releases); err != nil {
		return nil, "", fmt.Errorf("failed to parse releases: %w", err)
	}

	nextPage := parseNextPageURL(resp.Header.Get("Link"))
	return releases, nextPage, nil
}

// parseNextPageURL extracts the next page URL from the Link header.
func parseNextPageURL(linkHeader string) string {
	if linkHeader == "" {
		return ""
	}

	// Link header format: <url>; rel="next", <url>; rel="last"
	re := regexp.MustCompile(`<([^>]+)>;\s*rel="next"`)
	matches := re.FindStringSubmatch(linkHeader)
	if len(matches) < 2 {
		return ""
	}
	return matches[1]
}

// DownloadBinary downloads a binary from the given URL and returns its contents.
// It uses a longer timeout than the default API client to accommodate large binaries.
func (c *Client) DownloadBinary(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create download request: %w", err)
	}
	req.Header.Set("User-Agent", "vc-env")

	// Use a dedicated client with a longer timeout for binary downloads.
	downloadClient := &http.Client{Timeout: 10 * time.Minute}
	resp, err := downloadClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download binary: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("binary not found at %s. Check that the version exists", url)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read download: %w", err)
	}

	return data, nil
}
