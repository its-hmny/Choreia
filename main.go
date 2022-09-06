// Copyright 2020 Enea Guidi (hmny). All rights reserved.
// This files are distributed under the General Public License v3.0.
// A copy of abovesaid license can be found in the LICENSE file.

package main

import (
	"log"
	"os"
	"strings"

	"github.com/its-hmny/Choreia/metadata"
	"github.com/teris-io/cli"
)

// Usage message to print when the user uses the --help flag
// TODO Further elaborate and document
const Usage = `Associates a Choreography Automata to your Go program.`

// "metadata" subcommand, extracts the metadata from the given program through static analysis
var MetadataCmd = cli.
	NewCommand("metadata", "Extracts metadata through static analysis").WithShortcut("meta").
	// Arguments, options and flags registrations
	WithArg(cli.NewArg("input", "The program entrypoint (main.go)").WithType(cli.TypeString)).
	WithOption(cli.NewOption("output", "Output file path").WithChar('o').WithType(cli.TypeString)).
	WithOption(cli.NewOption("verbose", "Verbose logging").WithChar('v').WithType(cli.TypeBool)).
	// Registers an handler function that will dispatche the argument to the respective module
	WithAction(MetadataHandler)

// Parses and validates arguments coming from the CLI, eventually transforming them or
// replacing them with default values before calling the wrapped library function.
func MetadataHandler(args []string, options map[string]string) int {
	// Destructures the fields from the respective origins
	input, output, verbose := args[0], options["output"], options["verbose"] == "true"

	if len(args) == 0 {
		log.Fatal("You must provide a .go input file to be analyzed")
		return 1
	}
	if _, err := os.Stat(input); err != nil {
		log.Fatal("The path provided doesn't exist, please check it")
		return 1
	}
	if filestat, _ := os.Stat(input); filestat.IsDir() {
		log.Fatal("Argument must be a file path, directory are not supported yet")
		return 1
	}

	// If no output path is not provided then the file is saved in the same
	// directory as the input with just a different extension (in this case .json)
	if output == "" {
		output = strings.Replace(input, ".go", ".json", -1)
	}

	if _, err := metadata.ExtractAndSave(input, output, verbose); err != nil {
		log.Fatal("Argument must be a file path, directory are not supported yet")
		return 1
	}

	return 0
}

func main() {
	// Builds the CLI app with the respective subcommands
	app := cli.New(Usage).WithCommand(MetadataCmd)
	// Dispatch the arguments and executes the respective action
	os.Exit(app.Run(os.Args, os.Stdout))
}
