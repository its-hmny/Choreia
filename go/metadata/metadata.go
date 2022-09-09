// Copyright 2020 Enea Guidi (hmny). All rights reserved.
// This files are distributed under the General Public License v3.0.
// A copy of abovesaid license can be found in the LICENSE file.

// Package metadata declares the types used to store metadata about Go's AST nodes.
package metadata

const (
	// List of allowed meaningful argument types that requires special elaborations
	ArgTypeChannel = iota
	ArgTypeFunction
	ArgTypeWaitGroup
)

// ----------------------------------------------------------------------------
// PackageMetadata

// Represents and stores the informations extracted about any given package
// used inside the provided Go program/project. Since we focus on concurrency
// and message passing we're mainly interested extracting informations about
// channels, function (that uses channels) and, eventually, the 'init' function
// of the package (for any side effects during module mounting).
type PackageMetadata struct {
	Name      string             `json:"pkg_name"`      // Package name or identifier
	Channels  []ChannelMetadata  `json:"pkg_channels"`  // Channels declared inside the module
	Functions []FunctionMetadata `json:"pkg_functions"` // Function declared inside the module
	InitFlow  interface{}        `json:"pkg_init_flow"` // TODO: Add FSA package
}

// ----------------------------------------------------------------------------
// FunctionMetadata

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
type FunctionMetadata struct {
	Name        string             `json:"func_name"`         // Function name or identifier
	Arguments   []ArgumentMetadata `json:"func_arguments"`    // "Meaningful" arguments passed by the caller
	Channels    []ChannelMetadata  `json:"func_channels"`     // Channels declared inside the function scope
	ControlFlow interface{}        `json:"func_control_flow"` // TODO: Add FSA package
}

// ----------------------------------------------------------------------------
// ChannelMetadata

// Represents and stores informations extracted about any given channel
// declared throughout the program/source code. We're only interested in
// the Name (also Identifier) of the channel and the type of the message
// exchanged through it for visualization pourposes.
type ChannelMetadata struct {
	Name    string `json:"chan_name"`     // Channel name or identifier
	MsgType string `json:"chan_msg_type"` // Type of message exchanged on channel
}

// ----------------------------------------------------------------------------
// ArgumentMetadata

// Represents and stores the informations extracted about any meaningful
// argument declared in the function signature. By meaningful argument we mean
// an argument which value/initialization may change the Choreograpy of the
// program. By passing one channel instead of another the function may communicate
// with a whole different set of Goroutines, the same applies for functions and
// callbacks and "possibly" WaitGroups.
type ArgumentMetadata struct {
	Name     string `json:"arg_name"`     // Argument name or identifier
	Position int    `json:"arg_position"` // Argument position (first, second, ...)
	ArgType  int    `json:"arg_type"`     // Type of the argument (Channel, Function, WaitGroup, ...)
}
