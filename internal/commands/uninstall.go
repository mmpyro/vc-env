package commands

import (
	"fmt"
	"os"

	"github.com/user/vc-env/internal/config"
)

// UninstallHelp prints the help message for the uninstall command.
func UninstallHelp() {
	fmt.Println(`Usage: vc-env uninstall <version>

Remove an installed vcluster version from $VCENV_ROOT/versions.

Flags:
  -h, --help     Show this help message`)
}

// Uninstall removes an installed vcluster version.
func Uninstall(version string) error {
	if err := config.RequireInit(); err != nil {
		return err
	}

	if version == "" {
		return fmt.Errorf("version argument is required. Usage: vc-env uninstall <version>")
	}

	// Check if version is installed
	installed, err := config.IsVersionInstalled(version)
	if err != nil {
		return err
	}
	if !installed {
		return fmt.Errorf("version %s is not installed", version)
	}

	// Remove version directory
	versionDir, err := config.GetVersionDir(version)
	if err != nil {
		return err
	}
	if err := os.RemoveAll(versionDir); err != nil {
		return fmt.Errorf("failed to remove version %s: %w", version, err)
	}

	fmt.Printf("version %s uninstalled\n", version)
	return nil
}
