package commands

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/user/vc-env/internal/github"
)

func TestLatest(t *testing.T) {
	// Just verify the function signature exists.
	t.Run("function exists", func(t *testing.T) {
		_ = Latest
	})
}

func TestLatestWithMockClient(t *testing.T) {
	releases := []github.Release{
		{TagName: "v0.30.0", Prerelease: false, Draft: false},
		{TagName: "v0.31.0", Prerelease: false, Draft: false},
		{TagName: "v0.32.0-alpha.1", Prerelease: true, Draft: false},
	}

	t.Run("latest stable version uses GetLatestRelease", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			// GetLatestRelease hits /releases/latest and returns a single Release
			release := github.Release{TagName: "v0.31.0", Prerelease: false, Draft: false}
			if err := json.NewEncoder(w).Encode(release); err != nil {
				t.Fatalf("failed to encode: %v", err)
			}
		}))
		defer server.Close()

		client := &github.Client{
			BaseURL:    server.URL,
			HTTPClient: server.Client(),
		}

		version, err := client.GetLatestRelease()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if version != "0.31.0" {
			t.Fatalf("expected 0.31.0, got %s", version)
		}
	})

	t.Run("latest with prerelease uses ListReleases and returns first", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(releases); err != nil {
				t.Fatalf("failed to encode: %v", err)
			}
		}))
		defer server.Close()

		client := &github.Client{
			BaseURL:    server.URL,
			HTTPClient: server.Client(),
		}

		versions, err := client.ListReleases(true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(versions) == 0 {
			t.Fatal("expected at least one version")
		}
		// The first element should be the newest (sorted descending)
		if versions[0] != "0.32.0-alpha.1" {
			t.Fatalf("expected 0.32.0-alpha.1 as latest with prerelease, got %s", versions[0])
		}
	})

	t.Run("error handling when GitHub API fails", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		client := &github.Client{
			BaseURL:    server.URL,
			HTTPClient: server.Client(),
		}

		_, err := client.GetLatestRelease()
		if err == nil {
			t.Fatal("expected error for non-200 response, got nil")
		}
	})
}
