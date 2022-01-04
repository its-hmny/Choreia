// Copyright 2020 Enea Guidi (hmny). All rights reserved.
// This file are distributed under the General Public License v 3.0.
// A copy of abovesaid license can be found in the LICENSE file.

// Package transforms declares the types and functions used to represent and work with
// ProjectionAutomata (also referenced as Local Views) for a given Goroutine. They also implements
// general pourpose algorithm for Finite State Automata (FSA) such as Subset Construction Algorithm
//
package transforms

import (
	"log"

	"github.com/its-hmny/Choreia/internal/data_structures/fsa"
)

// ! Must be implemented
func GenerateDCA(localViews []*ProjectionAutomata) *fsa.FSA {
	// Takes the deterministic version of the Partial Automaton and merges them
	// in one DCA that will represent the choreography as a whole
	log.Fatalf("GenerateDCA not implemented")
	return fsa.New()
}
