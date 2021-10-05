// Copyright Enea Guidi (hmny).

// This package handles the parsing of a given *ast.File which represents
// the content of a Go source file as an Abstract Syntax Tree.

// TODO COMMENT
package parser

import (
	"fmt"
	"go/ast"
)

func GetIterationStmtMetadata(fm *FunctionMetadata, node ast.Node) {
	switch stmt := node.(type) {
	// ! Add it back once implemented (priority given to the builtin concurrency construct)
	// case *ast.ForStmt, *ast.RangeStmt:
	default:
		fmt.Printf("Unrecognized nested (iteration) scope: %T \n", stmt)
	}
}
