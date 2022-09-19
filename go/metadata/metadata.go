// Copyright 2020 Enea Guidi (hmny). All rights reserved.
// This files are distributed under the General Public License v3.0.
// A copy of abovesaid license can be found in the LICENSE file.

// Package metadata declares the types used to store metadata about Go's AST nodes.
package metadata

import (
	"os"

	log "github.com/sirupsen/logrus"
)

// List of meaningful argument types that requires further computations
// when passed as arguments to another function
type ArgType int

const (
	Chan ArgType = iota
	WaitGroup
	Func
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
	Name        string              `json:"func_name"`         // Function name or identifier
	Arguments   map[string]Argument `json:"func_arguments"`    // "Meaningful" arguments passed by the caller
	Channels    map[string]Channel  `json:"func_channels"`     // Channels declared inside the function scope
	ControlFlow interface{}         `json:"func_control_flow"` // TODO: Add FSA package
}

// ----------------------------------------------------------------------------
// Channel

// Represents and stores informations extracted about any given channel
// declared throughout the program/source code. We're only interested in
// the Name (also Identifier) of the channel and the type of the message
// exchanged through it for visualization pourposes.
type Channel struct {
	Name    string `json:"chan_name"`     // Channel name or identifier
	MsgType string `json:"chan_msg_type"` // Type of message exchanged on channel
}

// ----------------------------------------------------------------------------
// Argument

// Represents and stores the informations extracted about any meaningful
// argument declared in the function signature. By meaningful argument we mean
// an argument which value/initialization may change the Choreograpy of the
// program. By passing one channel instead of another the function may communicate
// with a whole different set of Goroutines, the same applies for functions and
// callbacks and "possibly" WaitGroups.
type Argument struct {
	Name     string  `json:"arg_name"`     // Argument name or identifier
	Position int     `json:"arg_position"` // Argument position in the function signature (first, second, ...)
	Type     ArgType `json:"arg_type"`     // Type of the argument (Channel, Function, WaitGroup, ...)
}

// Module initialization function, setups the logger module-wide
func init() {
	// Log as ASCII instead of the default JSON formatter.
	log.SetFormatter(&log.TextFormatter{ForceColors: true, FullTimestamp: true, TimestampFormat: "15:04:05"})
	// Output to stdout instead of the default stderr
	log.SetOutput(os.Stdout)
	// Only log the warning severity or above.
	log.SetLevel(log.TraceLevel)
}
