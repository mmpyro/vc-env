package shim

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestGenerateShimScript(t *testing.T) {
	t.Run("creates shim script", func(t *testing.T) {
		tmpDir := t.TempDir()

		err := GenerateShimScript(tmpDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		shimPath := filepath.Join(tmpDir, "shims", "vcluster")
		info, err := os.Stat(shimPath)
		if err != nil {
			t.Fatalf("shim script not created: %v", err)
		}

		// Check it's executable
		if info.Mode()&0o111 == 0 {
			t.Fatal("shim script is not executable")
		}

		// Check content
		data, err := os.ReadFile(shimPath)
		if err != nil {
			t.Fatalf("failed to read shim: %v", err)
		}
		content := string(data)

		if !strings.HasPrefix(content, "#!/bin/sh") {
			t.Fatal("shim should start with #!/bin/sh")
		}
		if !strings.Contains(content, "VCENV_ROOT=") {
			t.Fatal("shim should contain VCENV_ROOT")
		}
		if !strings.Contains(content, "VCENV_VERSION") {
			t.Fatal("shim should check VCENV_VERSION")
		}
		if !strings.Contains(content, ".vcluster-version") {
			t.Fatal("shim should check .vcluster-version")
		}
		if !strings.Contains(content, "exec") {
			t.Fatal("shim should exec the binary")
		}
		if !strings.Contains(content, "VCENV_AUTO_INSTALL") {
			t.Fatal("shim should honor VCENV_AUTO_INSTALL")
		}
		if !strings.Contains(content, `command vc-env resolve --install --silent`) {
			t.Fatal("shim should call 'vc-env resolve --install --silent' when auto-install is enabled")
		}
		if !strings.Contains(content, `command vc-env resolve --silent`) {
			t.Fatal("shim should call 'vc-env resolve --silent' as a fallback for aliases")
		}
		if strings.Contains(content, "vc-env ensure") {
			t.Fatal("shim should no longer call 'vc-env ensure'")
		}
	})

	t.Run("creates shims directory", func(t *testing.T) {
		tmpDir := t.TempDir()

		err := GenerateShimScript(tmpDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		shimsDir := filepath.Join(tmpDir, "shims")
		info, err := os.Stat(shimsDir)
		if err != nil {
			t.Fatalf("shims directory not created: %v", err)
		}
		if !info.IsDir() {
			t.Fatal("shims should be a directory")
		}
	})
}

func TestGenerateShellInit(t *testing.T) {
	t.Run("contains PATH update", func(t *testing.T) {
		output := GenerateShellInit("/home/user/.vcenv")
		if !strings.Contains(output, `/home/user/.vcenv/shims:$PATH`) {
			t.Fatal("shell init should prepend shims to PATH")
		}
	})

	t.Run("contains shell function", func(t *testing.T) {
		output := GenerateShellInit("/home/user/.vcenv")
		if !strings.Contains(output, "vc-env()") {
			t.Fatal("shell init should define vc-env function")
		}
	})

	t.Run("intercepts shell subcommand", func(t *testing.T) {
		output := GenerateShellInit("/home/user/.vcenv")
		if !strings.Contains(output, `"$1" = "shell"`) {
			t.Fatal("shell init should intercept shell subcommand")
		}
		if !strings.Contains(output, "export VCENV_VERSION") {
			t.Fatal("shell init should export VCENV_VERSION")
		}
	})

	t.Run("delegates other commands", func(t *testing.T) {
		output := GenerateShellInit("/home/user/.vcenv")
		if !strings.Contains(output, `command vc-env "$@"`) {
			t.Fatal("shell init should delegate other commands to binary")
		}
	})

	t.Run("shell flags pass through without exporting", func(t *testing.T) {
		// POSIX sh rejects "vc-env" as a function name; the init code
		// targets bash and zsh.
		bash, err := exec.LookPath("bash")
		if err != nil || runtime.GOOS == "windows" {
			t.Skip("requires bash")
		}
		// A fake vc-env binary that just echoes its arguments.
		bin := t.TempDir()
		fake := filepath.Join(bin, "vc-env")
		if err := os.WriteFile(fake, []byte("#!/bin/sh\necho \"args:$*\"\n"), 0o755); err != nil {
			t.Fatal(err)
		}
		script := GenerateShellInit(t.TempDir()) + `
vc-env shell --help
echo "VCENV_VERSION=${VCENV_VERSION:-unset}"
`
		cmd := exec.Command(bash, "-c", script)
		cmd.Env = []string{"PATH=" + bin + ":/usr/bin:/bin"}
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("bash failed: %v\n%s", err, out)
		}
		if !strings.Contains(string(out), "args:shell --help") {
			t.Fatalf("expected help to be delegated, got:\n%s", out)
		}
		if !strings.Contains(string(out), "VCENV_VERSION=unset") {
			t.Fatalf("shell --help must not export VCENV_VERSION, got:\n%s", out)
		}
	})
}
