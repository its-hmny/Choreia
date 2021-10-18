// Copyright Enea Guidi (hmny).

// This package contains the entry point of the whole program, it handles
// directly all the interaction with its utility module with the final pourpose
// of extracting a Choreography Automata from the given input file
package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"strings"

	"github.com/pborman/getopt/v2"

	choreia_parser "github.com/its-hmny/Choreia/internal/parser"
)

// ! Only for debugging pourposes will be removed later
type debugVisitor int

// ! Only for debugging pourposes will be removed later
func (ip debugVisitor) Visit(node ast.Node) ast.Visitor {
	if node == nil {
		return nil
	}
	fmt.Printf("%s %T\n", strings.Repeat("  ", int(ip)), node)
	return ip + 1
}

func main() {
	// Logger setup
	log.SetPrefix("[Choreia]: ")
	log.SetFlags(0)

	// Getopt setup for CLI argument parsing
	inputFile := getopt.StringLong("input", 'i', "", "The .go file from which extract the Choreography Automata")
	traceFlags := getopt.BoolLong("trace", 't', "Pretty prints on the console the AST", "false")
	getopt.BoolLong("help", 'h', "Display this help message", "false")
	// ! Only for debugging pourposes will be removed later
	debugFlags := getopt.BoolLong("debug", 'd', "Pretty prints on the console the AST", "false")

	// Parse the program arguments
	getopt.Parse()

	// Checks for the existence of an input file
	if inputFile == nil && *inputFile == "" {
		getopt.Usage()
		return
	}
	// Checks that the given input path actually exists
	if fStat, err := os.Stat(*inputFile); os.IsNotExist(err) || fStat.IsDir() {
		log.Fatal("A path to an existing go source file is neeeded")
	}

	// Positions are relative to fset
	fset := token.NewFileSet()

	// Parser mode flags, we want all every error possible
	flags := parser.DeclarationErrors | parser.AllErrors
	// And optionally a trace printed on the stdout
	if traceFlags != nil && *traceFlags {
		flags |= parser.Trace
	}

	// Parses the file and retrieves the AST
	f, err := parser.ParseFile(fset, *inputFile, nil, flags)
	if err != nil {
		log.Fatal(err)
		return
	}

	// ! Debug Visitor to print to terminal in a more human readable manner the AST
	if debugFlags != nil && *debugFlags {
		var debug debugVisitor
		ast.Walk(debug, f)
		fmt.Printf("\n------------------------ START DEBUG PRINT------------------------\n")
	}

	// Extracts the metadata about the given Go file and writes it to a JSON metadata file
	fileMetadata := choreia_parser.ParseFile(f)
	fileDump, fileErr := os.Create("file_meta.json")
	jsonDump, jsonErr := json.MarshalIndent(fileMetadata, "", "  ")

	if jsonErr != nil || fileErr != nil {
		log.Fatal("Error encountered while writing JSON metadata file")
	}

	fileDump.WriteString(string(jsonDump))
}
