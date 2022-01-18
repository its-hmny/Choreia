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

	list "github.com/emirpasic/gods/lists/singlylinkedlist"
	set "github.com/emirpasic/gods/sets/hashset"

	"github.com/its-hmny/Choreia/internal/data_structures/fsa"
)

const Wildcard = -1

type FrozenAutomata struct {
	localView *ProjectionAutomata
	state     int
}

type CompositionFSA struct {
	couples *list.List
}

// ! Must be implemented
// Takes the deterministic version of the Local Views (or Projection Automata) and merges them
// in one DCA that will represent the choreography as a whole (the global view)
func GenerateDCA(localViews map[string]*ProjectionAutomata) *fsa.FSA {
	// mainLW := localViews["Goroutine 0 (main)"]
	compositionFSA := FSAComposition(localViews)
	fmt.Printf("\nCompositionAutomata has %d states\n\n", compositionFSA.couples.Size())

	synchronizedFSA := Synchronization(compositionFSA, localViews["Goroutine 0 (main)"])

	return synchronizedFSA
}

// TODO COMMENT
// TODO COMMENT
// TODO COMMENT
func FSAComposition(localViews map[string]*ProjectionAutomata) CompositionFSA {
	cAutomata := CompositionFSA{list.New()}

	for _, lView := range localViews {
		for _, otherView := range localViews {
			// Avoids to compose the automata x with itself
			if lView == otherView {
				continue
			}

			lView.Automata.ForEachState(func(lViewId int) {
				otherView.Automata.ForEachState(func(otherViewId int) {
					// Creates the "frozen" instances (automata + state in which is frozen)
					frozenA := FrozenAutomata{lView, lViewId}
					frozenB := FrozenAutomata{otherView, otherViewId}

					// Checks that the couple hasn't been already indexed
					exist := cAutomata.couples.Any(func(_ int, item interface{}) bool {
						couple := item.(*set.Set)
						return couple.Contains(frozenA, frozenB)
					})

					// If the couple hasn't been indexed then is added
					if !exist {
						// ! REMOVE
						fmt.Printf("%s -> %d \t %s -> %d\n", lView.Name, lViewId, otherView.Name, otherViewId)
						cAutomata.couples.Add(set.New(frozenA, frozenB))
					}
				})
			})
		}
	}

	return cAutomata
}

// TODO COMMENT
// TODO COMMENT
// TODO COMMENT
func Synchronization(comp CompositionFSA, main *ProjectionAutomata) *fsa.FSA {
	synchAutomata := fsa.New()

	entrypoint := FrozenAutomata{main, 0}
	wildcard := FrozenAutomata{nil, Wildcard}

	visitedCouples := list.New(set.New(entrypoint, wildcard))

	for _, item := range comp.couples.Values() {
		// Preliminaries conversion and extraction
		couple := item.(*set.Set)
		values := couple.Values()
		frozenA := values[0].(FrozenAutomata)
		frozenB := values[1].(FrozenAutomata)

		frozenA.localView.Automata.ForEachTransition(func(fromA, toA int, tA fsa.Transition) {
			frozenB.localView.Automata.ForEachTransition(func(fromB, toB int, tB fsa.Transition) {
				if fromA != frozenA.state || fromB != frozenB.state {
					return
				}

				newFrozenA := FrozenAutomata{frozenA.localView, toA}
				newFrozenB := FrozenAutomata{frozenB.localView, toB}

				if tA.Move == fsa.Spawn {
					handleSpawn(synchAutomata, visitedCouples, frozenA, newFrozenA, tA)
				}

				if tB.Move == fsa.Spawn {
					handleSpawn(synchAutomata, visitedCouples, frozenB, newFrozenB, tB)
				}

				if tA.Move == fsa.Send && tB.Move == fsa.Recv && tA.Label == tB.Label {
					handleSync(synchAutomata, visitedCouples, frozenA, newFrozenA, frozenB, newFrozenB)
				} else if tB.Move == fsa.Send && tA.Move == fsa.Recv && tA.Label == tB.Label {
					handleSync(synchAutomata, visitedCouples, frozenB, newFrozenB, frozenA, newFrozenA)
				}
			})
		})
	}

	return synchAutomata
}

func handleSpawn(sync *fsa.FSA, couples *list.List, prev, next FrozenAutomata, t fsa.Transition) {
	wildcard := FrozenAutomata{nil, Wildcard}

	alreadyExist := couples.Any(func(_ int, item interface{}) bool {
		couple := item.(*set.Set)
		return couple.Contains(next, wildcard)
	})

	if !alreadyExist {
		interaction := fmt.Sprintf("%s ^ %s", prev.localView.Name, t.Label)
		newT := fsa.Transition{Move: fsa.Empty, Label: interaction}

		couples.Each(func(index int, item interface{}) {
			couple := item.(*set.Set)
			if couple.Contains(prev) {
				fmt.Printf("%d <> %d \t %s\n", index, couples.Size(), newT)
				sync.AddTransition(index, couples.Size(), newT)
			}
		})

		couples.Add(set.New(next, wildcard))
	}
}

func handleSync(sync *fsa.FSA, couples *list.List, sndrPrev, sndrNext, recvPrev, recvNext FrozenAutomata) {
	id, _ := couples.Find(func(_ int, item interface{}) bool {
		couple := item.(*set.Set)
		return couple.Contains(sndrNext, recvNext)
	})

	interaction := fmt.Sprintf("%s <- %s", recvNext.localView.Name, sndrNext.localView.Name)
	newT := fsa.Transition{Move: fsa.Empty, Label: interaction}

	if id == -1 {
		id = couples.Size()
		couples.Add(set.New(recvNext, sndrNext))
	}

	couples.Each(func(index int, item interface{}) {
		couple := item.(*set.Set)
		if couple.Contains(sndrPrev) || couple.Contains(recvPrev) {
			fmt.Printf("%d <> %d \t %s\n", index, id, newT)
			sync.AddTransition(index, id, newT)
		}
	})
}
