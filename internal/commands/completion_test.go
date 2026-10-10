package commands

import (
	"strings"
	"testing"
)

func TestCompletion_Bash(t *testing.T) {
	output := captureStdout(t, func() {
		if err := Completion("bash"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	for _, want := range []string{
		"_vc_env_completions()",
		"complete -F _vc_env_completions vc-env",
		"vc-env __complete-versions installed",
		"vc-env __complete-versions remote",
		"install uninstall shell local global latest which exec resolve status doctor prune upgrade version completion",
	} {
		if !strings.Contains(output, want) {
			t.Errorf("bash completion missing %q\n--- script ---\n%s", want, output)
		}
	}
}

func TestCompletion_Zsh(t *testing.T) {
	output := captureStdout(t, func() {
		if err := Completion("zsh"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	for _, want := range []string{
		"#compdef vc-env",
		"_vc_env()",
		"compdef _vc_env vc-env",
		"vc-env __complete-versions installed",
		"vc-env __complete-versions remote",
	} {
		if !strings.Contains(output, want) {
			t.Errorf("zsh completion missing %q\n--- script ---\n%s", want, output)
		}
	}
}

func TestCompletion_Fish(t *testing.T) {
	output := captureStdout(t, func() {
		if err := Completion("fish"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	for _, want := range []string{
		"complete -c vc-env -f",
		"__fish_use_subcommand",
		"__fish_seen_subcommand_from install",
		"__fish_seen_subcommand_from uninstall shell local global exec",
		"vc-env __complete-versions remote",
		"vc-env __complete-versions installed",
	} {
		if !strings.Contains(output, want) {
			t.Errorf("fish completion missing %q\n--- script ---\n%s", want, output)
		}
	}
}

func TestCompletion_PowerShell(t *testing.T) {
	for _, shell := range []string{"powershell", "pwsh"} {
		shell := shell
		t.Run(shell, func(t *testing.T) {
			output := captureStdout(t, func() {
				if err := Completion(shell); err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			})

			for _, want := range []string{
				"Register-ArgumentCompleter",
				"-CommandName vc-env",
				"vc-env __complete-versions remote",
				"vc-env __complete-versions installed",
			} {
				if !strings.Contains(output, want) {
					t.Errorf("powershell completion missing %q\n--- script ---\n%s", want, output)
				}
			}

			// The backtick placeholder must be replaced with a real backtick.
			if strings.Contains(output, "{BT}") {
				t.Errorf("powershell completion script still contains unresolved {BT} placeholder:\n%s", output)
			}
			if !strings.Contains(output, "`r?`n") {
				t.Errorf("powershell completion script missing escaped newline split pattern:\n%s", output)
			}
		})
	}
}

func TestCompletion_UnknownShell(t *testing.T) {
	err := Completion("tcsh")
	if err == nil {
		t.Fatal("expected error for unknown shell")
	}
	msg := err.Error()
	for _, want := range []string{"tcsh", "bash", "zsh", "fish", "powershell"} {
		if !strings.Contains(msg, want) {
			t.Errorf("error message should mention %q, got: %s", want, msg)
		}
	}
}

func TestCompletion_DetectsShellFromEnv(t *testing.T) {
	cases := []struct {
		shellEnv string
		want     string
	}{
		{"/bin/zsh", "#compdef vc-env"},
		{"/usr/bin/fish", "complete -c vc-env -f"},
		{"/usr/local/bin/pwsh", "Register-ArgumentCompleter"},
		{"/bin/bash", "complete -F _vc_env_completions vc-env"},
	}
	for _, tc := range cases {
		t.Run(tc.shellEnv, func(t *testing.T) {
			t.Setenv("SHELL", tc.shellEnv)
			var stderr string
			out := captureStdout(t, func() {
				stderr = captureStderr(t, func() {
					if err := Completion(""); err != nil {
						t.Fatalf("unexpected error: %v", err)
					}
				})
			})
			if !strings.Contains(out, tc.want) {
				t.Errorf("expected %q for SHELL=%s", tc.want, tc.shellEnv)
			}
			if stderr != "" {
				t.Errorf("expected no stderr note on successful detection, got %q", stderr)
			}
		})
	}
}

func TestCompletion_FallsBackToBash(t *testing.T) {
	for _, shellEnv := range []string{"", "/bin/tcsh"} {
		t.Run("SHELL="+shellEnv, func(t *testing.T) {
			t.Setenv("SHELL", shellEnv)
			var stderr string
			out := captureStdout(t, func() {
				stderr = captureStderr(t, func() {
					if err := Completion(""); err != nil {
						t.Fatalf("unexpected error: %v", err)
					}
				})
			})
			if !strings.Contains(out, "complete -F _vc_env_completions vc-env") {
				t.Errorf("expected bash script, got:\n%s", out)
			}
			if !strings.Contains(stderr, "defaulting to bash") {
				t.Errorf("expected stderr note, got %q", stderr)
			}
		})
	}
}

func TestCompletion_CommandLists(t *testing.T) {
	for _, shell := range []string{"bash", "zsh", "fish", "powershell"} {
		t.Run(shell, func(t *testing.T) {
			out := captureStdout(t, func() {
				if err := Completion(shell); err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			})
			for _, want := range []string{"doctor", "prune", "resolve", "completion", "pwsh"} {
				if !strings.Contains(out, want) {
					t.Errorf("%s completion missing %q", shell, want)
				}
			}
			// ensure is internal: its versions still complete, but it is
			// not offered as a command.
			for _, unwanted := range []string{" ensure\"", "'ensure:", "-a 'ensure'", "'resolve','ensure'", "autocompletion"} {
				if strings.Contains(out, unwanted) {
					t.Errorf("%s completion should not list %q", shell, unwanted)
				}
			}
		})
	}
}

func TestCompletionHelp(t *testing.T) {
	output := captureStdout(t, CompletionHelp)
	for _, want := range []string{
		"vc-env completion [<shell>] [--shell <shell>]",
		"pwsh",
		"$SHELL",
		"bash",
		"zsh",
		"fish",
		"powershell",
		"autocompletion",
	} {
		if !strings.Contains(output, want) {
			t.Errorf("completion help missing %q\n--- output ---\n%s", want, output)
		}
	}
}
