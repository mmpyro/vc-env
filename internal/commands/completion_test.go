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
		"install uninstall shell local global latest which exec status upgrade version completion",
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

func TestCompletion_MissingShell(t *testing.T) {
	err := Completion("")
	if err == nil {
		t.Fatal("expected error when shell argument is empty")
	}
	if !strings.Contains(err.Error(), "missing shell") {
		t.Errorf("expected 'missing shell' message, got: %s", err.Error())
	}
}

func TestCompletionHelp(t *testing.T) {
	output := captureStdout(t, CompletionHelp)
	for _, want := range []string{
		"vc-env completion <shell>",
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
