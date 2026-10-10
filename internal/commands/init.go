package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/user/vc-env/internal/config"
	"github.com/user/vc-env/internal/shim"
)

// InitHelp prints the help message for the init command.
func InitHelp() {
	fmt.Println(`Usage: vc-env init

Create $VCENV_ROOT/versions and the vcluster shim, then print the shell
initialization code.  Add to your shell rc file:

  bash/zsh:  eval "$(vc-env init)"
  fish:      vc-env init | source

Flags:
  -h, --help     Show this help message`)
}

// Init initializes the vc-env environment.
// If VCENV_ROOT is not set, it prints setup instructions.
// If set, it creates the versions directory, generates the shim, and outputs
// shell initialization code.
func Init() error {
	root, ok := config.GetVCEnvRoot()
	if !ok {
		fmt.Fprintln(os.Stderr, `VCENV_ROOT not set. Set VCENV_ROOT for your shell using the below commands.
echo 'export VCENV_ROOT="$HOME/.vcenv"' >> ~/.bashrc
or
echo 'export VCENV_ROOT="$HOME/.vcenv"' >> ~/.zshrc`)
		return fmt.Errorf("VCENV_ROOT not set")
	}

	// Create versions directory
	versionsDir := filepath.Join(root, "versions")
	if err := os.MkdirAll(versionsDir, 0o755); err != nil {
		return fmt.Errorf("failed to create versions directory: %w", err)
	}

	// Generate shim script
	if err := shim.GenerateShimScript(root); err != nil {
		return fmt.Errorf("failed to generate shim: %w", err)
	}

	// Output shell initialization code
	fmt.Print(shim.GenerateShellInit(root))

	return nil
}
