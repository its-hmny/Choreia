// Copyright 2020 Enea Guidi (hmny). All rights reserved.
// This files are distributed under the General Public License v3.0.
// A copy of aforesaid license can be found in the LICENSE file.

package metadata

import (
	"go/ast"

	log "github.com/sirupsen/logrus"
)

func (pkg Package) Visit(node ast.Node) ast.Visitor {
	if node == nil { // Skips leaf/empty nodes in the AST
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
		log.Infof("Unexpected statement '%T' at pos: %d", statement, node.Pos())
		return nil
	}
}

func (pkg *Package) FromGenDecl(declaration *ast.GenDecl) {
	for _, child := range declaration.Specs {
		switch specific := child.(type) {
		case *ast.ImportSpec:
			// TODO: Add module import recursive parsing support
			log.Warn("Found ast.ImportSpec in ast.GenDecl")
		case *ast.ValueSpec:
			// TODO: Add support for global declaration ('const' and 'var')
			log.Warn("Found ast.ValueSpec in ast.GenDecl")
		case *ast.TypeSpec:
			// TODO: Consider type definition only if it involves metadata.ArgType(s)
			log.Warn("Found ast.Typespec in ast.GenDecl")
		default:
			log.Fatalf("ast.GenDecl contains %T but it's not expected", specific)
		}
	}

}

func (pkg *Package) FromFuncDecl(function *ast.FuncDecl) {
	// Creates a new FunctionMetadata instance to save the info
	meta := Function{Name: function.Name.Name, Arguments: map[string]Argument{}, Channels: map[string]Channel{}, ControlFlow: nil}
	log.Tracef("Found function '%s' in package '%s'", meta.Name, pkg.Name)

	// TODO: divide and normalize receiver methods from classic functions
	// TODO: Parse function.Type to extract arguments and return type

	// Extracts metadata recursively on the function scope and control flow
	for _, statement := range function.Body.List {
		meta.Visit(statement)
	}

	// Registers the complete function meta in the parent scope
	pkg.Functions[meta.Name] = meta
}
