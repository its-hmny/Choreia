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

	// Command line argument checking
	//if len(os.Args) < 2 {
	//	log.Fatal("A path to an existing go source file is neeeded")
	//}

	// Positions are relative to fset
	fset := token.NewFileSet()
	// Parser mode flags, we want all every error possible and a trace printed on the stdout
	flags := parser.DeclarationErrors | parser.AllErrors
	// Parse the file identified by the given path and print the tree to the terminal.
	f, err := parser.ParseFile(fset, "example/_channel.go", nil, flags)

	if err != nil {
		log.Fatalf("ParseFile error: %s\n", err)
		return
	}

	// ! Debug Visitor to print to terminal in a more human readable manner the AST
	var debug debugVisitor
	ast.Walk(debug, f)
	fmt.Printf("\n\nGLOBAL SCOPE PARSER DEBUG PRINT \n")

	// Extracts the metadata about the given Go file and writes it to a JSON metadata file
	fileMetadata := choreia_parser.ExtractFileMetadata(f)
	fileDump, fileErr := os.Create("file_meta.json")
	jsonDump, jsonErr := json.MarshalIndent(fileMetadata, "", "  ")

	if jsonErr != nil || fileErr != nil {
		log.Fatal("Error encountered while writing JSON metadata file")
	}

	fileDump.WriteString(string(jsonDump))
}
