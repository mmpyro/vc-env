// Package github provides a client for interacting with the GitHub API
// to fetch vcluster release information.
package github

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"time"
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

	sort.Strings(versions)
	return versions, nil
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
func (c *Client) DownloadBinary(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create download request: %w", err)
	}
	req.Header.Set("User-Agent", "vc-env")

	resp, err := c.HTTPClient.Do(req)
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
