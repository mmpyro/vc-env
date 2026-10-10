package commands

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/user/vc-env/internal/config"
	"github.com/user/vc-env/internal/github"
)

// ExecOptions controls Exec behaviour.
type ExecOptions struct {
	// Spec is the version or alias to run.
	Spec string

	// Args are passed to the vcluster binary.  A single leading "--" is
	// stripped.
	Args []string

	// Auto forces auto-install of a missing version; NoAuto disables it even
	// when VCENV_AUTO_INSTALL=1.  NoAuto wins when both are set.
	Auto   bool
	NoAuto bool
}

// ExecHelp prints the help message for the exec command.
func ExecHelp() {
	fmt.Println(`Usage: vc-env exec [--auto|--no-auto] <version-or-alias> [--] <args...>

Run the vcluster binary of a specific version without changing the active
version.  VCENV_VERSION is set to the concrete version for the child process,
and the child's exit code is passed through.

The version is resolved like 'vc-env resolve': installed versions are matched
first; a missing version is installed only with --auto or VCENV_AUTO_INSTALL=1.
A single "--" before the arguments is stripped, so flags meant for vcluster
are never parsed by vc-env.

Flags:
  --auto         Install the version if it is missing
  --no-auto      Never install, even when VCENV_AUTO_INSTALL=1
  -h, --help     Show this help message

Examples:
  vc-env exec 0.21.1 version
  vc-env exec 0.21 -- --help`)
}

// Exec runs a specific vcluster version without changing the active version.
// A non-zero exit of the child is returned as *exec.ExitError.
func Exec(opts ExecOptions) error {
	return execWithClient(github.NewClient(), opts)
}

func execWithClient(client *github.Client, opts ExecOptions) error {
	if err := config.RequireInit(); err != nil {
		return err
	}

	if opts.Spec == "" {
		return fmt.Errorf("version not specified. Usage: vc-env exec <version> [--] <args...>")
	}

	args := opts.Args
	if len(args) > 0 && args[0] == "--" {
		args = args[1:]
	}
	if len(args) == 0 {
		return fmt.Errorf("command not specified. Usage: vc-env exec <version> [--] <args...>")
	}

	install := !opts.NoAuto && (opts.Auto || autoInstallEnabled())

	// Keep install progress off stdout; it belongs to the child process.
	restore := func() {}
	if install {
		restore = redirectStdoutToStderr()
	}
	version, err := resolveSpec(client, opts.Spec, install, false)
	restore()
	if err != nil {
		return err
	}

	binaryPath, err := config.GetBinaryPath(version)
	if err != nil {
		return err
	}

	cmd := exec.Command(binaryPath, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// Set VCENV_VERSION so any nested shim invocation picks the same version.
	cmd.Env = append(os.Environ(), fmt.Sprintf("VCENV_VERSION=%s", version))

	return cmd.Run()
}
