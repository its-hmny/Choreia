// Copyright 2020 Enea Guidi (hmny). All rights reserved.
// This files are distributed under the General Public License v3.0.
// A copy of aforesaid license can be found in the LICENSE file.

package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/teris-io/cli"

	analysis "github.com/choreia/go/static_analysis"
)

var ExtractMeta = cli.
	// "metadata" subcommand, extracts the metadata from the given program through static analysis
	NewCommand("extract-meta", "Extracts metadata through static analysis").WithShortcut("meta").

	// Arguments, options and flags registrations
	WithArg(cli.NewArg("input", "The program entrypoint (main.go)").WithType(cli.TypeString)).
	WithOption(cli.NewOption("output", "Output file path").WithChar('o').WithType(cli.TypeString)).

	// Registers an handler function that will dispatch the argument to the respective module
	WithAction(handlerExtractMeta)

// Thin wrapper function around 'analysis.ExtractAndSave()' it just maps the CLI
// arguments provided by the 'tetris-io/cli' package to the ones required by the
// 'static_analysis' package implemented by Choreia.
func handlerExtractMeta(args []string, options map[string]string) int {
	// Destructure the fields from the respective origins
	input, output := args[0], options["output"]

	// If no output path has been provided we'll just create 'choreia-out.json' in the 'cwd'
	if output == "" {
		log.Trace("No output path has been provided, using: './choreia-out.json'")
		output = "./choreia-out.json"
	}

	log.Infof("Writing extracted metadata to '%s'...", output)

	if err := analysis.ExtractAndSave(input, output); err != nil {
		log.Fatalf("Error during metadata extraction: %s", err)
	}

	return 0
}
