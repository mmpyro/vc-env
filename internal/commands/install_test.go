package commands

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/user/vc-env/internal/github"
)

func TestInstall(t *testing.T) {
	t.Run("fails when not initialized", func(t *testing.T) {
		t.Setenv("VCENV_ROOT", "")
		err := Install("0.31.0", true)
		if err == nil {
			t.Fatal("expected error when not initialized")
		}
	})

	t.Run("skips already installed version", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("VCENV_ROOT", tmpDir)

		version := "0.31.0"
		// Create version directory with binary
		versionDir := filepath.Join(tmpDir, "versions", version)
		if err := os.MkdirAll(versionDir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(versionDir, "vcluster"), []byte("binary"), 0o755); err != nil {
			t.Fatal(err)
		}

		output := captureStdout(t, func() {
			err := Install(version, false)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})

		if !strings.Contains(output, "already installed") {
			t.Fatalf("expected 'already installed' message, got %q", output)
		}
	})

	t.Run("silent flag suppresses output", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("VCENV_ROOT", tmpDir)
		version := "0.31.0"
		versionDir := filepath.Join(tmpDir, "versions", version)
		os.MkdirAll(versionDir, 0o755)
		os.WriteFile(filepath.Join(versionDir, "vcluster"), []byte("binary"), 0o755)

		output := captureStdout(t, func() {
			err := Install(version, true)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})

		if output != "" {
			t.Fatalf("expected no output with silent flag, got %q", output)
		}
	})

	t.Run("install with progress bar and checksum", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("VCENV_ROOT", tmpDir)
		// Initialize versions directory to satisfy config.RequireInit()
		if err := os.MkdirAll(filepath.Join(tmpDir, "versions"), 0o755); err != nil {
			t.Fatal(err)
		}
		version := "0.31.0"
		binaryData := []byte("fake binary content")
		checksum := sha256.Sum256(binaryData)
		checksumStr := hex.EncodeToString(checksum[:])

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasSuffix(r.URL.Path, "checksums.txt") {
				// Common platform binary names
				fmt.Fprintf(w, "%s  vcluster-linux-amd64\n%s  vcluster-linux-arm64\n%s  vcluster-darwin-amd64\n%s  vcluster-darwin-arm64\n", checksumStr, checksumStr, checksumStr, checksumStr)
				return
			}
			w.Header().Set("Content-Length", fmt.Sprintf("%d", len(binaryData)))
			w.Write(binaryData)
		}))
		defer server.Close()

		client := &github.Client{
			BaseURL:         server.URL,
			DownloadBaseURL: server.URL,
			HTTPClient:      server.Client(),
		}

		output := captureStdout(t, func() {
			err := installWithClient(client, version, false)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})

		if !strings.Contains(output, "[##################################################] 100%") {
			t.Errorf("expected progress bar, got %q", output)
		}
		if !strings.Contains(output, "Checksum verified successfully") {
			t.Errorf("expected checksum verification message, got %q", output)
		}

		// Verify binary was written
		binaryPath := filepath.Join(tmpDir, "versions", version, "vcluster")
		if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
			t.Fatal("binary was not written")
		}
	})
}
