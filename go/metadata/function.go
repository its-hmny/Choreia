// Copyright 2020 Enea Guidi (hmny). All rights reserved.
// This files are distributed under the General Public License v3.0.
// A copy of aforesaid license can be found in the LICENSE file.

package metadata

import (
	"go/ast"

	log "github.com/sirupsen/logrus"
)

// Extracts recursively Channel and ControlFlow metadata from the function body.
// Satisfies the 'ast.Visitor' interface for metadata.Function.
func (fun *Function) Visit(node ast.Node) ast.Visitor {
	if node == nil { // Skips leaf/empty nodes in the AST
		return nil
	}

	// Type switch on the runtime value of the node
	switch statement := node.(type) {
	case *ast.AssignStmt:
		fun.FromAssignStmt(statement)
		return fun
	case *ast.GoStmt:
		fun.FromGoStmt(statement)
		return fun
	case *ast.SendStmt:
		fun.FromSendStmt(statement)
		return fun
	default:
		log.Infof("Unexpected statement '%T' at pos: %d", statement, node.Pos())
		return nil
	}
}

// Extracts metadata from 'ast.AssignStmt' node and updates the relative metadata.
func (fun *Function) FromAssignStmt(node *ast.AssignStmt) {
	if len(node.Lhs) != len(node.Rhs) {
		log.Fatalf("Mismatch between LHS and RHS expressions in ast.AssignStmt")
	}

	for i := 0; i < len(node.Lhs); i++ {
		lhsExpr, rhsExpr := node.Lhs[i], node.Rhs[i]

		funCallExpr, isCallExpr := rhsExpr.(*ast.CallExpr)
		funcIdent, isIdent := funCallExpr.Fun.(*ast.Ident)
		if !isCallExpr || !isIdent {
			fun.Visit(funCallExpr)
		}

		if funcIdent.Name == "make" {
			// We're handling separately 'make' calls that initialize channels
			chanIdent, isIdent := lhsExpr.(*ast.Ident)
			chanTypeExpr, isChanType := funCallExpr.Args[0].(*ast.ChanType)
			chanTypeIdent, isTypeIdent := chanTypeExpr.Value.(*ast.Ident)
			// The current 'make' is creating something else (array, slice, map, ...)
			if !isIdent || !isChanType || !isTypeIdent {
				return
			}

			// Adds the channel metadata to the function scope's metadata
			fun.Channels[chanIdent.Name] = Channel{Name: chanIdent.Name, MsgType: chanTypeIdent.Name}
		} else {
			// Other kind of function call, recurse on it in order to extract the CallTransition
			fun.Visit(funCallExpr)
		}
	}
}

// Extracts metadata from 'ast.GOStmt' node and updates the relative metadata.
func (fun *Function) FromGoStmt(node *ast.GoStmt) {
	for _, arg := range node.Call.Args {
		varIdent, isIdent := arg.(*ast.Ident)
		_, exists := fun.Channels[varIdent.Name]
		if !isIdent || !exists {
			fun.Visit(arg)
		} else {
			// TODO (hmny): Complete this function
		}
	}
}

// Extracts metadata from 'ast.SendStmt' node and updates the relative metadata.
func (fun *Function) FromSendStmt(node *ast.SendStmt) {
	_, isIdent := node.Chan.(*ast.Ident)
	if !isIdent {
		log.Fatalf("Expected ast.Ident in ast.SendStmt but got %T", node)
	}

	// TODO (hmny): Complete this function
}
