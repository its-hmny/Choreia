// Copyright 2020 Enea Guidi (hmny). All rights reserved.
// This files are distributed under the General Public License v3.0.
// A copy of aforesaid license can be found in the LICENSE file.

// Package metadata declares the types used to store metadata about Go's AST nodes.
package metadata

import (
	"go/ast"
	"os"

	log "github.com/sirupsen/logrus"
)

// ----------------------------------------------------------------------------
// Package

// Represents and stores the information extracted about any given package
// used inside the provided Go program/project. Since we focus on concurrency
// and message passing we're mainly interested extracting information about
// channels, function (that uses channels) and, eventually, the 'init' function
// of the package (for any side effects during module mounting).
type Package struct {
	ast.Visitor `json:"-"`          // Implements the ast.Visitor interface (has the Visit(*ast.Node) function)
	Name        string              `json:"name"`      // Package name or identifier
	Channels    map[string]Channel  `json:"channels"`  // Channels declared inside the module
	Functions   map[string]Function `json:"functions"` // Function declared inside the module
	InitFlow    interface{}         `json:"init_flow"` // TODO: Add FSA package
}

// ----------------------------------------------------------------------------
// Function

// Represents and stores the information extracted about any given function
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
	Name        string              `json:"name"`         // Function name or identifier
	Arguments   map[string]Argument `json:"arguments"`    // "Meaningful" arguments passed by the caller
	Channels    map[string]Channel  `json:"channels"`     // Channels declared inside the function scope
	ControlFlow interface{}         `json:"control_flow"` // TODO: Add FSA package
}

// ----------------------------------------------------------------------------
// Argument

// Represents and stores the information extracted about any meaningful
// argument declared in the function signature. By meaningful argument we mean
// an argument which value/initialization may change the Choreography of the
// program. By passing one channel instead of another the function may communicate
// with a whole different set of Goroutines, the same applies for functions and
// callbacks and "possibly" WaitGroups.
type Argument struct {
	Name     string   `json:"name"`     // Argument name or identifier
	Position int      `json:"position"` // Argument position in the function signature (first, second, ...)
	Type     ArgtType `json:"type"`     // Type of the argument (Channel, Function, WaitGroup, ...)
}

// ----------------------------------------------------------------------------
// Channel

// Represents and stores information extracted about any given channel
// declared throughout the program/source code. We're only interested in
// the Name (also Identifier) of the channel and the type of the message
// exchanged through it for visualization purposes.
type Channel struct {
	Name    string `json:"name"`     // Channel name or identifier
	MsgType string `json:"msg_type"` // Type of message exchanged on channel
}

// Argument types that requires further computations when passed to another function
type ArgtType int

const (
	ArgChannel ArgtType = iota
	ArgWaitGroup
	ArgFunction
)

// Module initialization function, setups the logger module-wide
func init() {
	// Log as ASCII instead of the default JSON formatter.
	log.SetFormatter(&log.TextFormatter{ForceColors: true, FullTimestamp: true, TimestampFormat: "15:04:05"})
	// Output to stdout instead of the default stderr
	log.SetOutput(os.Stdout)
	// Only log the warning severity or above.
	log.SetLevel(log.TraceLevel)
}
