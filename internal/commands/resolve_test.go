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

// newInstallServer serves a release list (0.21.3, 0.20.0), a checksums.txt
// matching binary for every platform, and binary for any other path.
func newInstallServer(t *testing.T, binary []byte) *httptest.Server {
	t.Helper()
	sum := sha256.Sum256(binary)
	sumHex := hex.EncodeToString(sum[:])
	releases := []github.Release{
		{TagName: "v0.21.3"},
		{TagName: "v0.20.0"},
	}
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, "/releases") && !strings.Contains(r.URL.Path, "/download/"):
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(releases)
		case strings.HasSuffix(r.URL.Path, "checksums.txt"):
			for _, p := range []string{"linux-amd64", "linux-arm64", "darwin-amd64", "darwin-arm64"} {
				fmt.Fprintf(w, "%s  vcluster-%s\n", sumHex, p)
			}
		default:
			w.Header().Set("Content-Length", fmt.Sprintf("%d", len(binary)))
			_, _ = w.Write(binary)
		}
	}))
}

// initRoot sets VCENV_ROOT to a fresh initialized root and returns it.
func initRoot(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	t.Setenv("VCENV_ROOT", root)
	t.Setenv("VCENV_VERSION", "")
	t.Setenv("VCENV_AUTO_INSTALL", "")
	if err := os.MkdirAll(filepath.Join(root, "versions"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(t.TempDir())
	return root
}

// fakeInstall creates versions/<v>/vcluster under root.
func fakeInstall(t *testing.T, root string, versions ...string) {
	t.Helper()
	for _, v := range versions {
		dir := filepath.Join(root, "versions", v)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "vcluster"), []byte("x"), 0o755); err != nil {
			t.Fatal(err)
		}
	}
}

