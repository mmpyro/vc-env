package commands

import (
	"fmt"

	"github.com/user/vc-env/internal/config"
	"github.com/user/vc-env/internal/github"
	"github.com/user/vc-env/internal/semver"
)

// WhichHelp prints the help message for the which command.
func WhichHelp() {
	fmt.Println(`Usage: vc-env which

Print the full path to the active vcluster binary (shell > local > global).

Flags:
  -h, --help     Show this help message`)
}

// Which prints the absolute path to the active vcluster binary.
func Which() error {
	if err := config.RequireInit(); err != nil {
		return err
	}

	version, err := config.ResolveVersion()
	if err != nil {
		return err
	}

	// Resolve aliases to a concrete installed version so the printed path
	// points at an actual binary.
	if semver.IsAlias(version) {
		concrete, rerr := resolveAlias(github.NewClient(), version)
		if rerr != nil {
			return rerr
		}
		version = concrete
	}

	binaryPath, err := config.GetBinaryPath(version)
	if err != nil {
		return err
	}

	fmt.Println(binaryPath)
	return nil
}
