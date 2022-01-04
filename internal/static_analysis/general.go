// Copyright 2020 Enea Guidi (hmny). All rights reserved.
// This file are distributed under the General Public License v 3.0.
// A copy of abovesaid license can be found in the LICENSE file.

// Package static_analysis declares the types used to represent metedata extracted from the Go source code.
// The source code is transformed to an Abstract Syntax Tree via go/ast module and. Said AST is visited fully
// and all the metadata needed are extractred then returned in a single aggregate struct.
//
package static_analysis

import (
	"go/ast"
	"log"
)

// This function parses an AssignStmt statement and evaluates all the possible cases for it.
// In particular this statement can have a recv from a channel, a function call or a channel init
// all three with the return value then assigned to a variable/identifier
func parseAssignStmt(stmt *ast.AssignStmt, fm *FuncMetadata) {
	// Check that the number of rvalue are the same of lvalue (values assignments) in the statement
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
