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

	log.Fatalf("GenerateDCA not implemented")
	return fsa.New()
}

func handleRecvOp(simState State, t fsa.Transition, proc SimulationAutomata) {
	simChan, exist := simState.channelMap[t.Label]

	if !exist {
		// Retrieve channel metadata and create a new queue
		chanMeta := t.Payload.(meta.ChanMetadata)
		// Create a simulated channel and adds it to the map
		simChan = SimulationChannel{chanMeta, list.New(), list.New()}
		simState.channelMap[t.Label] = simChan
	}

	if !simChan.blockedSenders.Empty() {
		// Retrieves the blocked sender
		item, _ := simChan.blockedSenders.Get(0)
		blockedSender := item.(SimulationAutomata)
		// Adds the blocked sender and the current process to the list of active Goroutine
		simState.goroutineStack.Add(blockedSender)
		simState.goroutineStack.Insert(0, proc)
		// Removes from the blocked queue before returning
		simChan.blockedSenders.Remove(0)
	} else {
		simChan.blockedReceivers.Insert(0, proc)
	}
}

func handleSendOp(simState State, t fsa.Transition, proc SimulationAutomata) {
	simChan, exist := simState.channelMap[t.Label]

	if !exist {
		// Retrieve channel metadata and create a new queue
		chanMeta := t.Payload.(meta.ChanMetadata)
		// Create a simulated channel and adds it to the map
		simChan = SimulationChannel{chanMeta, list.New(), list.New()}
		simState.channelMap[t.Label] = simChan
	}

	if !simChan.blockedReceivers.Empty() {
		// Retrieves the blocked receiver
		item, _ := simChan.blockedReceivers.Get(0)
		blockedReceiver := item.(SimulationAutomata)
		// Adds the blocked receiver to the list of active Goroutine
		simState.goroutineStack.Add(blockedReceiver)
		simState.goroutineStack.Add(proc)
		// Removes from the blocked queue before returning
		simChan.blockedReceivers.Remove(0)
		return
	}

	// If the channel is synchronous (unbuffered) then the process is blocked
	//if !simChan.Async {
	simChan.blockedSenders.Insert(0, proc)
	//}
	// else { // Else the process is free to continue its execution
	//	simState.goroutineStack.Insert(0, proc)
	//}
}

func handleSpawnOp(simState State, t fsa.Transition, proc SimulationAutomata) {
	// Retrieves the ProjectionAutomata for the spawned Goroutine
	// and adds it to the simulation queue in order for it to be executed
	spawnedLW := t.Payload.(*ProjectionAutomata)
	simState.goroutineStack.Add(SimulationAutomata{*spawnedLW, 0})
	// Then the execution of the current Goroutine can continue
	simState.goroutineStack.Insert(0, proc)
}
