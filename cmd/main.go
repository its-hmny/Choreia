// Copyright 2020 Enea Guidi (hmny). All rights reserved.
// This files are distributed under the General Public License v3.0.
// A copy of abovesaid license can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"

	"github.com/teris-io/cli"
)

// Usage message to print when the user uses the --help flag
const Usage = `Associates a Choreography Automata to your Go program.`

// "metadata" subcommand, extracts the metadata from the given program through static analysis
var MetadataCmd = cli.
	NewCommand("metadata", "Extracts metadata through static analysis").WithShortcut("meta").
	// Arguments, options and flags registrations
	WithArg(cli.NewArg("input", "The program entrypoint (main.go)").WithType(cli.TypeString)).
	WithOption(cli.NewOption("output", "Output file path").WithChar('o').WithType(cli.TypeString)).
	WithOption(cli.NewOption("verbose", "Verbose logging").WithChar('v').WithType(cli.TypeBool)).
	// Registers an handler function that will dispatche the argument to the respective module
	WithAction(func(args []string, options map[string]string) int { fmt.Println(args, options); return 0 })

func main() {
	// Builds the CLI app with the respective subcommands
	app := cli.New(Usage).WithCommand(MetadataCmd)
	// Dispatch the arguments and executes the respective action
	os.Exit(app.Run(os.Args, os.Stdout))
}
