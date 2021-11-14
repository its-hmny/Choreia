// Copyright Enea Guidi (hmny).

// This package contains the entry point of the whole program, it handles
// directly all the interaction with its utility module with the final pourpose
// of extracting a Choreography Automata from the given input file
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/pborman/getopt/v2"

	// Choreia internal metatdata extraction module
	"github.com/its-hmny/Choreia/internal/meta"
	// Choreia internal module for CDA handling
	"github.com/its-hmny/Choreia/internal/automata"
)

func main() {
	// Getopt setup for CLI argument parsing
	inputFile := getopt.StringLong("input", 'i', "", "The .go file from which extract the Choreography Automata")
	traceFlag := getopt.BoolLong("trace", 't', "Pretty prints on the console the AST", "false")
	extTraceFlag := getopt.BoolLong("ext-trace", 'e', "Pretty prints on the console the expanded AST", "false")
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

	// ! Debug, will be removed
	if _, err := os.Stat("debug"); err == nil {
		os.RemoveAll("debug")
	}
	os.Mkdir("debug", 0775)

	// ! Debug print, will be removed
	fmt.Println("\t\t---------------------- PARSER DEBUG PRINT ----------------------")

	// Default level for trace option while parsing the file
	traceOpts := meta.NoTrace
	// If the extended mode is enabled, it overrides the basic mode
	if traceFlag != nil && *traceFlag {
		traceOpts = meta.BasicTrace
	} else if extTraceFlag != nil && *extTraceFlag {
		traceOpts = meta.ExtendedTrace
	}

	// Parses and esxtracts the metadata from the given file
	fileMetadata := meta.ExtractMetadata(*inputFile, traceOpts)

	// ! From here on is all a work in progress
	fmt.Println("\t\t------------------------- CDA DEBUG PRINT -------------------------")

	// Uses the metadata to generate a Deterministic Choreography Automata (DCA)
	automata.GenerateDCA(fileMetadata)

	// // ! Debugging export as SVG of the graphs
	// for name, meta := range fileMetadata.FunctionMeta {
	// meta.ScopeAutomata.ExportAsSVG(fmt.Sprintf("debug/%s.svg", name))
	// }

	//  //  ! Debugging export of the metadata as SVG
	// fileDump, fileErr := os.Create("debug/file_meta.json")
	// jsonDump, jsonErr := json.MarshalIndent(fileMetadata, "", "  ")
	// if jsonErr != nil || fileErr != nil {
	// log.Fatal("Error encountered while writing JSON metadata file")
	// }
	// fileDump.WriteString(string(jsonDump))
}
