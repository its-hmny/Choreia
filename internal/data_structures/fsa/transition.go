// Copyright 2020 Enea Guidi (hmny). All rights reserved.
// This file are distributed under the General Public License v 3.0.
// A copy of abovesaid license can be found in the LICENSE file.

// This package implements a Finite State Automata (FSA) data structure and its own API.
// For this specific use cases the implementation is quite simple & basic

// The only struct available from the outside is Transition and its own API adn related enum
package fsa

import "fmt"

const (
	// Transition type enum
	Call  MoveKind = "Call"
	Empty MoveKind = "Empty"
	Eps   MoveKind = "Epsilon"
	Recv  MoveKind = "Recv"
	Send  MoveKind = "Send"
	Spawn MoveKind = "Spawn"
)

// Type alias to abstact the MoveKind enum
type MoveKind string

// ----------------------------------------------------------------------------
// Transition

// A Transition struct is a basic representation of a transition made inside a FSA
//
// The transition has an associated Kind/Move/Type associated to it, a label for
// simple explanation on the transition itself and a optional generic payload container
type Transition struct {
	Move    MoveKind    // The MoveType of Transition (Call, Eps, Recv, Send, Spawn)
	Label   string      // An explicative label of the action that is being executed
	Payload interface{} // A generic payload container for further info memorization
}

// Converts the Transition struct to a general pourpose string format.
func (t Transition) String() string {
	switch t.Move {
	case Eps:
		return fmt.Sprintf("%q %s", '\u03B5', t.Label)
	case Recv:
		return fmt.Sprintf("%q %s", '\u2190', t.Label)
	case Send:
		return fmt.Sprintf("%q %s", '\u2192', t.Label)
	case Call:
		return fmt.Sprintf("%q %s", '\u2A0F', t.Label)
	case Spawn:
		return fmt.Sprintf("%q %s", '\u22C1', t.Label)
	case Empty:
		return t.Label
	default:
		return fmt.Sprintf("%q %s", '\u2048', t.Label)
	}
}
