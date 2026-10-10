// Package main is the entry point for the vc-env CLI tool.
package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/user/vc-env/internal/commands"
)

// Version is set at build time via ldflags:
//
//	go build -ldflags "-X main.Version=0.1.0"
var Version = "dev"

func main() {
	// Inject version into commands package
	commands.Version = Version

	args := os.Args[1:]

	if len(args) == 0 {
		commands.Help()
		os.Exit(0)
	}

	var err error

	switch args[0] {
	case "help", "--help", "-h":
		commands.Help()

	case "version", "--version", "-v":
		commands.PrintVersion()

	case "init":
		exitOnHelp(args[1:], commands.InitHelp)
		err = commands.Init()

	case "list":
		exitOnHelp(args[1:], commands.ListHelp)
		err = commands.List()

	case "list-remote":
		includePrerelease := false
		for _, arg := range args[1:] {
			switch arg {
			case "-h", "--help":
				commands.ListRemoteHelp()
				os.Exit(0)
			case "--prerelease":
				includePrerelease = true
			}
		}
		err = commands.ListRemote(includePrerelease)

	case "latest":
		includePrerelease := false
		for _, arg := range args[1:] {
			switch arg {
			case "-h", "--help":
				commands.LatestHelp()
				os.Exit(0)
			case "--prerelease":
				includePrerelease = true
			}
		}
		err = commands.Latest(includePrerelease)

	case "install":
		opts := commands.InstallOptions{}
		rest := args[1:]
		for i := 0; i < len(rest); i++ {
			arg := rest[i]
			switch {
			case arg == "-s" || arg == "--silent":
				opts.Silent = true
			case arg == "-h" || arg == "--help":
				commands.InstallHelp()
				os.Exit(0)
			case arg == "--from-file":
				if i+1 >= len(rest) {
					fmt.Fprintln(os.Stderr, "--from-file requires a path argument")
					os.Exit(1)
				}
				i++
				opts.FromFile = rest[i]
			case strings.HasPrefix(arg, "--from-file="):
				opts.FromFile = strings.TrimPrefix(arg, "--from-file=")
			case arg == "--sha256":
				if i+1 >= len(rest) {
					fmt.Fprintln(os.Stderr, "--sha256 requires a hex digest argument")
					os.Exit(1)
				}
				i++
				opts.SHA256 = rest[i]
			case strings.HasPrefix(arg, "--sha256="):
				opts.SHA256 = strings.TrimPrefix(arg, "--sha256=")
			case opts.Version == "" && !strings.HasPrefix(arg, "-"):
				opts.Version = arg
			}
		}
		err = commands.InstallWith(opts)

	case "uninstall":
		exitOnHelp(args[1:], commands.UninstallHelp)
		err = commands.Uninstall(firstPositional(args[1:]))

	case "shell":
		exitOnHelp(args[1:], commands.ShellHelp)
		err = commands.Shell(firstPositional(args[1:]))

	case "local":
		exitOnHelp(args[1:], commands.LocalHelp)
		err = commands.Local(firstPositional(args[1:]))

	case "global":
		exitOnHelp(args[1:], commands.GlobalHelp)
		err = commands.Global(firstPositional(args[1:]))

	case "which":
		exitOnHelp(args[1:], commands.WhichHelp)
		err = commands.Which()

	case "upgrade":
		// -h must never fall through to an actual upgrade.
		exitOnHelp(args[1:], commands.UpgradeHelp)
		err = commands.Upgrade()

	case "completion", "autocompletion":
		if args[0] == "autocompletion" {
			fmt.Fprintln(os.Stderr, "vc-env: 'autocompletion' is deprecated; use 'vc-env completion'")
		}
		shell := ""
		rest := args[1:]
		for i := 0; i < len(rest); i++ {
			arg := rest[i]
			switch {
			case arg == "-h" || arg == "--help":
				commands.CompletionHelp()
				os.Exit(0)
			case arg == "--shell":
				if i+1 >= len(rest) {
					fmt.Fprintln(os.Stderr, "--shell requires a shell name")
					os.Exit(1)
				}
				i++
				shell = rest[i]
			case strings.HasPrefix(arg, "--shell="):
				shell = strings.TrimPrefix(arg, "--shell=")
			case shell == "" && !strings.HasPrefix(arg, "-"):
				shell = arg
			}
		}
		err = commands.Completion(shell)

	case "__complete-versions":
		// Internal: used by generated completion scripts. Lists installed or
		// cached remote versions, one per line, without touching the network.
		err = commands.CompleteVersions(args[1:])

	case "status":
		exitOnHelp(args[1:], commands.StatusHelp)
		err = commands.Status()

	case "doctor":
		fix, deep := false, false
		for _, arg := range args[1:] {
			switch arg {
			case "-h", "--help":
				commands.DoctorHelp()
				os.Exit(0)
			case "--fix":
				fix = true
			case "--deep":
				deep = true
			default:
				fmt.Fprintf(os.Stderr, "Unknown flag for doctor: %s\n\n", arg)
				commands.DoctorHelp()
				os.Exit(1)
			}
		}
		err = commands.Doctor(fix, deep)

	case "prune":
		err = commands.Prune(parsePruneArgs(args[1:]))

	case "exec":
		opts := commands.ExecOptions{}
		rest := args[1:]
		// Flags are only recognised before the version; everything after it
		// belongs to vcluster.
		for i := 0; i < len(rest); i++ {
			arg := rest[i]
			if opts.Spec != "" {
				opts.Args = rest[i:]
				break
			}
			switch arg {
			case "-h", "--help":
				commands.ExecHelp()
				os.Exit(0)
			case "--auto":
				opts.Auto = true
			case "--no-auto":
				opts.NoAuto = true
			default:
				opts.Spec = arg
			}
		}
		err = commands.Exec(opts)

	case "resolve":
		opts := commands.ResolveOptions{}
		for _, arg := range args[1:] {
			switch {
			case arg == "-h" || arg == "--help":
				commands.ResolveHelp()
				os.Exit(0)
			case arg == "--install":
				opts.Install = true
			case arg == "-s" || arg == "--silent":
				opts.Silent = true
			case opts.Spec == "" && !strings.HasPrefix(arg, "-"):
				opts.Spec = arg
			}
		}
		err = commands.Resolve(opts)

	case "ensure":
		// Advanced / internal: equivalent to 'resolve <spec> --install --silent'.
		exitOnHelp(args[1:], commands.EnsureHelp)
		err = commands.Ensure(firstPositional(args[1:]))

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", args[0])
		commands.Help()
		os.Exit(1)
	}

	if err != nil {
		// A child started by 'exec' already reported its own failure; pass
		// its exit code through untouched.
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() >= 0 {
			os.Exit(exitErr.ExitCode())
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// exitOnHelp prints help and exits 0 when args contain -h or --help.
func exitOnHelp(args []string, help func()) {
	for _, arg := range args {
		if arg == "-h" || arg == "--help" {
			help()
			os.Exit(0)
		}
	}
}

// firstPositional returns the first argument that is not a flag, or "".
func firstPositional(args []string) string {
	for _, arg := range args {
		if !strings.HasPrefix(arg, "-") {
			return arg
		}
	}
	return ""
}

// parsePruneArgs parses the prune flags, exiting on invalid input.
func parsePruneArgs(rest []string) commands.PruneOptions {
	opts := commands.PruneOptions{}
	keepLast := func(val string) {
		n, parseErr := strconv.Atoi(val)
		if parseErr != nil || n < 0 {
			fmt.Fprintf(os.Stderr, "invalid --keep-last value %q: must be a non-negative integer\n", val)
			os.Exit(1)
		}
		opts.KeepLast = n
	}
	olderThan := func(val string) {
		d, _, parseErr := commands.ParseDurationShorthand(val)
		if parseErr != nil {
			fmt.Fprintf(os.Stderr, "invalid --older-than value %q: %v\n", val, parseErr)
			os.Exit(1)
		}
		opts.OlderThan = d
		opts.OlderThanProvided = true
	}
	for i := 0; i < len(rest); i++ {
		arg := rest[i]
		switch {
		case arg == "-h" || arg == "--help":
			commands.PruneHelp()
			os.Exit(0)
		case arg == "--dry-run":
			opts.DryRun = true
		case arg == "--yes":
			opts.Yes = true
		case arg == "--keep-last" || arg == "--older-than":
			if i+1 >= len(rest) {
				fmt.Fprintf(os.Stderr, "%s requires a value\n", arg)
				os.Exit(1)
			}
			i++
			if arg == "--keep-last" {
				keepLast(rest[i])
			} else {
				olderThan(rest[i])
			}
		case strings.HasPrefix(arg, "--keep-last="):
			keepLast(strings.TrimPrefix(arg, "--keep-last="))
		case strings.HasPrefix(arg, "--older-than="):
			olderThan(strings.TrimPrefix(arg, "--older-than="))
		default:
			fmt.Fprintf(os.Stderr, "Unknown flag for prune: %s\n\n", arg)
			commands.PruneHelp()
			os.Exit(1)
		}
	}
	return opts
}
