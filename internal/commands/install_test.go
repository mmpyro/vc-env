package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstall(t *testing.T) {
	t.Run("fails when not initialized", func(t *testing.T) {
		t.Setenv("VCENV_ROOT", "")
		err := Install("0.31.0")
		if err == nil {
			t.Fatal("expected error when not initialized")
		}
	})

	t.Run("skips already installed version", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("VCENV_ROOT", tmpDir)

		// Create version directory with binary
		versionDir := filepath.Join(tmpDir, "versions", "0.31.0")
		if err := os.MkdirAll(versionDir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(versionDir, "vcluster"), []byte("binary"), 0o755); err != nil {
			t.Fatal(err)
		}

		output := captureStdout(t, func() {
			err := Install("0.31.0")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})

		if !strings.Contains(output, "already installed") {
			t.Fatalf("expected 'already installed' message, got %q", output)
		}
	})
}
