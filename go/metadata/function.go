// Copyright 2020 Enea Guidi (hmny). All rights reserved.
// This files are distributed under the General Public License v3.0.
// A copy of abovesaid license can be found in the LICENSE file.

package metadata

import (
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

func (fun Function) Visit(node ast.Node) ast.Visitor {
	if node == nil { // Skips leaf/empty nodes in the AST
		return nil
	}

	// Type switch on the runtime value of the node
	switch statement := node.(type) {
	default:
		log.Fatalf("Unexpected statement '%T' at pos: %d", statement, node.Pos())
		statement.End()
		return nil
	}
}
