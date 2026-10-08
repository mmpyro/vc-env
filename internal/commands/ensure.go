package commands

import (
	"fmt"

	"github.com/user/vc-env/internal/config"
	"github.com/user/vc-env/internal/github"
)

// EnsureHelp prints the help message for the ensure command.
func EnsureHelp() {
	fmt.Println(`Usage: vc-env ensure <version-or-alias>

Resolve a version alias to a concrete vcluster version and ensure it is
installed.  Prints the resolved concrete version to stdout.

This is primarily called by the vcluster shim when VCENV_AUTO_INSTALL=1
and the requested version is not yet installed.  Running it manually is
equivalent to:

  CONCRETE=$(vc-env resolve "$VERSION") && vc-env install --silent "$CONCRETE"

See 'vc-env resolve --help' for the supported alias syntax.`)
}

// Ensure resolves aliasOrVersion to a concrete version, installs it if it is
// not already present, and prints the concrete version to stdout.
//
// It intentionally keeps stdout quiet (install runs in silent mode) so the
// shim can capture the final line as the resolved version.
func Ensure(aliasOrVersion string) error {
	return ensureWithClient(github.NewClient(), aliasOrVersion)
}

func ensureWithClient(client *github.Client, aliasOrVersion string) error {
	if aliasOrVersion == "" {
		return fmt.Errorf("version or alias not specified. Usage: vc-env ensure <version-or-alias>")
	}

	concrete, err := resolveAlias(client, aliasOrVersion)
	if err != nil {
		return err
	}

	installed, err := config.IsVersionInstalled(concrete)
	if err != nil {
		return err
	}
	if !installed {
		opts := InstallOptions{Version: concrete, Silent: true}
		if err := installWithOptions(client, opts); err != nil {
			return fmt.Errorf("auto-install of %s failed: %w", concrete, err)
		}
	}

	fmt.Println(concrete)
	return nil
}
