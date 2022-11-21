// Copyright 2020 Enea Guidi (hmny). All rights reserved.
// This files are distributed under the General Public License v3.0.
// A copy of aforesaid license can be found in the LICENSE file.

// Package metadata declares the types used to store metadata about Go's AST nodes.
package metadata

import (
	"os"

	log "github.com/sirupsen/logrus"
)

// Argument types that requires further computations when passed to another function
type ArgtType int

const (
	ArgChannel ArgtType = iota
	ArgWaitGroup
	ArgFunction
)

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

// Module initialization function, setups the logger module-wide
func init() {
	// Log as ASCII instead of the default JSON formatter.
	log.SetFormatter(&log.TextFormatter{ForceColors: true, FullTimestamp: true, TimestampFormat: "15:04:05"})
	// Output to stdout instead of the default stderr
	log.SetOutput(os.Stdout)
	// Only log the warning severity or above.
	log.SetLevel(log.TraceLevel)
}
