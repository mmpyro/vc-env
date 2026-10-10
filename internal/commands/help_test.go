package commands

import (
	"strings"
	"testing"
)

func TestHelp(t *testing.T) {
	out := captureStdout(t, Help)

	for _, want := range []string{
		"  doctor ", "  prune ", "  resolve ", "  exec ", "  completion ", "  upgrade ",
		"VCENV_ROOT", "VCENV_VERSION", "VCENV_AUTO_INSTALL", "VCENV_GITHUB_TOKEN",
		"GITHUB_TOKEN", "VCENV_GITHUB_API_URL", "VCENV_DOWNLOAD_MIRROR",
		"VCENV_CACHE_TTL", "VCENV_PRUNE_SCAN_ROOTS",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("help missing %q", want)
		}
	}
	// ensure and autocompletion are hidden.
	for _, unwanted := range []string{"ensure", "autocompletion"} {
		if strings.Contains(out, unwanted) {
			t.Errorf("help should not mention %q", unwanted)
		}
	}
}

func TestCommandHelps(t *testing.T) {
	helps := map[string]func(){
		"vc-env init":      InitHelp,
		"vc-env list":      ListHelp,
		"vc-env uninstall": UninstallHelp,
		"vc-env shell":     ShellHelp,
		"vc-env local":     LocalHelp,
		"vc-env global":    GlobalHelp,
		"vc-env which":     WhichHelp,
		"vc-env status":    StatusHelp,
		"vc-env upgrade":   UpgradeHelp,
		"vc-env exec":      ExecHelp,
		"vc-env resolve":   ResolveHelp,
		"vc-env ensure":    EnsureHelp,
		"vc-env doctor":    DoctorHelp,
		"vc-env prune":     PruneHelp,
	}
	for usage, fn := range helps {
		t.Run(usage, func(t *testing.T) {
			out := captureStdout(t, fn)
			if !strings.HasPrefix(out, "Usage: "+usage) {
				t.Errorf("expected help to start with %q, got:\n%s", "Usage: "+usage, out)
			}
		})
	}
}

func TestUpgradeHelp_ListsEnvVars(t *testing.T) {
	out := captureStdout(t, UpgradeHelp)
	for _, want := range []string{"latest", "VCENV_GITHUB_TOKEN", "GITHUB_TOKEN", "VCENV_GITHUB_API_URL", "VCENV_DOWNLOAD_MIRROR"} {
		if !strings.Contains(out, want) {
			t.Errorf("upgrade help missing %q", want)
		}
	}
}
