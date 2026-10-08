package github

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestListReleases(t *testing.T) {
	releases := []Release{
		{TagName: "v0.30.0", Prerelease: false, Draft: false},
		{TagName: "v0.31.0", Prerelease: false, Draft: false},
		{TagName: "v0.32.0-alpha.1", Prerelease: true, Draft: false},
		{TagName: "v0.32.0", Prerelease: false, Draft: false},
		{TagName: "v0.33.0-draft", Prerelease: false, Draft: true},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(releases); err != nil {
			t.Fatalf("failed to encode releases: %v", err)
		}
	}))
	defer server.Close()

	client := &Client{
		BaseURL:    server.URL,
		HTTPClient: server.Client(),
	}

	t.Run("excludes prereleases by default", func(t *testing.T) {
		versions, err := client.ListReleases(false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Expect newest-to-oldest (descending semver order)
		expected := []string{"0.32.0", "0.31.0", "0.30.0"}
		if len(versions) != len(expected) {
			t.Fatalf("expected %d versions, got %d: %v", len(expected), len(versions), versions)
		}
		for i, v := range versions {
			if v != expected[i] {
				t.Fatalf("expected %s at index %d, got %s", expected[i], i, v)
			}
		}
	})

	t.Run("includes prereleases when requested", func(t *testing.T) {
		versions, err := client.ListReleases(true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Expect newest-to-oldest; pre-release 0.32.0-alpha.1 < 0.32.0 so it comes after
		expected := []string{"0.32.0", "0.32.0-alpha.1", "0.31.0", "0.30.0"}
		if len(versions) != len(expected) {
			t.Fatalf("expected %d versions, got %d: %v", len(expected), len(versions), versions)
		}
		for i, v := range versions {
			if v != expected[i] {
				t.Fatalf("expected %s at index %d, got %s", expected[i], i, v)
			}
		}
	})
}

func TestListReleasesPagination(t *testing.T) {
	page1 := []Release{
		{TagName: "v0.30.0", Prerelease: false, Draft: false},
	}
	page2 := []Release{
		{TagName: "v0.31.0", Prerelease: false, Draft: false},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page := r.URL.Query().Get("page")
		w.Header().Set("Content-Type", "application/json")

		switch page {
		case "", "1":
			// Set Link header pointing to page 2
			nextURL := fmt.Sprintf("<%s/repos/loft-sh/vcluster/releases?per_page=100&page=2>; rel=\"next\"", r.URL.Scheme+"://"+r.Host)
			w.Header().Set("Link", nextURL)
			if err := json.NewEncoder(w).Encode(page1); err != nil {
				t.Fatalf("failed to encode: %v", err)
			}
		case "2":
			if err := json.NewEncoder(w).Encode(page2); err != nil {
				t.Fatalf("failed to encode: %v", err)
			}
		}
	}))
	defer server.Close()

	client := &Client{
		BaseURL:    server.URL,
		HTTPClient: server.Client(),
	}

	versions, err := client.ListReleases(false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Expect newest-to-oldest (descending semver order)
	expected := []string{"0.31.0", "0.30.0"}
	if len(versions) != len(expected) {
		t.Fatalf("expected %d versions, got %d: %v", len(expected), len(versions), versions)
	}
}

func TestGetLatestRelease(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		release := Release{TagName: "v0.32.0", Prerelease: false, Draft: false}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(release); err != nil {
			t.Fatalf("failed to encode: %v", err)
		}
	}))
	defer server.Close()

	client := &Client{
		BaseURL:    server.URL,
		HTTPClient: server.Client(),
	}

	version, err := client.GetLatestRelease()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if version != "0.32.0" {
		t.Fatalf("expected 0.32.0, got %s", version)
	}
}

