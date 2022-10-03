// Copyright 2020 Enea Guidi (hmny). All rights reserved.
// This files are distributed under the General Public License v3.0.
// A copy of abovesaid license can be found in the LICENSE file.

package metadata

import (
	"fmt"
	"go/ast"

	log "github.com/sirupsen/logrus"
)

// ----------------------------------------------------------------------------
// Function

// Represents and stores the informations extracted about any given function
// declared inside a given module. Mainly we're interested in saving the Name
// of the function and the Channels declared inside its scope. Also quite useful
// is an approximation of the ControlFlow of the function made with a Finite
// State Automata (FSA) which, in our case keeps track of Goroutine spawning
// and send/receive operations on both global and local-scoped channels.
// Also we keep track of Arguments which are meaningful to the function
// concurrent execution, some example may be channels, callbacks and waitgroups
// passed by the caller that may have some side effects on the concurrent
// system and overall 'Choreography'.
type Function struct {
	ast.Visitor `json:"-"`          // Implements the ast.Visitor interface (has the Visit(*ast.Node) function)
	Name        string              `json:"func_name"`         // Function name or identifier
	Arguments   map[string]Argument `json:"func_arguments"`    // "Meaningful" arguments passed by the caller
	Channels    map[string]Channel  `json:"func_channels"`     // Channels declared inside the function scope
	ControlFlow interface{}         `json:"func_control_flow"` // TODO: Add FSA package
}

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
		log.Fatalf("Unexpected statement '%T' at pos: %d", statement, node.Pos())
		statement.End()
		return nil
	}
}

// Extracts metadata from 'ast.AssignStmt' node and updates the relative metatdata.
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

// Extracts metadata from 'ast.GOStmt' node and updates the relative metatdata.
func (fun *Function) FromGoStmt(node *ast.GoStmt) {
	for _, arg := range node.Call.Args {
		varIdent, isIdent := arg.(*ast.Ident)
		chanMeta, exists := fun.Channels[varIdent.Name]
		if !isIdent || !exists {
			fun.Visit(arg)
		} else {
			// tx := Transaction{Op: Send, Channel: ChanMeta}
			// fun.Scope.AddTx(tx)
			fmt.Printf("Found channels passed to function %+v \n", chanMeta)
		}
	}
}

// Extracts metadata from 'ast.SendStmt' node and updates the relative metatdata.
func (fun *Function) FromSendStmt(node *ast.SendStmt) {
	chanIdent, isIdent := node.Chan.(*ast.Ident)
	if !isIdent {
		log.Fatalf("Expected ast.Ident in ast.SendStmt but got %T", node)
	}

	log.Tracef("Found 'Send op' on channel '%s' in function '%s'", chanIdent.Name, fun.Name)
	// ! sendTransition := SendTransition{Channel: chanIdent.Name}
	// ! fun.ControlFlow.AddTransition(t)
}
