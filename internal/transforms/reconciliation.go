// Copyright 2020 Enea Guidi (hmny). All rights reserved.
// This file are distributed under the General Public License v 3.0.
// A copy of abovesaid license can be found in the LICENSE file.

// Package transforms declares the types and functions used to represent and work with
// ProjectionAutomata (also referenced as Local Views) for a given Goroutine. They also implements
// general pourpose algorithm for Finite State Automata (FSA) such as Subset Construction Algorithm
//
package transforms

import (
	"fmt"
	"log"

	list "github.com/emirpasic/gods/lists/singlylinkedlist"
	set "github.com/emirpasic/gods/sets/hashset"

	"github.com/its-hmny/Choreia/internal/data_structures/fsa"
)

type CompositionFSA *list.List // A list of (FrozenAutomata, FrozenAutomata) tuples

// A struct representing a "frozen" state of an FSA
type FrozenFSA struct {
	localView *ProjectionFSA // The "frozen" Automata
	state     int            // The state on which the automata is frozen
}

var wildcard = FrozenFSA{&ProjectionFSA{Name: "Wildcard"}, -1}

// TODO COMMENT
// TODO COMMENT
// TODO COMMENT
// TODO Refactor
func ForEachCoupleTransition(cFSA CompositionFSA, f func(A, B FrozenFSA, tA, tB fsa.Transition, toA, toB int)) {
	for _, item := range (*list.List)(cFSA).Values() {
		// Preliminaries conversion and extraction
		couple := item.(*set.Set)
		values := couple.Values()
		frozenA := values[0].(FrozenFSA)
		frozenB := values[1].(FrozenFSA)

		frozenA.localView.Automata.ForEachTransition(func(fromA, toA int, tA fsa.Transition) {
			frozenB.localView.Automata.ForEachTransition(func(fromB, toB int, tB fsa.Transition) {
				if fromA != frozenA.state || fromB != frozenB.state {
					return
				}

				f(frozenA, frozenB, tA, tB, toA, toB)
			})
		})
	}
}

// ! Must be implemented
// Takes the deterministic version of the Local Views (or Projection Automata) and merges them
// in one DCA that will represent the choreography as a whole (the global view)
func GenerateDCA(localViews map[string]*ProjectionFSA) *fsa.FSA {
	cFSA := fsaComposition(localViews)
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
func fsaComposition(localViews map[string]*ProjectionFSA) CompositionFSA {
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

			lView.Automata.ForEachState(func(lViewId int) {
				otherView.Automata.ForEachState(func(otherViewId int) {
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
func precalcSynchedCouples(cFSA CompositionFSA, entrypoint *set.Set) *list.List {
	// Creates the list with the synched couples
	synchedCouples := list.New(entrypoint)

	ForEachCoupleTransition(cFSA, func(fA, fB FrozenFSA, tA, tB fsa.Transition, toA, toB int) {
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

// TODO COMMENT
// TODO COMMENT
// TODO COMMENT
func fsaSynchronization(cFSA CompositionFSA, synchedCouples *list.List) *fsa.FSA {
	// Initializes the synchronized FSA
	synchAutomata := fsa.New()

	// ! synchedCouples.Any(func(_ int, item interface{}) bool {
	// ! 	couple := item.(*set.Set)
	// ! 	val := couple.Values()
	// ! 	A := val[0].(FrozenAutomata)
	// ! 	B := val[1].(FrozenAutomata)
	// ! 	fmt.Println(A.localView.Name, A.state, B.localView.Name, B.state)
	// ! 	return false
	// ! })
	// ! fmt.Println()

	ForEachCoupleTransition(cFSA, func(frozenA, frozenB FrozenFSA, tA, tB fsa.Transition, toA, toB int) {
		newFrozenA := FrozenFSA{frozenA.localView, toA}
		newFrozenB := FrozenFSA{frozenB.localView, toB}

		if tA.Move == fsa.Spawn {
			handleSpawn(synchAutomata, synchedCouples, frozenA, newFrozenA, tA)
		}

		if tB.Move == fsa.Spawn {
			handleSpawn(synchAutomata, synchedCouples, frozenB, newFrozenB, tB)
		}

		if tA.Move == fsa.Send && tB.Move == fsa.Recv && tA.Label == tB.Label {
			handleSync(synchAutomata, synchedCouples, frozenA, newFrozenA, frozenB, newFrozenB)
		} else if tB.Move == fsa.Send && tA.Move == fsa.Recv && tA.Label == tB.Label {
			handleSync(synchAutomata, synchedCouples, frozenB, newFrozenB, frozenA, newFrozenA)
		}
	})

	return synchAutomata
}

// TODO unify with handleSynch
func handleSpawn(sync *fsa.FSA, couples *list.List, prev, next FrozenFSA, t fsa.Transition) {

	id, _ := couples.Find(func(_ int, item interface{}) bool {
		couple := item.(*set.Set)
		return couple.Contains(next, wildcard)
	})

	if id == -1 {
		log.Fatal("Could not find couple")
	}

	interaction := fmt.Sprintf("%s ^ %s", prev.localView.Name, t.Label)
	newT := fsa.Transition{Move: fsa.Empty, Label: interaction}

	couples.Each(func(index int, item interface{}) {
		couple := item.(*set.Set)
		if couple.Contains(prev) {
			fmt.Printf("%d <> %d \t %s\n", index, id, newT)
			sync.AddTransition(index, id, newT)
		}
	})
}

// TODO unify with handleSpawn
func handleSync(sync *fsa.FSA, couples *list.List, sndrPrev, sndrNext, recvPrev, recvNext FrozenFSA) {
	id, _ := couples.Find(func(_ int, item interface{}) bool {
		couple := item.(*set.Set)
		return couple.Contains(sndrNext, recvNext)
	})

	if id == -1 {
		log.Fatal("Could not find couple")
	}

	interaction := fmt.Sprintf("%s %q %s", recvNext.localView.Name, '\u2190', sndrNext.localView.Name)
	newT := fsa.Transition{Move: fsa.Empty, Label: interaction}

	couples.Each(func(index int, item interface{}) {
		couple := item.(*set.Set)
		if couple.Contains(sndrPrev) || couple.Contains(recvPrev) {
			fmt.Printf("%d <> %d \t %s\n", index, id, newT)
			sync.AddTransition(index, id, newT)
		}
	})
}
