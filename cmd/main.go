// Copyright Enea Guidi (hmny).

// This package contains the entry point of the whole program, it handles
// directly all the interaction with its utility module with the final pourpose
// of extracting a Choreography Automata from the given input file
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

	// Default level for trace option while parsing the file
	traceOpts := static_analysis.NoTrace
	// If the extended mode is enabled, it overrides the basic mode
	if traceFlag != nil && *traceFlag {
		traceOpts = static_analysis.BasicTrace
	} else if extTraceFlag != nil && *extTraceFlag {
		traceOpts = static_analysis.ExtendedTrace
	}

	// Parses and extracts the metadata from the given file
	fileMetadata := static_analysis.ExtractMetadata(*inputFile, traceOpts)

	// ! Debug, will be removed
	for _, funcMeta := range fileMetadata.FunctionMeta {
		filename := fmt.Sprintf("debug/%s.svg", funcMeta.Name)
		funcMeta.ScopeAutomata.Export(filename, graphviz.SVG)
	}

	// Retrieves the "main" function metadata
	mainMeta, exist := fileMetadata.FunctionMeta["main"]
	if !exist {
		log.Fatal("Cannot extract Partial Automata, 'main' function metadata not found")
	}

	// Extracts the local views starting from the program entrypoint ("main" function)
	localViews := transforms.GetLocalViews(mainMeta, fileMetadata)

	// ! Debug, will be removed
	for _, lView := range localViews {
		// Exports the local view (NFA version)
		filename := fmt.Sprintf("debug/NFA-%s.svg", lView.Name)
		lView.Automata.Export(filename, graphviz.SVG)

		// Determinization of the local view FSA
		lViewDFA := transforms.SubsetConstruction(lView.Automata)

		// Constructs and exports the local view (DFA version)
		tmp := fmt.Sprintf("debug/DFA-%s.svg", lView.Name)
		lViewDFA.Export(tmp, graphviz.SVG)

		// Updates the automata for the local view
		lView.Automata = lViewDFA.Copy()
	}

	// TODO Uses the metadata to generate a Deterministic Choreography Automata (DCA)
	// TODO transforms.GenerateDCA(fileMetadata)

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
