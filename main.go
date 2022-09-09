// Copyright 2020 Enea Guidi (hmny). All rights reserved.
// This files are distributed under the General Public License v3.0.
// A copy of abovesaid license can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"
	"path/filepath"

	analysis "github.com/its-hmny/Choreia/go/static_analysis"
	log "github.com/sirupsen/logrus"
	"github.com/teris-io/cli"
)

// Usage message to print when the user uses the --help flag
// TODO Further elaborate and document
const Usage = `Associates a Choreography Automata to your Go program.`

// "metadata" subcommand, extracts the metadata from the given program through static analysis
var metadataCmd = cli.
	NewCommand("metadata", "Extracts metadata through static analysis").WithShortcut("meta").
	// Arguments, options and flags registrations
	WithArg(cli.NewArg("input", "The program entrypoint (main.go)").WithType(cli.TypeString)).
	WithOption(cli.NewOption("output", "Output file path").WithChar('o').WithType(cli.TypeString)).
	// Registers an handler function that will dispatche the argument to the respective module
	WithAction(func(args []string, options map[string]string) int {
		// Destructures the fields from the respective origins
		input, output := args[0], options["output"]

		// If no output path is not provided then the file is saved in the same
		// directory as the input with just a different extension (in this case .json)
		if output == "" {
			// Error is ignored since it will be catched later on
			abspath, _ := filepath.Abs(input)
			// If no output path is provided then a JSON file with the name of the given
			// input directory will be saved in the current working directory
			basename := filepath.Base(abspath)
			output = fmt.Sprintf("%s.json", basename)
		}

		if _, err := analysis.ExtractAndSave(input, output); err != nil {
			log.Fatal(err)
		}
		return 0
	})

func init() {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.TextFormatter{ForceColors: true, FullTimestamp: true, TimestampFormat: "15:04:05"})
	// Output to stdout instead of the default stderr
	log.SetOutput(os.Stdout)
	// Only log the warning severity or above.
	log.SetLevel(log.TraceLevel)
}

func main() {
	// Builds the CLI app with the respective subcommands
	app := cli.New(Usage).WithCommand(metadataCmd)
	// Dispatch the arguments and executes the respective action
	os.Exit(app.Run(os.Args, os.Stdout))
}
