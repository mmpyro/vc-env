package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Completion prints a shell completion script for the requested shell.
// Supported shells: bash, zsh, fish, powershell (alias: pwsh).  When shell is
// empty it is detected from $SHELL, falling back to bash with a note on
// stderr.
func Completion(shell string) error {
	if shell == "" {
		shell = detectShell()
	}
	switch shell {
	case "bash":
		fmt.Print(bashCompletionScript)
	case "zsh":
		fmt.Print(zshCompletionScript)
	case "fish":
		fmt.Print(fishCompletionScript)
	case "powershell", "pwsh":
		fmt.Print(powershellCompletionScript)
	default:
		return fmt.Errorf("completion: unsupported shell %q.\n"+
			"Supported shells: bash, zsh, fish, powershell (pwsh)", shell)
	}
	return nil
}

// detectShell maps the basename of $SHELL to a supported shell name.  It
// returns "bash" (and prints a note to stderr) when $SHELL is unset or names
// an unsupported shell.
func detectShell() string {
	name := filepath.Base(os.Getenv("SHELL"))
	name = strings.TrimSuffix(name, ".exe")
	switch name {
	case "bash", "zsh", "fish", "powershell", "pwsh":
		return name
	}
	fmt.Fprintln(os.Stderr, "vc-env: could not detect a supported shell from $SHELL; defaulting to bash")
	return "bash"
}

// CompletionHelp prints help for the completion command.
func CompletionHelp() {
	fmt.Fprintln(os.Stdout, `Usage: vc-env completion [<shell>] [--shell <shell>]

Generate a shell completion script for vc-env.

Supported shells: bash, zsh, fish, powershell (alias: pwsh)

When no shell is given, it is detected from $SHELL; if that fails, bash is
used and a note is printed to stderr.

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

Note: 'vc-env autocompletion' is a deprecated alias for 'vc-env completion'.`)
}

// bashCompletionScript is a native bash completion. It delegates version
// lookup to the internal `vc-env __complete-versions` helper so the list stays
// authoritative in Go.
const bashCompletionScript = `_vc_env_completions() {
    local cur prev opts
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"
    opts="help list list-remote init install uninstall shell local global latest which exec resolve status doctor prune upgrade version completion"

    if [[ ${COMP_CWORD} -eq 1 ]] ; then
        COMPREPLY=( $(compgen -W "${opts}" -- "${cur}") )
        return 0
    fi

    case "${prev}" in
        install|ensure|resolve)
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
        completion|--shell)
            COMPREPLY=( $(compgen -W "bash zsh fish powershell pwsh --shell" -- "${cur}") )
            return 0
            ;;
    esac

    case "${COMP_WORDS[1]}" in
        doctor)
            COMPREPLY=( $(compgen -W "--fix --deep --help" -- "${cur}") )
            ;;
        prune)
            COMPREPLY=( $(compgen -W "--keep-last --older-than --dry-run --yes --help" -- "${cur}") )
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
        'resolve:Resolve a version or alias to a concrete installed version'
        'status:Show current vc-env environment status'
        'doctor:Diagnose the vc-env environment'
        'prune:Remove unreferenced installed versions'
        'upgrade:Upgrade vc-env to the latest version'
        'version:Print the version of vc-env'
        'completion:Generate shell completion script'
    )
    shells=(bash zsh fish powershell pwsh)

    if (( CURRENT == 2 )); then
        _describe -t commands 'vc-env command' subs
        return
    fi

    case ${words[2]} in
        install|ensure|resolve)
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
            compadd -- $shells --shell
            ;;
        doctor)
            compadd -- --fix --deep --help
            ;;
        prune)
            compadd -- --keep-last --older-than --dry-run --yes --help
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
complete -c vc-env -n '__fish_use_subcommand' -a 'resolve'        -d 'Resolve a version or alias to a concrete installed version'
complete -c vc-env -n '__fish_use_subcommand' -a 'status'         -d 'Show current vc-env environment status'
complete -c vc-env -n '__fish_use_subcommand' -a 'doctor'         -d 'Diagnose the vc-env environment'
complete -c vc-env -n '__fish_use_subcommand' -a 'prune'          -d 'Remove unreferenced installed versions'
complete -c vc-env -n '__fish_use_subcommand' -a 'upgrade'        -d 'Upgrade vc-env to the latest version'
complete -c vc-env -n '__fish_use_subcommand' -a 'version'        -d 'Print the vc-env version'
complete -c vc-env -n '__fish_use_subcommand' -a 'completion'     -d 'Generate shell completion script'

# Version arguments
complete -c vc-env -n '__fish_seen_subcommand_from install ensure resolve' \
    -a '(vc-env __complete-versions remote 2>/dev/null)'
complete -c vc-env -n '__fish_seen_subcommand_from uninstall shell local global exec' \
    -a '(vc-env __complete-versions installed 2>/dev/null)'

# completion <shell>
complete -c vc-env -n '__fish_seen_subcommand_from completion' \
    -a 'bash zsh fish powershell pwsh --shell'

# doctor / prune flags
complete -c vc-env -n '__fish_seen_subcommand_from doctor' -a '--fix --deep --help'
complete -c vc-env -n '__fish_seen_subcommand_from prune' \
    -a '--keep-last --older-than --dry-run --yes --help'
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
        'latest','which','exec','resolve','status','doctor','prune','upgrade','version','completion'
    )
    $shells = @('bash','zsh','fish','powershell','pwsh')

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
        { @('install','ensure','resolve') -contains $_ } {
            $candidates = (vc-env __complete-versions remote 2>$null) -split "{BT}r?{BT}n" | Where-Object { $_ }
        }
        { @('uninstall','shell','local','global','exec') -contains $_ } {
            $candidates = (vc-env __complete-versions installed 2>$null) -split "{BT}r?{BT}n" | Where-Object { $_ }
        }
        'completion' { $candidates = $shells + @('--shell') }
        'doctor'     { $candidates = @('--fix','--deep','--help') }
        'prune'      { $candidates = @('--keep-last','--older-than','--dry-run','--yes','--help') }
        default      { $candidates = @() }
    }

    $candidates |
        Where-Object { $_ -like "$wordToComplete*" } |
        ForEach-Object { [System.Management.Automation.CompletionResult]::new($_, $_, 'ParameterValue', $_) }
}
`
