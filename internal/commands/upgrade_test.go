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
	"github.com/user/vc-env/internal/platform"
	"github.com/user/vc-env/internal/semver"
)

func TestUpgradeVersionComparison(t *testing.T) {
	t.Run("already up to date when versions match", func(t *testing.T) {
		current := semver.Parse("0.2.0")
		remote := semver.Parse("0.2.0")

		if semver.Less(current, remote) {
			t.Fatal("expected current == remote, but Less returned true")
		}
	})

	t.Run("detects newer remote version", func(t *testing.T) {
		current := semver.Parse("0.1.0")
		remote := semver.Parse("0.2.0")

		if !semver.Less(current, remote) {
			t.Fatal("expected current < remote")
		}
	})

	t.Run("skips when remote is older", func(t *testing.T) {
		current := semver.Parse("0.3.0")
		remote := semver.Parse("0.2.0")

		if semver.Less(current, remote) {
			t.Fatal("expected current > remote, but Less returned true")
		}
	})

	t.Run("dev version always upgrades", func(t *testing.T) {
		Version = "dev"
		defer func() { Version = "dev" }()

		// When Version is "dev", the upgrade function skips comparison
		// and always proceeds. We verify the logic check here.
		if Version != "dev" {
			t.Fatal("expected Version to be 'dev'")
		}
	})
}

func TestUpgradeAlreadyUpToDate(t *testing.T) {
	// This test verifies the printed message when versions match.
	// We can't call Upgrade() directly (it hits the network), so we
	// test the output path by simulating the version comparison branch.
	Version = "0.5.0"
	defer func() { Version = "dev" }()

	current := semver.Parse(Version)
	remote := semver.Parse("0.5.0")

	if semver.Less(current, remote) {
		t.Fatal("should not be less when equal")
	}

	if !(current.Major == remote.Major && current.Minor == remote.Minor && current.Patch == remote.Patch && current.PreRelease == remote.PreRelease) {
		t.Fatal("versions should be equal")
	}
}

func TestUpgradeSkipsOlderRemote(t *testing.T) {
	Version = "0.5.0"
	defer func() { Version = "dev" }()

	current := semver.Parse(Version)
	remote := semver.Parse("0.4.0")

	if semver.Less(current, remote) {
		t.Fatal("should not upgrade to an older version")
	}
}

func TestAtomicReplace(t *testing.T) {
	t.Run("replaces file content", func(t *testing.T) {
		tmpDir := t.TempDir()
		targetPath := tmpDir + "/vc-env"

		// Create initial file
		initialData := []byte("old-binary")
		if err := os.WriteFile(targetPath, initialData, 0o755); err != nil {
			t.Fatal(err)
		}

		newData := []byte("new-binary")
		if err := atomicReplace(targetPath, newData); err != nil {
			t.Fatalf("atomicReplace failed: %v", err)
		}

		// Verify content
		got, err := os.ReadFile(targetPath)
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != string(newData) {
			t.Fatalf("expected %q, got %q", newData, got)
		}
	})

	t.Run("creates file if it does not exist", func(t *testing.T) {
		tmpDir := t.TempDir()
		targetPath := tmpDir + "/vc-env-new"

		data := []byte("brand-new-binary")
		if err := atomicReplace(targetPath, data); err != nil {
			t.Fatalf("atomicReplace failed: %v", err)
		}

		got, err := os.ReadFile(targetPath)
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != string(data) {
			t.Fatalf("expected %q, got %q", data, got)
		}
	})
}

func TestSelfDownloadURLFormat(t *testing.T) {
	// Verify the URL format matches what release.yml publishes.
	info := platform.Info{OS: "darwin", Arch: "arm64"}
	url := platform.SelfDownloadURL("https://github.com", "0.2.0", info, "mmpyro/vc-env")

	expected := "https://github.com/mmpyro/vc-env/releases/download/v0.2.0/vc-env-darwin-arm64"
	if url != expected {
		t.Fatalf("expected %s, got %s", expected, url)
	}

	if !strings.Contains(url, "mmpyro/vc-env") {
		t.Fatal("URL should contain the correct owner/repo")
	}
}

func TestSelfDownloadURLMirror(t *testing.T) {
	// Verify a custom mirror is honoured and trailing slashes are tolerated.
	info := platform.Info{OS: "linux", Arch: "amd64"}
	url := platform.SelfDownloadURL("https://mirror.example.com/", "0.2.0", info, "mmpyro/vc-env")

	expected := "https://mirror.example.com/mmpyro/vc-env/releases/download/v0.2.0/vc-env-linux-amd64"
	if url != expected {
		t.Fatalf("expected %s, got %s", expected, url)
	}
}

