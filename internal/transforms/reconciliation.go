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
	meta "github.com/its-hmny/Choreia/internal/static_analysis"

	"github.com/its-hmny/Choreia/internal/data_structures/fsa"
)

type SimulationChannel struct {
	meta.ChanMetadata
	blockedReceivers *list.List
	blockedSenders   *list.List
}

type SimulationAutomata struct {
	ProjectionAutomata
	currentState int
}

type State struct {
	goroutineStack *list.List
	channelMap     map[string]SimulationChannel
}

func (s SimulationAutomata) forEachNextTransition(callback func(from, to int, t fsa.Transition)) {
	s.Automata.ForEachTransition(func(from, to int, t fsa.Transition) {
		if from != s.currentState {
			return
		}

		callback(from, to, t)
	})
}

// ! Must be implemented
// Takes the deterministic version of the Local Views (or Projection Automata) and merges them
// in one DCA that will represent the choreography as a whole (the global view)
func GenerateDCA(localViews map[string]*ProjectionAutomata) *fsa.FSA {
	mainLW := localViews["Goroutine 0 (main)"]

	simState := State{
		goroutineStack: list.New(SimulationAutomata{*mainLW, 0}),
		channelMap:     make(map[string]SimulationChannel),
	}

	for !simState.goroutineStack.Empty() {
		// Retireves the Goroutine to be processed (simulated)
		item, _ := simState.goroutineStack.Get(0)
		simProc := item.(SimulationAutomata)

		// Removes it from the queue (will eventually be added later)
		simState.goroutineStack.Remove(0)

		fmt.Printf("%s =>\t", simProc.Name)
		for _, tmp := range simState.goroutineStack.Values() {
			fmt.Printf("%s,  ", tmp.(SimulationAutomata).Name)
		}
		fmt.Println()

		// Process every transition possible from the current simulation state
		simProc.forEachNextTransition(func(from, to int, t fsa.Transition) {
			copyProcess := simProc
			copyProcess.currentState = to

			switch t.Move {
			case fsa.Send:
				handleSendOp(simState, t, copyProcess)
			case fsa.Recv:
				handleRecvOp(simState, t, copyProcess)
			case fsa.Spawn:
				handleSpawnOp(simState, t, copyProcess)
			default:
				log.Fatalf("Unexpected transition: %s", t)
			}
		})

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
