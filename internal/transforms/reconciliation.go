// Copyright 2020 Enea Guidi (hmny). All rights reserved.
// This file are distributed under the General Public License v 3.0.
// A copy of abovesaid license can be found in the LICENSE file.

// Package transforms declares the types and functions used to transform and work with some type of FSA.
// Come of the transformation implemented here are standard such as determinization (Subset Construction),
// minimization but more are specifically related to Choreia (GoroutineFSA extraction & Composition)
//
package transforms

import (
	"fmt"
	"log"

	list "github.com/emirpasic/gods/lists/singlylinkedlist"
	set "github.com/emirpasic/gods/sets/hashset"

	"github.com/its-hmny/Choreia/internal/data_structures/fsa"
)

type ProductFSA *list.List // A list of (FrozenAutomata, FrozenAutomata) tuples

// A struct representing a "frozen" state of an FSA
type FrozenFSA struct {
	localView *GoroutineFSA // The "frozen" Automata
	state     int           // The state on which the automata is frozen
}

// A wildcard variable used as second item in a couple when needed
var wildcard = FrozenFSA{&GoroutineFSA{Name: "Wildcard"}, -1}

// Utility function to iterate over every possible combination of transition (tA and tB) for
// a given state of the Composition FSA which is a couple of states from different FSAs
func forEachCoupleTransition(cFSA ProductFSA, f func(A, B FrozenFSA, tA, tB fsa.Transition, toA, toB int)) {
	for _, item := range (*list.List)(cFSA).Values() {
		// Preliminaries conversion and extraction
		couple := item.(*set.Set)
		values := couple.Values()
		frozenA := values[0].(FrozenFSA)
		frozenB := values[1].(FrozenFSA)

		frozenA.localView.Automaton.ForEachTransition(func(fromA, toA int, tA fsa.Transition) {
			frozenB.localView.Automaton.ForEachTransition(func(fromB, toB int, tB fsa.Transition) {
				if fromA != frozenA.state || fromB != frozenB.state {
					return
				}

				f(frozenA, frozenB, tA, tB, toA, toB)
			})
		})
	}
}

// Utility function that searches for a couple (the set) into a list of said couples.
// Since the list is assumed to have all the couples the case in which the couple is not
// found is not contemplated and will stop the execution with an error
func findCoupleId(list *list.List, toFind *set.Set) int {
	id, _ := list.Find(func(_ int, item interface{}) bool {
		couple := item.(*set.Set)
		return couple.Contains(toFind.Values()...)
	})

	if id == -1 {
		log.Fatal("Could not find couple")
	}

	return id
}

// Utility functions that creates a transition from every state that contains at least one
// element in the fromCouple to the state identified by destId and with newT transitions
func createTransitions(syncFSA *fsa.FSA, couples *list.List, fromCouple *set.Set, destId int, newT fsa.Transition) {
	couples.Each(func(currentId int, item interface{}) {
		couple := item.(*set.Set)
		for _, frozenFSA := range fromCouple.Values() {
			if couple.Contains(frozenFSA) {
				fmt.Printf("(From %d => to %d) \t %s\n", currentId, destId, newT)
				syncFSA.AddTransition(currentId, destId, newT)
			}
		}
	})
}

// Takes the deterministic version of the Local Views (or Projection Automata) and merges them
// in one DCA that will represent the choreography as a whole (the global view). This is possible
// by composing all the Local View's FSAs into one and then appply a Synchronization transform on it
func LocalViewsComposition(localViews map[string]*GoroutineFSA) *fsa.FSA {
	cFSA := fsaProduct(localViews)
	fmt.Printf("CompositionAutomata has %d states\n\n", ((*list.List)(cFSA)).Size())

	// Creates the entrypoint couples (main - 0, wildcard), the starting couple of the program
	mainKey := fmt.Sprintf(nameTemplate, "main", 0)
	entrypointCouple := set.New(FrozenFSA{localViews[mainKey], 0}, wildcard)

	// Precalc the "synched" couples, the one in which the two process could interact between them
	precalcCouples := precalcSynchedCouples(cFSA, entrypointCouple)

	// With the precalc couple in which the local views synchs and the full composition automata
	// the full Choreography Automata (global view) is generated and returned
	return fsaSynchronization(cFSA, precalcCouples)
}

// Takes two or more FSA given as input and returns the composition FSA of given automata
// the returned automata is a FSA with m x n x ... z states and all the transitions of the
// starting FSAs combined, every possible combination is only added once.
func fsaProduct(localViews map[string]*GoroutineFSA) ProductFSA {
	// Creates a new list (type alias of CompositionFSA)
	cAutomata := list.New()

	// Creates all the couples iterating on each automata and each state of the latter
	// and composes it with each other automata and their respective states
	for _, lView := range localViews {
		for _, otherView := range localViews {
			// Avoids to compose the automata x with itself
			if lView == otherView {
				continue
			}

			lView.Automaton.ForEachState(func(lViewId int) {
				otherView.Automaton.ForEachState(func(otherViewId int) {
					// Creates the "frozen" instances (automata + state in which is frozen)
					frozenA := FrozenFSA{lView, lViewId}
					frozenB := FrozenFSA{otherView, otherViewId}

					// Checks that the couple hasn't been already indexed
					exist := cAutomata.Any(func(_ int, item interface{}) bool {
						couple := item.(*set.Set)
						return couple.Contains(frozenA, frozenB)
					})

					if !exist { // If the couple hasn't been indexed then is added
						cAutomata.Add(set.New(frozenA, frozenB))
					}
				})
			})
		}
	}

	return cAutomata // Returns the composition finite state automata
}