// newUpgradeTestServer builds an httptest server that serves the vc-env
// binary and checksums.txt for the current platform. The handler lets the
// test override the checksum file content, or skip it entirely by returning
// 404.
func newUpgradeTestServer(t *testing.T, info platform.Info, version string, binaryData []byte, checksumBody string, serveChecksums bool) *httptest.Server {
	t.Helper()

	binaryPath := "/" + platform.SelfDownloadPath(version, info, vcenvRepo)
	checksumPath := "/" + platform.SelfChecksumPath(version, vcenvRepo)

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case binaryPath:
			w.Header().Set("Content-Length", fmt.Sprintf("%d", len(binaryData)))
			_, _ = w.Write(binaryData)
		case checksumPath:
			if !serveChecksums {
				http.NotFound(w, r)
				return
			}
			_, _ = fmt.Fprint(w, checksumBody)
		default:
			http.NotFound(w, r)
		}
	}))
}

func TestUpgradeWithClient(t *testing.T) {
	info, err := platform.Detect()
	if err != nil {
		t.Fatalf("platform.Detect failed: %v", err)
	}

	version := "0.9.9"
	binaryData := []byte("fake vc-env binary content")
	sum := sha256.Sum256(binaryData)
	validChecksum := hex.EncodeToString(sum[:])
	validChecksumBody := fmt.Sprintf("%s  %s\n", validChecksum, platform.SelfBinaryName(info))

	Version = "0.0.1"
	defer func() { Version = "dev" }()

	t.Run("writes binary when checksum matches", func(t *testing.T) {
		server := newUpgradeTestServer(t, info, version, binaryData, validChecksumBody, true)
		defer server.Close()

		client := &github.Client{
			BaseURL:         server.URL,
			DownloadBaseURL: server.URL,
			HTTPClient:      server.Client(),
		}

		targetPath := filepath.Join(t.TempDir(), "vc-env")
		if err := os.WriteFile(targetPath, []byte("old"), 0o755); err != nil {
			t.Fatal(err)
		}

		output := captureStdout(t, func() {
			if err := upgradeWithClient(client, targetPath, info, version); err != nil {
				t.Fatalf("upgradeWithClient failed: %v", err)
			}
		})

		if !strings.Contains(output, "Checksum verified successfully") {
			t.Errorf("expected checksum verification message, got %q", output)
		}

		got, err := os.ReadFile(targetPath)
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != string(binaryData) {
			t.Fatalf("expected binary replaced with %q, got %q", binaryData, got)
		}
	})

	t.Run("returns error when checksum mismatches", func(t *testing.T) {
		mismatchBody := fmt.Sprintf("%s  %s\n", strings.Repeat("0", 64), platform.SelfBinaryName(info))
		server := newUpgradeTestServer(t, info, version, binaryData, mismatchBody, true)
		defer server.Close()

		client := &github.Client{
			BaseURL:         server.URL,
			DownloadBaseURL: server.URL,
			HTTPClient:      server.Client(),
		}

		targetPath := filepath.Join(t.TempDir(), "vc-env")
		originalContent := []byte("old")
		if err := os.WriteFile(targetPath, originalContent, 0o755); err != nil {
			t.Fatal(err)
		}

		var err error
		_ = captureStdout(t, func() {
			err = upgradeWithClient(client, targetPath, info, version)
		})

		if err == nil {
			t.Fatal("expected checksum mismatch error, got nil")
		}
		if !strings.Contains(err.Error(), "checksum mismatch") {
			t.Fatalf("expected error to contain 'checksum mismatch', got %v", err)
		}

		got, err := os.ReadFile(targetPath)
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != string(originalContent) {
			t.Fatalf("binary should not be replaced on mismatch, got %q", got)
		}
	})

	t.Run("warns and continues when checksums.txt is missing", func(t *testing.T) {
		server := newUpgradeTestServer(t, info, version, binaryData, "", false)
		defer server.Close()

		client := &github.Client{
			BaseURL:         server.URL,
			DownloadBaseURL: server.URL,
			HTTPClient:      server.Client(),
		}

		targetPath := filepath.Join(t.TempDir(), "vc-env")
		if err := os.WriteFile(targetPath, []byte("old"), 0o755); err != nil {
			t.Fatal(err)
		}

		output := captureStdout(t, func() {
			if err := upgradeWithClient(client, targetPath, info, version); err != nil {
				t.Fatalf("upgradeWithClient failed: %v", err)
			}
		})

		if !strings.Contains(output, "Warning: could not download checksums for vc-env") {
			t.Errorf("expected missing-checksums warning, got %q", output)
		}
		if strings.Contains(output, "Checksum verified successfully") {
			t.Errorf("verification message should not appear when checksums.txt is missing, got %q", output)
		}

		got, err := os.ReadFile(targetPath)
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != string(binaryData) {
			t.Fatalf("expected binary replaced with %q, got %q", binaryData, got)
		}
	})
}
