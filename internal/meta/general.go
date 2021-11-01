// Copyright Enea Guidi (hmny).

// This package handles the parsing of a given *ast.File which represents
// the content of a Go source file as an Abstract Syntax Tree.

// The only method avaiable from the outside is ParseAssignStmt and ParseExprStmt,
// both are two generic statement in which multiple interesting thing can occur
// and this method acts as "orchestrator" of their specilized counterparts
package meta

import (
	"go/ast"
	"log"
)

// This function parses an AssignStmt statement and evaluates all the possible cases for it.
// In particular this statement can have a recv from a channel, a function call or a channel init
// all three with the return value then assigned to a variable/identifier
func parseAssignStmt(stmt *ast.AssignStmt, fm *FuncMetadata) {
	// Check that the number of rvalues are the same of lvalues (values assignments) in the statement
	if len(stmt.Lhs) != len(stmt.Rhs) {
		log.Fatalf("Not the same number of lVal and rVal in AssignStmt at line %d\n", stmt.Pos())
	}

	// Now iterates over each assignment
	for i := range stmt.Lhs {
		lVal, rVal := stmt.Lhs[i], stmt.Rhs[i]
		// At the moment of writing this cast should always be successful
		identName := lVal.(*ast.Ident)

		switch castStmt := rVal.(type) {
		// Function call (+ assignment) or channel init
		case *ast.CallExpr:
			parseCallExpr(castStmt, fm)
			chanMeta := parseMakeCall(castStmt, identName.Name)
			fm.addChannels(chanMeta)
		// Receive (+ assignment) from a channel
		case *ast.UnaryExpr:
			parseRecvStmt(castStmt, fm)
		}
	}
}

// This function parses an ExprStmt statement and evaluates all the possible cases for it.
// In particular this statement can have a recv from a channel or a function call, both transition
// are extracted and handled specifically
func parseExprStmt(stmt *ast.ExprStmt, fm *FuncMetadata) {
	switch castStmt := stmt.X.(type) {
	case *ast.CallExpr:
		parseCallExpr(castStmt, fm)
	case *ast.UnaryExpr:
		parseRecvStmt(castStmt, fm)
	}

}
