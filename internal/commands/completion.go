package commands

import (
	"fmt"
	"os"
	"strings"
)

// Completion prints a shell completion script for the requested shell.
// Supported shells: bash, zsh, fish, powershell (alias: pwsh).
func Completion(shell string) error {
	switch shell {
	case "bash":
		fmt.Print(bashCompletionScript)
	case "zsh":
		fmt.Print(zshCompletionScript)
	case "fish":
		fmt.Print(fishCompletionScript)
	case "powershell", "pwsh":
		fmt.Print(powershellCompletionScript)
	case "":
		return fmt.Errorf("completion: missing shell argument.\n" +
			"Usage: vc-env completion <bash|zsh|fish|powershell>")
	default:
		return fmt.Errorf("completion: unsupported shell %q.\n"+
			"Supported shells: bash, zsh, fish, powershell", shell)
	}
	return nil
}

// CompletionHelp prints help for the completion command.
func CompletionHelp() {
	fmt.Fprintln(os.Stdout, `Usage: vc-env completion <shell>

Generate a shell completion script for vc-env.

Supported shells: bash, zsh, fish, powershell

Install (bash):
  source <(vc-env completion bash)

Install (zsh):
  # Ensure the completion system is loaded once in ~/.zshrc:
  #   autoload -Uz compinit && compinit
  source <(vc-env completion zsh)

Install (fish):
  vc-env completion fish | source
  # Or persistently:
  vc-env completion fish > ~/.config/fish/completions/vc-env.fish

Install (PowerShell):
  vc-env completion powershell | Out-String | Invoke-Expression

Note: 'vc-env autocompletion' is a deprecated alias for 'vc-env completion bash'.`)
}

// bashCompletionScript is a native bash completion. It delegates version
// lookup to the internal `vc-env __complete-versions` helper so the list stays
// authoritative in Go.
const bashCompletionScript = `_vc_env_completions() {
    local cur prev opts
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"
    opts="help list list-remote init install uninstall shell local global latest which exec status upgrade version completion"

    if [[ ${COMP_CWORD} -eq 1 ]] ; then
        COMPREPLY=( $(compgen -W "${opts}" -- "${cur}") )
        return 0
    fi

    case "${prev}" in
        install)
            local versions
            versions=$(vc-env __complete-versions remote 2>/dev/null)
            COMPREPLY=( $(compgen -W "${versions}" -- "${cur}") )
            return 0
            ;;
        uninstall|shell|local|global|exec)
            local versions
            versions=$(vc-env __complete-versions installed 2>/dev/null)
            COMPREPLY=( $(compgen -W "${versions}" -- "${cur}") )
            return 0
            ;;
        completion)
            COMPREPLY=( $(compgen -W "bash zsh fish powershell" -- "${cur}") )
            return 0
            ;;
    esac
}
complete -F _vc_env_completions vc-env
`

// zshCompletionScript is a native zsh completion using the compsys framework.
// Users must have `autoload -Uz compinit && compinit` sourced beforehand.
const zshCompletionScript = `#compdef vc-env
_vc_env() {
    local -a subs shells
    subs=(
        'help:Display help message'
        'list:List installed vcluster versions'
        'list-remote:List available vcluster versions from GitHub'
        'init:Initialize vc-env setup'
        'install:Install a specific version (or latest)'
        'uninstall:Uninstall a specific version'
        'shell:Set or show the shell version'
        'local:Set or show the local version'
        'global:Set or show the global version'
        'latest:Print the latest available vcluster version'
        'which:Print the path to the active vcluster binary'
        'exec:Run a command using a specific vcluster version'
        'status:Show current vc-env environment status'
        'upgrade:Upgrade vc-env to the latest version'
        'version:Print the version of vc-env'
        'completion:Generate shell completion script'
    )
    shells=(bash zsh fish powershell)

    if (( CURRENT == 2 )); then
        _describe -t commands 'vc-env command' subs
        return
    fi

    case ${words[2]} in
        install)
            local -a versions
            versions=(${(f)"$(vc-env __complete-versions remote 2>/dev/null)"})
            compadd -- $versions
            ;;
        uninstall|shell|local|global|exec)
            local -a versions
            versions=(${(f)"$(vc-env __complete-versions installed 2>/dev/null)"})
            compadd -- $versions
            ;;
        completion)
            compadd -- $shells
            ;;
    esac
}
compdef _vc_env vc-env
`

