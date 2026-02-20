package commands

import "fmt"

// Version is set at build time via ldflags.
var Version = "dev"

// PrintVersion prints the vc-env version.
func PrintVersion() {
	fmt.Println(Version)
}
