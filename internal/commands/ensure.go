package commands

import (
	"fmt"

	"github.com/user/vc-env/internal/github"
)

// EnsureHelp prints the help message for the ensure command.
func EnsureHelp() {
	fmt.Println(`Usage: vc-env ensure <version-or-alias>

Advanced / internal.  Resolve a version or alias to a concrete vcluster
version, install it if missing, and print the concrete version to stdout.

Equivalent to:

  vc-env resolve <version-or-alias> --install --silent

See 'vc-env resolve --help' for the supported alias syntax.`)
}

// Ensure resolves aliasOrVersion to a concrete version, installs it if it is
// not already present, and prints the concrete version to stdout.
func Ensure(aliasOrVersion string) error {
	return ensureWithClient(github.NewClient(), aliasOrVersion)
}

func ensureWithClient(client *github.Client, aliasOrVersion string) error {
	if aliasOrVersion == "" {
		return fmt.Errorf("version or alias not specified. Usage: vc-env ensure <version-or-alias>")
	}
	return resolveWithClient(client, ResolveOptions{Spec: aliasOrVersion, Install: true, Silent: true})
}
