package commands

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/user/vc-env/internal/github"
)

func TestEnsure(t *testing.T) {
	t.Run("installed alias is echoed without downloading", func(t *testing.T) {
		root := t.TempDir()
		t.Setenv("VCENV_ROOT", root)
		if err := os.MkdirAll(filepath.Join(root, "versions", "0.21.2"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, "versions", "0.21.2", "vcluster"), []byte("x"), 0o755); err != nil {
			t.Fatal(err)
		}

		// A failing client proves no network call is made.
		failClient := &github.Client{
			BaseURL:         "http://127.0.0.1:0",
			DownloadBaseURL: "http://127.0.0.1:0",
			HTTPClient:      &http.Client{},
		}

		out := captureStdout(t, func() {
			if err := ensureWithClient(failClient, "0.21"); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		if strings.TrimSpace(out) != "0.21.2" {
			t.Fatalf("expected 0.21.2, got %q", out)
		}
	})

	t.Run("missing concrete version is installed and printed", func(t *testing.T) {
		root := t.TempDir()
		t.Setenv("VCENV_ROOT", root)
		if err := os.MkdirAll(filepath.Join(root, "versions"), 0o755); err != nil {
			t.Fatal(err)
		}

		binaryData := []byte("fake binary")
		checksum := sha256.Sum256(binaryData)
		checksumStr := hex.EncodeToString(checksum[:])

		releases := []github.Release{
			{TagName: "v0.21.3", Prerelease: false, Draft: false},
			{TagName: "v0.20.0", Prerelease: false, Draft: false},
		}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case strings.Contains(r.URL.Path, "/releases") && !strings.Contains(r.URL.Path, "/latest"):
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(releases)
			case strings.HasSuffix(r.URL.Path, "checksums.txt"):
				fmt.Fprintf(w, "%s  vcluster-linux-amd64\n%s  vcluster-linux-arm64\n%s  vcluster-darwin-amd64\n%s  vcluster-darwin-arm64\n",
					checksumStr, checksumStr, checksumStr, checksumStr)
			default:
				w.Header().Set("Content-Length", fmt.Sprintf("%d", len(binaryData)))
				_, _ = w.Write(binaryData)
			}
		}))
		defer server.Close()

		client := &github.Client{
			BaseURL:         server.URL,
			DownloadBaseURL: server.URL,
			HTTPClient:      server.Client(),
		}

		out := captureStdout(t, func() {
			if err := ensureWithClient(client, "0.21"); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		if strings.TrimSpace(out) != "0.21.3" {
			t.Fatalf("expected 0.21.3, got %q", out)
		}
		if _, err := os.Stat(filepath.Join(root, "versions", "0.21.3", "vcluster")); err != nil {
			t.Fatalf("binary not installed: %v", err)
		}
	})

	t.Run("empty input is an error", func(t *testing.T) {
		if err := ensureWithClient(nil, ""); err == nil {
			t.Fatal("expected error for empty input")
		}
	})
}
