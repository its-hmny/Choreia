// Copyright 2020 Enea Guidi (hmny). All rights reserved.
// This file are distributed under the General Public License v 3.0.
// A copy of abovesaid license can be found in the LICENSE file.

// Package transforms declares the types and functions used to transform and work with some type of FSA.
// Come of the transformation implemented here are standard such as determinization (Subset Construction),
// minimization but more are specifically related to Choreia (GoroutineFSA extraction & Composition)
//
package transforms

import (
	list "github.com/emirpasic/gods/lists/singlylinkedlist"
	set "github.com/emirpasic/gods/sets/hashset"

	"github.com/its-hmny/Choreia/internal/data_structures/fsa"
)

// An adapted version of the classic Subset Construction Algorithm for FSA determinization.
// Allows to transform a Nondeterministic Finite State Automaton (NFA) to an equivalent
// Deterministic Finite State Automaton (DFA), the latter doesn't present eps-transition
// or duplicated parallel labels and its easier to be understood by humans
func SubsetConstruction(NCA *fsa.FSA) *fsa.FSA {
	DCA := fsa.New() // The deterministic version of the FSA

	// Initialization of the eps-closure of the initial state,
	initialClosure := newEpsClosure(NCA, set.New(0))
	//Init the tSet (a set of eps-closure)
	tSet := list.New(initialClosure)

	// Since the range statement uses a "frozen" version of the variable we use this trick
	// to enable working with "live" data and catch the mutations that are happining inside the loop
	for nIteration := 0; nIteration < tSet.Size(); nIteration++ {
		// Extracts the current closure to be evaluated
		item, _ := tSet.Get(nIteration)
		closure := item.(*set.Set)

		NCA.ForEachTransition(func(from, to int, t fsa.Transition) {
			if !closure.Contains(from) || t.Move == fsa.Eps {
				return // Skips the transitions that don't start from within the current closure
			}

			// Extracts the states that can be reached from the eps-closure with transition t.
			// Then computes the aggregate eps-closure of these reachable states
			moveEpsClosure := newEpsClosure(NCA, getReachable(NCA, closure, t))

			// Ignores empty eps-closure, this means that the transition function is not defined
			if moveEpsClosure.Size() <= 0 {
				return
			}

			// Checks if at least one state in the closure is a final state.
			// Then the new state in the DCA (the current closure) will be final as well
			containsFinalState := NCA.FinalStates.Any(func(_ int, value interface{}) bool {
				finalStateId := value.(int)
				return moveEpsClosure.Contains(finalStateId)
			})

			// If the eps-closure extracted already exist in tSet (has been already discovered)
			// then retrieves its twin's id from the map, and use the latter instead of the current id
			twinIndex, twinId := tSet.Find(func(_ int, item interface{}) bool {
				c := item.(*set.Set)
				// Simple tricK: If A is contained in B and viceversa then A equals B
				isAContained := c.Contains(moveEpsClosure.Values()...)
				isBContained := moveEpsClosure.Contains(c.Values()...)
				return isAContained && isBContained
			})

			if twinId == nil { // A twindId doesn't exist so a new state is created
				tSet.Add(moveEpsClosure)
				DCA.AddTransition(nIteration, fsa.NewState, t)
				// The new state as to be added to the final state list as well
				if containsFinalState {
					DCA.FinalStates.Add(DCA.GetLastId())
				}
			} else { // If a twin closure already exist its index is used to link the states with t
				DCA.AddTransition(nIteration, twinIndex, t)
			}
		})
	}

	return DCA
}

// Given a set of states extracts recursively the aggregate epsilon closure of said states
func newEpsClosure(automata *fsa.FSA, states *set.Set) *set.Set {
	// A set to keep track of all the states already reached
	reachedStates := set.New(states.Values()...) // Each state belongs to its own eps-closure

	automata.ForEachTransition(func(from, to int, t fsa.Transition) {
		// If the current is a eps transition starting from one of the already reached states
		if t.Move == fsa.Eps && reachedStates.Contains(from) {
			// We add the destination state to the eps-reachable list (the eps-closure)
			reachedStates.Add(to)
		}

	})

	// If we've reached more states than the previous call we search recursively
	if reachedStates.Size() > states.Size() {
		recursiveEpsClosure := newEpsClosure(automata, reachedStates)
		reachedStates.Add(recursiveEpsClosure.Values()...)
	}

	// Else we found all the states reachable and we return the full aggregate closure
	return reachedStates
}

// Returns a set of reachable states from a closure (or set of state) "clos" with the given move
// For move we mean a specific transition with a Move and Label fields
func getReachable(automata *fsa.FSA, clos *set.Set, move fsa.Transition) *set.Set {
	// Init an empty list of states reachable
	tReachable := set.New()

	automata.ForEachTransition(func(from, to int, t fsa.Transition) {
		if move.Move == t.Move && move.Label == t.Label && clos.Contains(from) {
			tReachable.Add(to)
		}
	})

	// Return the reachable states list
	return tReachable
}