func TestResolve_ExplicitSpec(t *testing.T) {
	cases := []struct {
		name      string
		installed []string
		spec      string
		want      string
	}{
		{"concrete installed version", []string{"0.21.1"}, "0.21.1", "0.21.1"},
		{"MAJOR.MINOR picks highest installed patch", []string{"0.21.1", "0.21.2", "0.22.0"}, "0.21", "0.21.2"},
		{"tilde range", []string{"0.21.0", "0.21.4"}, "~0.21.1", "0.21.4"},
		{"latest is newest installed stable", []string{"0.21.1", "0.22.0", "0.23.0-alpha.1"}, "latest", "0.22.0"},
		{"latest-stable is newest installed stable", []string{"0.21.1", "0.23.0-alpha.1"}, "latest-stable", "0.21.1"},
		{"latest-prerelease includes prereleases", []string{"0.22.0", "0.23.0-alpha.1"}, "latest-prerelease", "0.23.0-alpha.1"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root := initRoot(t)
			fakeInstall(t, root, tc.installed...)
			out := captureStdout(t, func() {
				if err := resolveWithClient(offlineClient(), ResolveOptions{Spec: tc.spec}); err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			})
			if strings.TrimSpace(out) != tc.want {
				t.Fatalf("expected %s, got %q", tc.want, out)
			}
		})
	}

	t.Run("dangling version directory does not match", func(t *testing.T) {
		root := initRoot(t)
		if err := os.MkdirAll(filepath.Join(root, "versions", "0.21.9"), 0o755); err != nil {
			t.Fatal(err)
		}
		fakeInstall(t, root, "0.21.1")
		out := captureStdout(t, func() {
			if err := resolveWithClient(offlineClient(), ResolveOptions{Spec: "0.21"}); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		if strings.TrimSpace(out) != "0.21.1" {
			t.Fatalf("expected 0.21.1, got %q", out)
		}
	})

	t.Run("unresolved alias fails without --install", func(t *testing.T) {
		root := initRoot(t)
		fakeInstall(t, root, "0.20.0")
		err := resolveWithClient(offlineClient(), ResolveOptions{Spec: "0.21"})
		if err == nil || !strings.Contains(err.Error(), "no installed version matches") {
			t.Fatalf("expected unresolved error, got %v", err)
		}
	})

	t.Run("missing concrete version fails without --install", func(t *testing.T) {
		initRoot(t)
		err := resolveWithClient(offlineClient(), ResolveOptions{Spec: "0.21.1"})
		if err == nil || !strings.Contains(err.Error(), "not installed") {
			t.Fatalf("expected not installed error, got %v", err)
		}
	})

	t.Run("requires init", func(t *testing.T) {
		t.Setenv("VCENV_ROOT", "")
		if err := resolveWithClient(offlineClient(), ResolveOptions{Spec: "0.21.1"}); err == nil {
			t.Fatal("expected error when not initialized")
		}
	})
}

func TestResolve_NoArgUsesActiveVersion(t *testing.T) {
	t.Run("global alias resolves against installed", func(t *testing.T) {
		root := initRoot(t)
		fakeInstall(t, root, "0.21.1", "0.21.2")
		if err := os.WriteFile(filepath.Join(root, "version"), []byte("0.21\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		out := captureStdout(t, func() {
			if err := resolveWithClient(offlineClient(), ResolveOptions{}); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		if strings.TrimSpace(out) != "0.21.2" {
			t.Fatalf("expected 0.21.2, got %q", out)
		}
	})

	t.Run("shell version wins", func(t *testing.T) {
		root := initRoot(t)
		fakeInstall(t, root, "0.21.1", "0.22.0")
		if err := os.WriteFile(filepath.Join(root, "version"), []byte("0.21.1\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		t.Setenv("VCENV_VERSION", "0.22.0")
		out := captureStdout(t, func() {
			if err := resolveWithClient(offlineClient(), ResolveOptions{}); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		if strings.TrimSpace(out) != "0.22.0" {
			t.Fatalf("expected 0.22.0, got %q", out)
		}
	})

	t.Run("fails when nothing configured", func(t *testing.T) {
		initRoot(t)
		err := resolveWithClient(offlineClient(), ResolveOptions{})
		if err == nil || !strings.Contains(err.Error(), "no vcluster version configured") {
			t.Fatalf("expected no version error, got %v", err)
		}
	})
}

func TestResolve_Install(t *testing.T) {
	binary := []byte("fake binary")

	run := func(t *testing.T, opts ResolveOptions) (string, string) {
		t.Helper()
		server := newInstallServer(t, binary)
		defer server.Close()
		client := &github.Client{BaseURL: server.URL, DownloadBaseURL: server.URL, HTTPClient: server.Client()}

		var stderr string
		stdout := captureStdout(t, func() {
			stderr = captureStderr(t, func() {
				if err := resolveWithClient(client, opts); err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			})
		})
		return stdout, stderr
	}

	t.Run("--install installs best remote match and prints only the version", func(t *testing.T) {
		root := initRoot(t)
		stdout, stderr := run(t, ResolveOptions{Spec: "0.21", Install: true})
		if stdout != "0.21.3\n" {
			t.Fatalf("expected bare version on stdout, got %q", stdout)
		}
		if !strings.Contains(stderr, "Installed vcluster 0.21.3") {
			t.Fatalf("expected install progress on stderr, got %q", stderr)
		}
		if _, err := os.Stat(filepath.Join(root, "versions", "0.21.3", "vcluster")); err != nil {
			t.Fatalf("binary not installed: %v", err)
		}
	})

	t.Run("--install --silent is quiet", func(t *testing.T) {
		initRoot(t)
		stdout, stderr := run(t, ResolveOptions{Spec: "0.20.0", Install: true, Silent: true})
		if stdout != "0.20.0\n" {
			t.Fatalf("expected 0.20.0, got %q", stdout)
		}
		if stderr != "" {
			t.Fatalf("expected silent stderr, got %q", stderr)
		}
	})

	t.Run("VCENV_AUTO_INSTALL=1 implies --install", func(t *testing.T) {
		root := initRoot(t)
		t.Setenv("VCENV_AUTO_INSTALL", "1")
		stdout, _ := run(t, ResolveOptions{Spec: "0.21", Silent: true})
		if stdout != "0.21.3\n" {
			t.Fatalf("expected 0.21.3, got %q", stdout)
		}
		if _, err := os.Stat(filepath.Join(root, "versions", "0.21.3", "vcluster")); err != nil {
			t.Fatalf("binary not installed: %v", err)
		}
	})

	t.Run("installed match wins over remote", func(t *testing.T) {
		root := initRoot(t)
		fakeInstall(t, root, "0.21.1")
		stdout, _ := run(t, ResolveOptions{Spec: "0.21", Install: true})
		if stdout != "0.21.1\n" {
			t.Fatalf("expected installed 0.21.1, got %q", stdout)
		}
	})
}