func TestGetLatestReleaseRateLimit(t *testing.T) {
	t.Run("rate limited without token mentions env vars and reset", func(t *testing.T) {
		reset := time.Now().Add(2 * time.Minute).Unix()
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-RateLimit-Remaining", "0")
			w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(reset, 10))
			w.WriteHeader(http.StatusForbidden)
		}))
		defer server.Close()

		client := &Client{
			BaseURL:    server.URL,
			HTTPClient: server.Client(),
		}

		_, err := client.GetLatestRelease()
		if err == nil {
			t.Fatal("expected error on rate limit")
		}
		msg := err.Error()
		if !strings.Contains(msg, "rate limit exceeded") {
			t.Fatalf("expected rate limit error, got: %v", err)
		}
		if !strings.Contains(msg, "resets in") {
			t.Fatalf("expected 'resets in' hint, got: %v", err)
		}
		if !strings.Contains(msg, "VCENV_GITHUB_TOKEN") || !strings.Contains(msg, "GITHUB_TOKEN") {
			t.Fatalf("expected token env hint, got: %v", err)
		}
	})

	t.Run("rate limited with token mentions already set", func(t *testing.T) {
		reset := time.Now().Add(90 * time.Second).Unix()
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-RateLimit-Remaining", "0")
			w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(reset, 10))
			w.WriteHeader(http.StatusForbidden)
		}))
		defer server.Close()

		client := &Client{
			BaseURL:    server.URL,
			Token:      "abc123",
			HTTPClient: server.Client(),
		}

		_, err := client.GetLatestRelease()
		if err == nil {
			t.Fatal("expected error on rate limit")
		}
		if !strings.Contains(err.Error(), "already set") {
			t.Fatalf("expected 'already set' hint when token configured, got: %v", err)
		}
	})

	t.Run("non-rate-limit 403 returns generic status error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// No X-RateLimit headers set.
			w.WriteHeader(http.StatusForbidden)
		}))
		defer server.Close()

		client := &Client{
			BaseURL:    server.URL,
			HTTPClient: server.Client(),
		}

		_, err := client.GetLatestRelease()
		if err == nil {
			t.Fatal("expected error on 403")
		}
		msg := err.Error()
		if strings.Contains(msg, "rate limit") {
			t.Fatalf("expected generic status error without 'rate limit', got: %v", err)
		}
		if !strings.Contains(msg, "status 403") {
			t.Fatalf("expected 'status 403' in message, got: %v", err)
		}
	})
}

func TestForbiddenNotRateLimit(t *testing.T) {
	// 403 with remaining > 0 must NOT be labelled as rate limiting.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-RateLimit-Remaining", "5")
		w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(time.Hour).Unix(), 10))
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	client := &Client{
		BaseURL:    server.URL,
		HTTPClient: server.Client(),
	}

	_, err := client.GetLatestRelease()
	if err == nil {
		t.Fatal("expected error on 403")
	}
	if strings.Contains(err.Error(), "rate limit") {
		t.Fatalf("403 with remaining quota should not be labelled as rate limit, got: %v", err)
	}
	if !strings.Contains(err.Error(), "status 403") {
		t.Fatalf("expected 'status 403' in message, got: %v", err)
	}
}

func TestDownloadBinary(t *testing.T) {
	expectedData := []byte("fake-binary-data")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(expectedData)
	}))
	defer server.Close()

	client := &Client{
		BaseURL:    server.URL,
		HTTPClient: server.Client(),
	}

	data, err := client.DownloadBinary(server.URL + "/download")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(data) != string(expectedData) {
		t.Fatalf("expected %s, got %s", expectedData, data)
	}
}

func TestDownloadBinaryNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := &Client{
		BaseURL:    server.URL,
		HTTPClient: server.Client(),
	}

	_, err := client.DownloadBinary(server.URL + "/download")
	if err == nil {
		t.Fatal("expected error on 404")
	}
}

