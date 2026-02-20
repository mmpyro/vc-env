package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInit(t *testing.T) {
	t.Run("fails when VCENV_ROOT not set", func(t *testing.T) {
		t.Setenv("VCENV_ROOT", "")
		err := Init()
		if err == nil {
			t.Fatal("expected error when VCENV_ROOT not set")
		}
	})

	t.Run("creates versions directory and shim", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("VCENV_ROOT", tmpDir)

		output := captureStdout(t, func() {
			err := Init()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})

		// Check versions directory was created
		versionsDir := filepath.Join(tmpDir, "versions")
		info, err := os.Stat(versionsDir)
		if err != nil {
			t.Fatalf("versions directory not created: %v", err)
		}
		if !info.IsDir() {
			t.Fatal("versions should be a directory")
		}

		// Check shim was created
		shimPath := filepath.Join(tmpDir, "shims", "vcluster")
		if _, err := os.Stat(shimPath); err != nil {
			t.Fatalf("shim not created: %v", err)
		}

		// Check shell init output
		if !strings.Contains(output, "vc-env()") {
			t.Fatal("output should contain shell function")
		}
		if !strings.Contains(output, "shims:$PATH") {
			t.Fatal("output should contain PATH update")
		}
	})
}
