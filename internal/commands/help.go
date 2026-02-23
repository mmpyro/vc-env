// Package commands implements all vc-env CLI commands.
package commands

import "fmt"

// Help prints the help message with all available commands.
func Help() {
	fmt.Println(`Usage: vc-env <command> [arguments]

Commands:
  help          Display this help message and all available commands
  list          List all installed versions of vcluster cli
  list-remote   List all available versions of vcluster cli from GitHub
  init          Initialize vc-env setup
  install       Install a specific version (or latest if not specified)
  uninstall     Uninstall a specific version
  shell         Set or show the shell version of vcluster cli
  local         Set or show the local version of vcluster cli
  global        Set or show the global version of vcluster cli
  latest				Print the latest available version of vcluster cli from GitHub releases.
  which         Print the full path to the active vcluster binary
  upgrade       Upgrade vc-env to the latest version
  version       Print the version of vc-env`)
}