// Given a composition FSA and the entrypoint (the first state) for the first it precalculate
// the state of the cFSA in which a synchronization occurs. this means it returns a subset of tuples
// <state, state> in which 2 actor or local views interact between them
func precalcSynchedCouples(cFSA ProductFSA, entrypoint *set.Set) *list.List {
	// Creates the list with the synched couples
	synchedCouples := list.New(entrypoint)

	forEachCoupleTransition(cFSA, func(fA, fB FrozenFSA, tA, tB fsa.Transition, toA, toB int) {
		var couple *set.Set // Initializes and empty couple

		// Retrieve the "destination" couple of the current one
		newFrozenA := FrozenFSA{fA.localView, toA}
		newFrozenB := FrozenFSA{fB.localView, toB}

		// Check for interaction between A and B (A sends, B receives or the opposite)
		hasA2B := tA.Move == fsa.Send && tB.Move == fsa.Recv && tA.Label == tB.Label
		hasB2A := tB.Move == fsa.Send && tA.Move == fsa.Recv && tA.Label == tB.Label

		// If A or B have a Spawn transition then the couple <spawner, *> is considered "synched"
		if tA.Move == fsa.Spawn {
			couple = set.New(newFrozenA, wildcard)
		} else if tB.Move == fsa.Spawn {
			couple = set.New(newFrozenB, wildcard)
		} else if hasA2B || hasB2A { // If A and B interact between them the couple is "synched"
			couple = set.New(newFrozenA, newFrozenB)
		} else { // Else the couple is not "synched" and we skip the iteration
			return
		}

		// Checks that the couple has not been already indexed (every couple is indexed only once)
		alreadyExist := synchedCouples.Any(func(_ int, item interface{}) bool {
			current := item.(*set.Set)
			return current.Contains(couple.Values()...)
		})

		if !alreadyExist {
			synchedCouples.Add(couple)
		}
	})

	return synchedCouples // Returns the "synched" couple list
}

// Iterates over the composition FSA and whenever it found a couple of state (and their respective FSA
// & transitions) that can be synchronized: 1) they make their own operations (e.g. Spawn) they make
// opposite transition on the same channel (Send & Receive on x) then it links this couple with every other
// couple in the synchronization FSA that can reach the current one.
func fsaSynchronization(cFSA ProductFSA, synchedCouples *list.List) *fsa.FSA {
	// Initializes the synchronized FSA
	synchAutomata := fsa.New()

	// ! Refactor this mess
	forEachCoupleTransition(cFSA, func(frozenA, frozenB FrozenFSA, tA, tB fsa.Transition, toA, toB int) {
		newFrozenA := FrozenFSA{frozenA.localView, toA}
		newFrozenB := FrozenFSA{frozenB.localView, toB}

		if tA.Move == fsa.Spawn {
			// Find the id of the current couple in the precalc list
			id := findCoupleId(synchedCouples, set.New(newFrozenA, wildcard))
			// Generate the new transition with label
			interactionLabel := fmt.Sprintf("%s %q %s", frozenA.localView.Name, '\u22C1', tA.Label)
			newT := fsa.Transition{Move: fsa.Empty, Label: interactionLabel}
			// Add said transition to the final synchronization FSA
			createTransitions(synchAutomata, synchedCouples, set.New(frozenA), id, newT)
		}

		if tB.Move == fsa.Spawn {
			// Find the id of the current couple in the precalc list
			id := findCoupleId(synchedCouples, set.New(newFrozenB, wildcard))
			// Generate the new transition with label
			interactionLabel := fmt.Sprintf("%s %q %s", frozenB.localView.Name, '\u22C1', tB.Label)
			newT := fsa.Transition{Move: fsa.Empty, Label: interactionLabel}
			// Add said transition to the final synchronization FSA
			createTransitions(synchAutomata, synchedCouples, set.New(frozenB), id, newT)
		}

		if tA.Move == fsa.Send && tB.Move == fsa.Recv && tA.Label == tB.Label {
			// Find the id of the current couple in the precalc list
			id := findCoupleId(synchedCouples, set.New(newFrozenA, newFrozenB))
			// Generate the new transition with label
			interactionLabel := fmt.Sprintf("%s %q %s", frozenB.localView.Name, '\u2190', frozenA.localView.Name)
			newT := fsa.Transition{Move: fsa.Empty, Label: interactionLabel}
			// Add said transition to the final synchronization FSA
			createTransitions(synchAutomata, synchedCouples, set.New(frozenA, frozenB), id, newT)
		} else if tB.Move == fsa.Send && tA.Move == fsa.Recv && tA.Label == tB.Label {
			// Find the id of the current couple in the precalc list
			id := findCoupleId(synchedCouples, set.New(newFrozenA, newFrozenB))
			// Generate the new transition with label
			interactionLabel := fmt.Sprintf("%s %q %s", frozenA.localView.Name, '\u2190', frozenB.localView.Name)
			newT := fsa.Transition{Move: fsa.Empty, Label: interactionLabel}
			// Add said transition to the final synchronization FSA
			createTransitions(synchAutomata, synchedCouples, set.New(frozenA, frozenB), id, newT)
		}
	})

	return synchAutomata
}
