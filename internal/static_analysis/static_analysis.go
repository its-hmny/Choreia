// Copyright 2020 Enea Guidi (hmny). All rights reserved.
// This file are distributed under the General Public License v 3.0.
// A copy of abovesaid license can be found in the LICENSE file.

// Package static_analysis declares the types used to represent metedata extracted from the Go source code.
// The source code is transformed to an Abstract Syntax Tree via go/ast module and. Said AST is visited fully
// and all the metadata needed are extractred then returned in a single aggregate struct.
//
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
