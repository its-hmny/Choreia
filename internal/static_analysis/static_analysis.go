// Copyright Enea Guidi (hmny).

// This package handles the parsing of a given file which should represents
// the content of a Go source file as an Abstract Syntax Tree.

// The only method available from the outside is ExtractMetadata which is the only entrypoint
// of the module and will return a FileMetadata struct containing some info bout provided file.
package static_analysis

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"strings"
)

const (
	NoTrace TraceMode = iota
	BasicTrace
	ExtendedTrace

	// parser.ParseFIle default flags, we want all every error possible
	defaultFlags = parser.DeclarationErrors | parser.AllErrors
)

// Simple type alias to wrap trace option definition
type TraceMode int

// A simple type alias that fullfills the ast.Visitor interface and implements
// the Visit() function, this will print a trace on the stdin of the structure of the AST
type traceVisitor int

func (v traceVisitor) Visit(node ast.Node) ast.Visitor {
	if node == nil {
		return nil
	}

	fmt.Printf("%s %T\n", strings.Repeat("  ", int(v)), node)
	return v + 1
}

// ----------------------------------------------------------------------------
// Meta package API

// Parses the file identified by the given path, if the latter is valid, if the user
// opted in the available trace option handles the traces as well then extracts the metadata
// from the AST and returns said metadata to the caller
func ExtractMetadata(filePath string, traceOpts TraceMode) FileMetadata {
	// At first checks that the given input path actually exists
	if fStat, err := os.Stat(filePath); os.IsNotExist(err) || fStat.IsDir() {
		log.Fatal("A path to an existing go source file is needed")
	}

	// Parses the file and retrieves the AST
	f, err := parser.ParseFile(token.NewFileSet(), filePath, nil, defaultFlags)

	if err != nil {
		log.Fatal(err)
	}

	if traceOpts == BasicTrace {
		// Descend the AST in depth first order, printing a small trace on the stdin
		ast.Walk(traceVisitor(0), f)
	} else if traceOpts == ExtendedTrace {
		// Parses again the whole file but adds the built-in trace option with more insights
		parser.ParseFile(token.NewFileSet(), filePath, nil, defaultFlags|parser.Trace)
	}

	return parseAstFile(f)
}
