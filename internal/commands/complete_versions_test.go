package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/user/vc-env/internal/cache"
)

func TestCompleteVersions_Installed(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("VCENV_ROOT", tmpDir)

	for _, v := range []string{"0.21.1", "0.22.0", "0.20.0"} {
		if err := os.MkdirAll(filepath.Join(tmpDir, "versions", v), 0o755); err != nil {
			t.Fatal(err)
		}
	}

	output := captureStdout(t, func() {
		if err := CompleteVersions([]string{"installed"}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	lines := strings.Split(strings.TrimSpace(output), "\n")
	expected := []string{"0.22.0", "0.21.1", "0.20.0"}
	if len(lines) != len(expected) {
		t.Fatalf("expected %d lines, got %d: %v", len(expected), len(lines), lines)
	}
	for i, want := range expected {
		if lines[i] != want {
			t.Errorf("line %d: want %q, got %q", i, want, lines[i])
		}
	}
}

func TestCompleteVersions_Installed_SilentWhenRootUnset(t *testing.T) {
	t.Setenv("VCENV_ROOT", "")

	output := captureStdout(t, func() {
		if err := CompleteVersions([]string{"installed"}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if strings.TrimSpace(output) != "" {
		t.Fatalf("expected empty output when VCENV_ROOT is unset, got: %q", output)
	}
}

func TestCompleteVersions_Installed_SilentWhenVersionsDirMissing(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("VCENV_ROOT", tmpDir)
	// Intentionally do not create the versions/ subdirectory.

	output := captureStdout(t, func() {
		if err := CompleteVersions([]string{"installed"}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	if strings.TrimSpace(output) != "" {
		t.Fatalf("expected empty output when versions dir missing, got: %q", output)
	}
}

func TestCompleteVersions_Remote_FromCache(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("VCENV_ROOT", tmpDir)
	// Force any TTL check to pass regardless of clock skew.
	t.Setenv("VCENV_CACHE_TTL", "1000h")

	c := cache.New(filepath.Join(tmpDir, "cache"))
	stable := []string{"0.22.0", "0.21.1"}
	pre := []string{"0.23.0-alpha.1"}
	if err := c.Save(stable, pre); err != nil {
		t.Fatalf("cache save: %v", err)
	}

	// Stable (default).
	output := captureStdout(t, func() {
		if err := CompleteVersions([]string{"remote"}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != len(stable) {
		t.Fatalf("stable: expected %d lines, got %d: %v", len(stable), len(lines), lines)
	}
	for i, want := range stable {
		if lines[i] != want {
			t.Errorf("stable line %d: want %q, got %q", i, want, lines[i])
		}
	}

	// Prerelease.
	output = captureStdout(t, func() {
		if err := CompleteVersions([]string{"remote", "--prerelease"}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	lines = strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != len(pre) {
		t.Fatalf("prerelease: expected %d lines, got %d: %v", len(pre), len(lines), lines)
	}
	for i, want := range pre {
		if lines[i] != want {
			t.Errorf("prerelease line %d: want %q, got %q", i, want, lines[i])
		}
	}
}

func TestCompleteVersions_Remote_FallsBackToBaseline(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("VCENV_ROOT", tmpDir)
	// No cache file written.

	output := captureStdout(t, func() {
		if err := CompleteVersions([]string{"remote"}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	baseline := cache.BaselineVersions()
	if len(baseline) == 0 {
		t.Skip("baseline is empty; nothing to assert")
	}
	if !strings.Contains(output, baseline[0]) {
		t.Fatalf("expected output to include baseline version %q, got: %q", baseline[0], output)
	}
}

func TestCompleteVersions_UnknownSubcommand(t *testing.T) {
	output := captureStdout(t, func() {
		if err := CompleteVersions([]string{"unknown"}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if strings.TrimSpace(output) != "" {
		t.Fatalf("expected empty output for unknown subcommand, got: %q", output)
	}
}

func TestCompleteVersions_NoArgs(t *testing.T) {
	output := captureStdout(t, func() {
		if err := CompleteVersions(nil); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if strings.TrimSpace(output) != "" {
		t.Fatalf("expected empty output for no args, got: %q", output)
	}
}
