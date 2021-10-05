// Copyright Enea Guidi (hmny).

// This package handles the parsing of a given *ast.File which represents
// the content of a Go source file as an Abstract Syntax Tree.

// TODO COMMENT
package parser

import (
	"fmt"
	"go/ast"
)

func GetBranchStmtMetadata(fm *FunctionMetadata, node ast.Node) {
	switch stmt := node.(type) {
	// ! Add it back once implemented (priority given to the builtin concurrency construct)
	// case *ast.IfStmt:
	// parseIfStmt(fm, stmt)
	// case *ast.SwitchStmt, case *ast.TypeSwitchStmt:
	default:
		fmt.Printf("Unrecognized nested (branch) scope: %T \n", stmt)
	}
}

// func parseIfStmt(fm *FunctionMetadata, stmt *ast.IfStmt) {
// Parses the if statement condition and initialize statements
// ast.Walk(fm, stmt.Init)
// ast.Walk(fm, stmt.Cond)
//
// Makes two copies of the state value
// stateCopy_1, stateCopy_2 := *fm.currentState, *fm.currentState
//
// Add then an eps transition before parsing the body of the statement since
// at runtime isn't granted that the function will enter there
// ast.Walk(fm, stmt.Body)
//
// If the if statement contains an else branch then that is parsed alternatively
// an eps transition is put in place to represent and a false condition statement case
// if stmt.Else != nil {
// ast.Walk(fm, stmt.Else)
// } else {
// epsTransition := Transaction{Eps, "EpsTrans", initialStateCopy, *fm.currentState}
// fm.addTransactions(epsTransition)
// }
// }