// fishCompletionScript is a native fish completion. Fish filters by prefix
// automatically, so we emit the full list from the helper on every tab.
const fishCompletionScript = `# vc-env fish completion
complete -c vc-env -f

# Top-level subcommands
complete -c vc-env -n '__fish_use_subcommand' -a 'help'           -d 'Display help message'
complete -c vc-env -n '__fish_use_subcommand' -a 'list'           -d 'List installed vcluster versions'
complete -c vc-env -n '__fish_use_subcommand' -a 'list-remote'    -d 'List available vcluster versions from GitHub'
complete -c vc-env -n '__fish_use_subcommand' -a 'init'           -d 'Initialize vc-env setup'
complete -c vc-env -n '__fish_use_subcommand' -a 'install'        -d 'Install a specific version'
complete -c vc-env -n '__fish_use_subcommand' -a 'uninstall'      -d 'Uninstall a specific version'
complete -c vc-env -n '__fish_use_subcommand' -a 'shell'          -d 'Set or show the shell version'
complete -c vc-env -n '__fish_use_subcommand' -a 'local'          -d 'Set or show the local version'
complete -c vc-env -n '__fish_use_subcommand' -a 'global'         -d 'Set or show the global version'
complete -c vc-env -n '__fish_use_subcommand' -a 'latest'         -d 'Print the latest available vcluster version'
complete -c vc-env -n '__fish_use_subcommand' -a 'which'          -d 'Print the active vcluster binary path'
complete -c vc-env -n '__fish_use_subcommand' -a 'exec'           -d 'Run a command using a specific vcluster version'
complete -c vc-env -n '__fish_use_subcommand' -a 'status'         -d 'Show current vc-env environment status'
complete -c vc-env -n '__fish_use_subcommand' -a 'upgrade'        -d 'Upgrade vc-env to the latest version'
complete -c vc-env -n '__fish_use_subcommand' -a 'version'        -d 'Print the vc-env version'
complete -c vc-env -n '__fish_use_subcommand' -a 'completion'     -d 'Generate shell completion script'

# Version arguments
complete -c vc-env -n '__fish_seen_subcommand_from install' \
    -a '(vc-env __complete-versions remote 2>/dev/null)'
complete -c vc-env -n '__fish_seen_subcommand_from uninstall shell local global exec' \
    -a '(vc-env __complete-versions installed 2>/dev/null)'

# completion <shell>
complete -c vc-env -n '__fish_seen_subcommand_from completion' \
    -a 'bash zsh fish powershell'
`

// powershellCompletionScript registers a native PowerShell argument completer.
// Works with both Windows PowerShell 5.1 and PowerShell 7+ (pwsh).
//
// Go raw-string literals cannot contain backticks (`), and PowerShell uses the
// backtick as its escape character, so we use a placeholder (BT) and swap it
// at package init.
var powershellCompletionScript = strings.ReplaceAll(powershellCompletionScriptTemplate, "{BT}", "`")

const powershellCompletionScriptTemplate = `# vc-env PowerShell completion
Register-ArgumentCompleter -Native -CommandName vc-env -ScriptBlock {
    param($wordToComplete, $commandAst, $cursorPosition)

    $subcommands = @(
        'help','list','list-remote','init','install','uninstall','shell','local','global',
        'latest','which','exec','status','upgrade','version','completion'
    )
    $shells = @('bash','zsh','fish','powershell')

    # Tokenise the command line, dropping the program name itself.
    $tokens = @($commandAst.CommandElements | ForEach-Object { $_.ToString() })
    if ($tokens.Count -gt 0) { $tokens = $tokens[1..($tokens.Count - 1)] }

    # Decide whether the user is completing the subcommand slot or an argument.
    $completingSubcommand = $tokens.Count -eq 0 -or
        ($tokens.Count -eq 1 -and $wordToComplete -eq $tokens[0])

    if ($completingSubcommand) {
        $subcommands |
            Where-Object { $_ -like "$wordToComplete*" } |
            ForEach-Object { [System.Management.Automation.CompletionResult]::new($_, $_, 'ParameterValue', $_) }
        return
    }

    $sub = $tokens[0]
    $candidates = @()
    switch ($sub) {
        'install' {
            $candidates = (vc-env __complete-versions remote 2>$null) -split "{BT}r?{BT}n" | Where-Object { $_ }
        }
        { @('uninstall','shell','local','global','exec') -contains $_ } {
            $candidates = (vc-env __complete-versions installed 2>$null) -split "{BT}r?{BT}n" | Where-Object { $_ }
        }
        'completion' { $candidates = $shells }
        default      { $candidates = @() }
    }

    $candidates |
        Where-Object { $_ -like "$wordToComplete*" } |
        ForEach-Object { [System.Management.Automation.CompletionResult]::new($_, $_, 'ParameterValue', $_) }
}
`
