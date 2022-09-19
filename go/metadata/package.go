// Copyright 2020 Enea Guidi (hmny). All rights reserved.
// This files are distributed under the General Public License v3.0.
// A copy of abovesaid license can be found in the LICENSE file.

package metadata

import (
	"go/ast"

	log "github.com/sirupsen/logrus"
)

// ----------------------------------------------------------------------------
// Package

// Represents and stores the informations extracted about any given package
// used inside the provided Go program/project. Since we focus on concurrency
// and message passing we're mainly interested extracting informations about
// channels, function (that uses channels) and, eventually, the 'init' function
// of the package (for any side effects during module mounting).
type Package struct {
	ast.Visitor                     // Implements the ast.Visitor interface (has the Visit(*ast.Node) function)
	Name        string              `json:"pkg_name"`      // Package name or identifier
	Channels    map[string]Channel  `json:"pkg_channels"`  // Channels declared inside the module
	Functions   map[string]Function `json:"pkg_functions"` // Function declared inside the module
	InitFlow    interface{}         `json:"pkg_init_flow"` // TODO: Add FSA package
}

func (pkg Package) Visit(node ast.Node) ast.Visitor {
	// Skips leaf nodes in the AST
	if node == nil {
		return nil
	}

	// Type switch on the runtime value of the node
	switch statement := node.(type) {
	case *ast.GenDecl:
		pkg.FromGenDecl(statement)
		return pkg
	case *ast.FuncDecl:
		pkg.FromFuncDecl(statement)
		return pkg
	default:
		log.Fatalf("Unexpected statement '%T' at pos: %d", statement, node.Pos())
		return pkg
	}
}

func (pkg Package) FromGenDecl(declaration *ast.GenDecl) {
	for _, child := range declaration.Specs {
		switch specific := child.(type) {
		case *ast.ImportSpec:
			// TODO: Add module name <-> alias registration
			log.Warn("Found ast.ImportSpec in ast.GenDecl but ignored")
		case *ast.ValueSpec:
			log.Warn("Found ast.ValueSpec in ast.GenDecl but ignored")
		case *ast.TypeSpec:
			// TODO: Consider type definition only if it involves metadata.ArgType(s)
			log.Warn("Found ast.Typespec in ast.GenDecl but ignored")
		default:
			log.Fatalf("ast.GenDecl contains %T but it's not expected", specific)
		}
	}

}

func (pkg Package) FromFuncDecl(function *ast.FuncDecl) {
	// Creates a new FunctionMetadata instance to save the info
	meta := Function{Name: function.Name.Name, Arguments: map[string]Argument{}, Channels: map[string]Channel{}, ControlFlow: nil}
	log.Tracef("Found function '%s' in package '%s'", meta.Name, pkg.Name)

	// TODO: divide and normalize receiver methods from classic functions
	// TODO: Parse function.Type to extract arguments and return type

	// Extracts metadata recursively on the function scope and control flow
	for _, statement := range function.Body.List {
		statement.End()
		// meta.Visit(statement)
	}

}
