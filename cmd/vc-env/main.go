// Package main is the entry point for the vc-env CLI tool.
package main

import (
	"fmt"
	"os"

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
		err = commands.Init()

	case "list":
		err = commands.List()

	case "list-remote":
		includePrerelease := false
		for _, arg := range args[1:] {
			if arg == "--prerelease" {
				includePrerelease = true
			}
		}
		err = commands.ListRemote(includePrerelease)

	case "install":
		version := ""
		if len(args) > 1 {
			version = args[1]
		}
		err = commands.Install(version)

	case "uninstall":
		version := ""
		if len(args) > 1 {
			version = args[1]
		}
		err = commands.Uninstall(version)

	case "shell":
		version := ""
		if len(args) > 1 {
			version = args[1]
		}
		err = commands.Shell(version)

	case "local":
		version := ""
		if len(args) > 1 {
			version = args[1]
		}
		err = commands.Local(version)

	case "global":
		version := ""
		if len(args) > 1 {
			version = args[1]
		}
		err = commands.Global(version)

	case "which":
		err = commands.Which()

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", args[0])
		commands.Help()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