func TestParseNextPageURL(t *testing.T) {
	tests := []struct {
		name     string
		header   string
		expected string
	}{
		{
			name:     "empty header",
			header:   "",
			expected: "",
		},
		{
			name:     "with next link",
			header:   `<https://api.github.com/repos/loft-sh/vcluster/releases?page=2>; rel="next", <https://api.github.com/repos/loft-sh/vcluster/releases?page=5>; rel="last"`,
			expected: "https://api.github.com/repos/loft-sh/vcluster/releases?page=2",
		},
		{
			name:     "no next link",
			header:   `<https://api.github.com/repos/loft-sh/vcluster/releases?page=1>; rel="prev"`,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseNextPageURL(tt.header)
			if result != tt.expected {
				t.Fatalf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestNewClientReadsEnv(t *testing.T) {
	t.Run("defaults when no env vars set", func(t *testing.T) {
		t.Setenv("VCENV_GITHUB_TOKEN", "")
		t.Setenv("GITHUB_TOKEN", "")
		t.Setenv("VCENV_GITHUB_API_URL", "")
		t.Setenv("VCENV_DOWNLOAD_MIRROR", "")

		c := NewClient()
		if c.BaseURL != "https://api.github.com" {
			t.Fatalf("expected default BaseURL, got %q", c.BaseURL)
		}
		if c.DownloadBaseURL != "https://github.com" {
			t.Fatalf("expected default DownloadBaseURL, got %q", c.DownloadBaseURL)
		}
		if c.Token != "" {
			t.Fatalf("expected empty Token, got %q", c.Token)
		}
	})

	t.Run("env vars override defaults", func(t *testing.T) {
		t.Setenv("VCENV_GITHUB_TOKEN", "")
		t.Setenv("GITHUB_TOKEN", "gh-token")
		t.Setenv("VCENV_GITHUB_API_URL", "https://ghe.example.com/api/v3")
		t.Setenv("VCENV_DOWNLOAD_MIRROR", "https://mirror.example.com")

		c := NewClient()
		if c.BaseURL != "https://ghe.example.com/api/v3" {
			t.Fatalf("expected overridden BaseURL, got %q", c.BaseURL)
		}
		if c.DownloadBaseURL != "https://mirror.example.com" {
			t.Fatalf("expected overridden DownloadBaseURL, got %q", c.DownloadBaseURL)
		}
		if c.Token != "gh-token" {
			t.Fatalf("expected GITHUB_TOKEN fallback, got %q", c.Token)
		}
	})

	t.Run("VCENV_GITHUB_TOKEN wins over GITHUB_TOKEN", func(t *testing.T) {
		t.Setenv("VCENV_GITHUB_TOKEN", "vcenv-token")
		t.Setenv("GITHUB_TOKEN", "generic-token")

		c := NewClient()
		if c.Token != "vcenv-token" {
			t.Fatalf("expected VCENV_GITHUB_TOKEN to win, got %q", c.Token)
		}
	})
}

func TestAuthorizationHeaderSent(t *testing.T) {
	releases := []Release{
		{TagName: "v0.32.0", Prerelease: false, Draft: false},
	}

	t.Run("Authorization header present when token set", func(t *testing.T) {
		var seenAuths []string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			seenAuths = append(seenAuths, r.Header.Get("Authorization"))
			w.Header().Set("Content-Type", "application/json")
			switch {
			case strings.HasSuffix(r.URL.Path, "/releases/latest"):
				_ = json.NewEncoder(w).Encode(releases[0])
			case strings.Contains(r.URL.Path, "/releases"):
				_ = json.NewEncoder(w).Encode(releases)
			case strings.HasPrefix(r.URL.Path, "/download"):
				_, _ = w.Write([]byte("binary"))
			default:
				http.NotFound(w, r)
			}
		}))
		defer server.Close()

		client := &Client{
			BaseURL:    server.URL,
			Token:      "xyz",
			HTTPClient: server.Client(),
		}

		if _, err := client.GetLatestRelease(); err != nil {
			t.Fatalf("GetLatestRelease: %v", err)
		}
		if _, err := client.ListReleases(true); err != nil {
			t.Fatalf("ListReleases: %v", err)
		}
		if _, err := client.DownloadBinary(server.URL + "/download/file"); err != nil {
			t.Fatalf("DownloadBinary: %v", err)
		}

		if len(seenAuths) < 3 {
			t.Fatalf("expected at least 3 requests, got %d", len(seenAuths))
		}
		for i, got := range seenAuths {
			if got != "Bearer xyz" {
				t.Fatalf("request %d: expected Authorization 'Bearer xyz', got %q", i, got)
			}
		}
	})

	t.Run("Authorization header absent when token empty", func(t *testing.T) {
		var seenAuths []string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			seenAuths = append(seenAuths, r.Header.Get("Authorization"))
			w.Header().Set("Content-Type", "application/json")
			switch {
			case strings.HasSuffix(r.URL.Path, "/releases/latest"):
				_ = json.NewEncoder(w).Encode(releases[0])
			case strings.Contains(r.URL.Path, "/releases"):
				_ = json.NewEncoder(w).Encode(releases)
			case strings.HasPrefix(r.URL.Path, "/download"):
				_, _ = w.Write([]byte("binary"))
			default:
				http.NotFound(w, r)
			}
		}))
		defer server.Close()

		client := &Client{
			BaseURL:    server.URL,
			HTTPClient: server.Client(),
		}

		if _, err := client.GetLatestRelease(); err != nil {
			t.Fatalf("GetLatestRelease: %v", err)
		}
		if _, err := client.ListReleases(true); err != nil {
			t.Fatalf("ListReleases: %v", err)
		}
		if _, err := client.DownloadBinary(server.URL + "/download/file"); err != nil {
			t.Fatalf("DownloadBinary: %v", err)
		}

		for i, got := range seenAuths {
			if got != "" {
				t.Fatalf("request %d: expected no Authorization header, got %q", i, got)
			}
		}
	})
}
