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
		if err := os.MkdirAll(versionDir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(versionDir, "vcluster"), []byte("binary"), 0o755); err != nil {
			t.Fatal(err)
		}

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
			_, _ = w.Write(binaryData)
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

func TestInstallWithOptions(t *testing.T) {
	t.Run("alias resolves to concrete and installs", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("VCENV_ROOT", tmpDir)
		if err := os.MkdirAll(filepath.Join(tmpDir, "versions"), 0o755); err != nil {
			t.Fatal(err)
		}

		binaryData := []byte("fake binary content")
		checksum := sha256.Sum256(binaryData)
		checksumStr := hex.EncodeToString(checksum[:])

		releases := []github.Release{
			{TagName: "v0.21.3", Prerelease: false, Draft: false},
			{TagName: "v0.21.0", Prerelease: false, Draft: false},
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

		err := installWithOptions(client, InstallOptions{Version: "0.21", Silent: true})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		binaryPath := filepath.Join(tmpDir, "versions", "0.21.3", "vcluster")
		if _, err := os.Stat(binaryPath); err != nil {
			t.Fatalf("expected alias to install 0.21.3: %v", err)
		}
	})

	t.Run("from-file installs local binary", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("VCENV_ROOT", tmpDir)
		if err := os.MkdirAll(filepath.Join(tmpDir, "versions"), 0o755); err != nil {
			t.Fatal(err)
		}

		src := filepath.Join(tmpDir, "src", "vcluster")
		if err := os.MkdirAll(filepath.Dir(src), 0o755); err != nil {
			t.Fatal(err)
		}
		data := []byte("local binary bytes")
		if err := os.WriteFile(src, data, 0o755); err != nil {
			t.Fatal(err)
		}

		checksum := sha256.Sum256(data)
		checksumStr := hex.EncodeToString(checksum[:])

		err := installWithOptions(nil, InstallOptions{
			Version:  "0.42.0",
			Silent:   true,
			FromFile: src,
			SHA256:   checksumStr,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		dst := filepath.Join(tmpDir, "versions", "0.42.0", "vcluster")
		got, err := os.ReadFile(dst)
		if err != nil {
			t.Fatalf("binary not written: %v", err)
		}
		if string(got) != string(data) {
			t.Fatal("installed binary bytes differ from source")
		}
	})

	t.Run("from-file with wrong sha256 aborts", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("VCENV_ROOT", tmpDir)
		if err := os.MkdirAll(filepath.Join(tmpDir, "versions"), 0o755); err != nil {
			t.Fatal(err)
		}

		src := filepath.Join(tmpDir, "src", "vcluster")
		if err := os.MkdirAll(filepath.Dir(src), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(src, []byte("bytes"), 0o755); err != nil {
			t.Fatal(err)
		}

		err := installWithOptions(nil, InstallOptions{
			Version:  "0.42.0",
			Silent:   true,
			FromFile: src,
			SHA256:   strings.Repeat("0", 64),
		})
		if err == nil {
			t.Fatal("expected checksum mismatch error")
		}
		if !strings.Contains(err.Error(), "checksum mismatch") {
			t.Fatalf("expected checksum mismatch error, got %v", err)
		}

		// Binary must NOT have been written.
		if _, err := os.Stat(filepath.Join(tmpDir, "versions", "0.42.0", "vcluster")); !os.IsNotExist(err) {
			t.Fatal("binary should not have been written on checksum failure")
		}
	})

	t.Run("from-file rejects alias version", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("VCENV_ROOT", tmpDir)
		if err := os.MkdirAll(filepath.Join(tmpDir, "versions"), 0o755); err != nil {
			t.Fatal(err)
		}

		err := installWithOptions(nil, InstallOptions{
			Version:  "latest",
			Silent:   true,
			FromFile: "/nonexistent",
		})
		if err == nil {
			t.Fatal("expected error for alias with --from-file")
		}
	})

	t.Run("explicit sha256 on remote download is used instead of checksums.txt", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("VCENV_ROOT", tmpDir)
		if err := os.MkdirAll(filepath.Join(tmpDir, "versions"), 0o755); err != nil {
			t.Fatal(err)
		}

		binaryData := []byte("content")
		expected := sha256.Sum256(binaryData)
		expectedHex := hex.EncodeToString(expected[:])

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasSuffix(r.URL.Path, "checksums.txt") {
				t.Errorf("checksums.txt should not be fetched when --sha256 is provided; path=%s", r.URL.Path)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Length", fmt.Sprintf("%d", len(binaryData)))
			_, _ = w.Write(binaryData)
		}))
		defer server.Close()

		client := &github.Client{
			BaseURL:         server.URL,
			DownloadBaseURL: server.URL,
			HTTPClient:      server.Client(),
		}

		err := installWithOptions(client, InstallOptions{
			Version: "0.21.1",
			Silent:  true,
			SHA256:  expectedHex,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if _, err := os.Stat(filepath.Join(tmpDir, "versions", "0.21.1", "vcluster")); err != nil {
			t.Fatalf("binary not written: %v", err)
		}
	})
}
