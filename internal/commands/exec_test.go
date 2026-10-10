package commands

import (
	"errors"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/user/vc-env/internal/github"
)

// installScript writes an executable shell script as the vcluster binary of
// version under root.
func installScript(t *testing.T, root, version, script string) {
	t.Helper()
	dir := filepath.Join(root, "versions", version)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "vcluster"), []byte("#!/bin/sh\n"+script+"\n"), 0o755); err != nil {
		t.Fatal(err)
	}
}

// offlineClient fails every request, proving no network call is made.
func offlineClient() *github.Client {
	return &github.Client{
		BaseURL:         "http://127.0.0.1:0",
		DownloadBaseURL: "http://127.0.0.1:0",
		HTTPClient:      &http.Client{},
	}
}

func TestExec(t *testing.T) {
	t.Run("fails when not initialized", func(t *testing.T) {
		t.Setenv("VCENV_ROOT", "")
		err := Exec(ExecOptions{Spec: "0.31.0", Args: []string{"version"}})
		if err == nil {
			t.Fatal("expected error when not initialized")
		}
	})

	t.Run("fails when version not installed", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("VCENV_ROOT", tmpDir)
		t.Setenv("VCENV_AUTO_INSTALL", "")
		if err := os.MkdirAll(filepath.Join(tmpDir, "versions"), 0o755); err != nil {
			t.Fatal(err)
		}

		err := execWithClient(offlineClient(), ExecOptions{Spec: "0.31.0", Args: []string{"version"}})
		if err == nil || !strings.Contains(err.Error(), "not installed") {
			t.Fatalf("expected not installed error, got %v", err)
		}
	})

	t.Run("fails when version or command missing", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("VCENV_ROOT", tmpDir)
		if err := os.MkdirAll(filepath.Join(tmpDir, "versions", "0.31.0"), 0o755); err != nil {
			t.Fatal(err)
		}

		err := Exec(ExecOptions{Args: []string{"version"}})
		if err == nil || !strings.HasPrefix(err.Error(), "version not specified") {
			t.Fatalf("expected version missing error, got %v", err)
		}

		err = Exec(ExecOptions{Spec: "0.31.0"})
		if err == nil || !strings.HasPrefix(err.Error(), "command not specified") {
			t.Fatalf("expected command missing error, got %v", err)
		}

		// A lone "--" is not a command.
		err = Exec(ExecOptions{Spec: "0.31.0", Args: []string{"--"}})
		if err == nil || !strings.HasPrefix(err.Error(), "command not specified") {
			t.Fatalf("expected command missing error for lone --, got %v", err)
		}
	})
}

func TestExec_ExitCodePassthrough(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("requires /bin/sh")
	}
	root := t.TempDir()
	t.Setenv("VCENV_ROOT", root)
	installScript(t, root, "0.21.1", "exit 3")

	err := execWithClient(offlineClient(), ExecOptions{Spec: "0.21.1", Args: []string{"version"}})
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("expected *exec.ExitError, got %T: %v", err, err)
	}
	if exitErr.ExitCode() != 3 {
		t.Fatalf("expected exit code 3, got %d", exitErr.ExitCode())
	}
}

func TestExec_StripsDoubleDashAndSetsVersion(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("requires /bin/sh")
	}
	root := t.TempDir()
	t.Setenv("VCENV_ROOT", root)
	t.Setenv("VCENV_VERSION", "")
	out := filepath.Join(root, "out.txt")
	installScript(t, root, "0.21.3", `echo "$VCENV_VERSION|$*" > "`+out+`"`)

	// The alias resolves offline to the installed 0.21.3; only the first
	// "--" is stripped.
	err := execWithClient(offlineClient(), ExecOptions{Spec: "0.21", Args: []string{"--", "version", "--", "-x"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(string(data)); got != "0.21.3|version -- -x" {
		t.Fatalf("unexpected child view %q", got)
	}
}

func TestExec_AutoInstall(t *testing.T) {
	root := t.TempDir()
	t.Setenv("VCENV_ROOT", root)
	if err := os.MkdirAll(filepath.Join(root, "versions"), 0o755); err != nil {
		t.Fatal(err)
	}

	t.Run("--no-auto beats VCENV_AUTO_INSTALL", func(t *testing.T) {
		t.Setenv("VCENV_AUTO_INSTALL", "1")
		err := execWithClient(offlineClient(), ExecOptions{Spec: "0.21.1", Args: []string{"version"}, NoAuto: true})
		if err == nil || !strings.Contains(err.Error(), "not installed") {
			t.Fatalf("expected not installed error, got %v", err)
		}
	})

	t.Run("--auto installs the missing version", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("requires /bin/sh")
		}
		t.Setenv("VCENV_AUTO_INSTALL", "")
		server := newInstallServer(t, []byte("#!/bin/sh\nexit 0\n"))
		defer server.Close()
		client := &github.Client{BaseURL: server.URL, DownloadBaseURL: server.URL, HTTPClient: server.Client()}

		stdout := captureStdout(t, func() {
			if err := execWithClient(client, ExecOptions{Spec: "0.21", Args: []string{"version"}, Auto: true}); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
		if stdout != "" {
			t.Fatalf("install progress leaked to stdout: %q", stdout)
		}
		if _, err := os.Stat(filepath.Join(root, "versions", "0.21.3", "vcluster")); err != nil {
			t.Fatalf("expected 0.21.3 to be installed: %v", err)
		}
	})
}

func TestExecHelp(t *testing.T) {
	out := captureStdout(t, ExecHelp)
	for _, want := range []string{"vc-env exec", "--auto", "--no-auto", "--"} {
		if !strings.Contains(out, want) {
			t.Errorf("exec help missing %q:\n%s", want, out)
		}
	}
}
