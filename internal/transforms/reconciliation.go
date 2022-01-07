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

	list "github.com/emirpasic/gods/lists/singlylinkedlist"
	meta "github.com/its-hmny/Choreia/internal/static_analysis"

	"github.com/its-hmny/Choreia/internal/data_structures/fsa"
)

type SimulationChannel struct {
	meta.ChanMetadata
	msgQueue *list.List
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

	// ToDo change with full on iterator
	for !simState.goroutineStack.Empty() {
		// Retireves the Goroutine to be processed (simulated)
		item, _ := simState.goroutineStack.Get(0)
		simProc := item.(SimulationAutomata)

		// Removes it from the queue (will eventually be added later)
		simState.goroutineStack.Remove(0)

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
				log.Fatalf("Unexpected transition %s", t)
			}
		})

	}

	log.Fatalf("GenerateDCA not implemented")
	return fsa.New()
}

func handleRecvOp(simState State, t fsa.Transition, proc SimulationAutomata) {
	simChan, exist := simState.channelMap[t.Label]

	// Since the data about the channel doesn't even exist the process is blocked in the queue
	if !exist {
		// Retrieve channel metadata and create a new queue
		chanMeta := t.Payload.(meta.ChanMetadata)
		msgQueue := list.New(proc)
		// Create a simulated channel and adds it to the map before returning
		newSimChan := SimulationChannel{chanMeta, msgQueue}
		simState.channelMap[t.Label] = newSimChan
		return
	}

	// If the channel is buffered (synchronous) or the buffer is empty then
	// the proces must block onto it
	if !simChan.Async || simChan.msgQueue.Empty() {
		simChan.msgQueue.Add(proc)
		return
	}

	// Get the first element and return it without blocking the receiver
	msg, _ := simChan.msgQueue.Get(0)
	simState.goroutineStack.Add(msg)
	simChan.msgQueue.Remove(0)
}

func handleSendOp(simState State, t fsa.Transition, proc SimulationAutomata) {
	simChan, exist := simState.channelMap[t.Label]

	if !exist {
		// Retrieve channel metadata and create a new queue
		chanMeta := t.Payload.(meta.ChanMetadata)
		msgQueue := list.New(proc)
		// Create a simulated channel and adds it to the map before returning
		newSimChan := SimulationChannel{chanMeta, msgQueue}
		simState.channelMap[t.Label] = newSimChan
		return
	}

	if !simChan.Async || simChan.msgQueue.Size() > simChan.BufferSize {
		simChan.msgQueue.Add(proc)
		return
	}

	msg, _ := simChan.msgQueue.Get(0)
	simState.goroutineStack.Add(msg)
	simChan.msgQueue.Remove(0)
}

func handleSpawnOp(simState State, t fsa.Transition, proc SimulationAutomata) {
	// Retrieves the ProjectionAutomata for the spawned Goroutine
	// and adds it to the simulation queue in order for it to be executed
	spawnedLW := t.Payload.(*ProjectionAutomata)
	simState.goroutineStack.Add(SimulationAutomata{*spawnedLW, 0})
	// Then the execution of the current Goroutine can continue
	simState.goroutineStack.Insert(0, proc)
}
