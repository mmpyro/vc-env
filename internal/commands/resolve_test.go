package commands

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/user/vc-env/internal/github"
)

func TestResolve(t *testing.T) {
	releases := []github.Release{
		{TagName: "v0.32.0-alpha.1", Prerelease: true, Draft: false},
		{TagName: "v0.31.0", Prerelease: false, Draft: false},
		{TagName: "v0.30.0", Prerelease: false, Draft: false},
		{TagName: "v0.21.3", Prerelease: false, Draft: false},
		{TagName: "v0.21.1", Prerelease: false, Draft: false},
	}

	newServer := func(t *testing.T) *httptest.Server {
		return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(releases); err != nil {
				t.Fatalf("encode: %v", err)
			}
		}))
	}

	t.Run("concrete version is echoed verbatim", func(t *testing.T) {
		t.Setenv("VCENV_ROOT", t.TempDir())
		out := captureStdout(t, func() {
			if err := resolveWithClient(nil, "0.21.1"); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		if strings.TrimSpace(out) != "0.21.1" {
			t.Fatalf("expected 0.21.1, got %q", out)
		}
	})

	t.Run("latest resolves via remote list", func(t *testing.T) {
		t.Setenv("VCENV_ROOT", t.TempDir())
		server := newServer(t)
		defer server.Close()
		client := &github.Client{BaseURL: server.URL, HTTPClient: server.Client()}

		out := captureStdout(t, func() {
			if err := resolveWithClient(client, "latest"); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		if strings.TrimSpace(out) != "0.31.0" {
			t.Fatalf("expected 0.31.0, got %q", out)
		}
	})

	t.Run("latest-prerelease resolves via remote list", func(t *testing.T) {
		t.Setenv("VCENV_ROOT", t.TempDir())
		server := newServer(t)
		defer server.Close()
		client := &github.Client{BaseURL: server.URL, HTTPClient: server.Client()}

		out := captureStdout(t, func() {
			if err := resolveWithClient(client, "latest-prerelease"); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		if strings.TrimSpace(out) != "0.32.0-alpha.1" {
			t.Fatalf("expected 0.32.0-alpha.1, got %q", out)
		}
	})

	t.Run("MAJOR.MINOR prefers installed over remote", func(t *testing.T) {
		root := t.TempDir()
		t.Setenv("VCENV_ROOT", root)
		// Install only 0.21.2 locally; remote has 0.21.3.  Installed should win.
		installed := filepath.Join(root, "versions", "0.21.2")
		if err := os.MkdirAll(installed, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(installed, "vcluster"), []byte("x"), 0o755); err != nil {
			t.Fatal(err)
		}

		server := newServer(t)
		defer server.Close()
		client := &github.Client{BaseURL: server.URL, HTTPClient: server.Client()}

		out := captureStdout(t, func() {
			if err := resolveWithClient(client, "0.21"); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		if strings.TrimSpace(out) != "0.21.2" {
			t.Fatalf("expected 0.21.2 (installed), got %q", out)
		}
	})

	t.Run("MAJOR.MINOR falls back to remote when nothing installed matches", func(t *testing.T) {
		t.Setenv("VCENV_ROOT", t.TempDir())
		server := newServer(t)
		defer server.Close()
		client := &github.Client{BaseURL: server.URL, HTTPClient: server.Client()}

		out := captureStdout(t, func() {
			if err := resolveWithClient(client, "0.21"); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		if strings.TrimSpace(out) != "0.21.3" {
			t.Fatalf("expected 0.21.3, got %q", out)
		}
	})

	t.Run("tilde range resolves via remote", func(t *testing.T) {
		t.Setenv("VCENV_ROOT", t.TempDir())
		server := newServer(t)
		defer server.Close()
		client := &github.Client{BaseURL: server.URL, HTTPClient: server.Client()}

		out := captureStdout(t, func() {
			if err := resolveWithClient(client, "~0.21.1"); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		if strings.TrimSpace(out) != "0.21.3" {
			t.Fatalf("expected 0.21.3, got %q", out)
		}
	})

	t.Run("no match returns error", func(t *testing.T) {
		t.Setenv("VCENV_ROOT", t.TempDir())
		server := newServer(t)
		defer server.Close()
		client := &github.Client{BaseURL: server.URL, HTTPClient: server.Client()}

		err := resolveWithClient(client, "0.99")
		if err == nil {
			t.Fatal("expected error for unmatched alias")
		}
	})

	t.Run("empty input is an error", func(t *testing.T) {
		if err := resolveWithClient(nil, ""); err == nil {
			t.Fatal("expected error for empty input")
		}
	})
}
