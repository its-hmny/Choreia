// Copyright 2020 Enea Guidi (hmny). All rights reserved.
// This file are distributed under the General Public License v 3.0.
// A copy of abovesaid license can be found in the LICENSE file.

// This is the entry point of the whole program (Choreia).
// It handles directly all the interaction with the respective utilities module in order
// to extract the Choreography Automata from the given input file
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/goccy/go-graphviz"
	"github.com/pborman/getopt/v2"

	// Choreia internal static analysis and metatdata extraction module
	"github.com/its-hmny/Choreia/internal/static_analysis"
	// Choreia internal Choreography Automata transformation module
	"github.com/its-hmny/Choreia/internal/transforms"
)

func main() {
	// Getopt setup for CLI argument parsing
	inputFile := getopt.StringLong("input", 'i', "", "The .go file from which extract the Choreography Automata")
	outputPath := getopt.StringLong("output", 'o', "./choreia.out", "The path to where the extracted data will be saved")
	traceFlag := getopt.BoolLong("trace", 't', "Pretty prints on the console the AST", "false")
	showUsage := getopt.BoolLong("help", 'h', "Display this help message", "false")
	getopt.Parse() // Parses the program arguments

	// Logger setup
	log.SetPrefix("[Choreia] ")
	log.SetFlags(log.Ltime | log.Lshortfile)

	// Checks that the input file is provided via CLI argument
	if *showUsage || inputFile == nil || *inputFile == "" {
		getopt.Usage()
		return
	}

	if _, err := os.Stat(*outputPath); err == nil {
		os.RemoveAll(*outputPath)
	}
	os.Mkdir(*outputPath, 0775)

	// Default level for trace option while parsing the file
	traceOpts := static_analysis.NoTrace
	// If the extended mode is enabled, it overrides the basic mode
	if traceFlag != nil && *traceFlag {
		traceOpts = static_analysis.Trace
	}

	// Parses and extracts the metadata from the given file
	fileMetadata := static_analysis.ExtractMetadata(*inputFile, traceOpts)

	// Retrieves the "main" function metadata (the entrypoint of the "root" Goroutine)
	mainMeta, exist := fileMetadata.FunctionMeta["main"]
	if !exist {
		log.Fatal("Cannot extract Partial Automata, 'main' function metadata not found")
	}

	// Extracts the local views starting from the program entrypoint ("main" function)
	localViews := transforms.GetLocalViews(mainMeta, fileMetadata)

	// For each local view of the Choreography Automata applies transformations (determinization, minimization)
	for _, lView := range localViews {
		// Exports the local view (NFA version)
		filenameNFA := fmt.Sprintf("%s/NFA %s.svg", *outputPath, lView.Name)
		lView.Automata.Export(filenameNFA, graphviz.SVG)

		// Determinization of the local view FSA
		lViewDFA := transforms.SubsetConstruction(lView.Automata)
		// TODO: Add minimization of the DFA

		// Constructs and exports the local view (DFA version)
		filenameDFA := fmt.Sprintf("%s/DFA %s.svg", *outputPath, lView.Name)
		lViewDFA.Export(filenameDFA, graphviz.SVG)

		// Updates the automata for the local view
		lView.Automata = lViewDFA.Copy()
	}

	// At last extracts the Choreography Automata (also known as "global view")
	finalCA := transforms.GenerateDCA(localViews)
	finalCA.Export(fmt.Sprintf("%s/Choreography Automata.svg", *outputPath), graphviz.SVG)
}
