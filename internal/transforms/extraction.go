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

	"github.com/its-hmny/Choreia/internal/data_structures/fsa"
	meta "github.com/its-hmny/Choreia/internal/static_analysis"
)

var (
	nGoroutineStarted = 0
	inlinedCache      = make(map[string]*fsa.FSA)
)

const nameTemplate = "%s (%d)"

// -------------------------------------------------------------------------------------------
// GoroutineFSA

// A FSA that represents the execution flow of a single Goroutine (identified by its own name)
// this will be used in the future phases as a local view for the whole choreography.
// This means that this FSA will provide an "isolated" view of the choreography from the perspective
// of the Goroutine that takes part in it (almost like a projection of the whole choreography)
type GoroutineFSA struct {
	Name string // An identifier for the Automata
	meta.FuncMetadata
}

// Given the metadata associated to a file it linearizes the automata found in it
// (function calls inlining). Once done that extracts recursively the FSA associated to
// each Goroutine spawned during the program execution, the latter are returned as output
func ExtractGoroutineFSA(file meta.FileMetadata) map[string]*GoroutineFSA {
	// Cleanup function that resets the global variable nGoroutineStarted & inlinedCache
	defer func() {
		nGoroutineStarted = 0
		inlinedCache = make(map[string]*fsa.FSA)
	}()

	for _, function := range file.FunctionMeta {
		// Cache hit: The current automaton has already been linearized.
		if inlinedCache[function.Name] != nil {
			// This means its function calls in the automaton have been already inlined and the latter
			// contains (as subgraphs) the FSA of the function called (just like compiler inlining)
			continue
		}

		linearizeFSA(function, file, inlinedCache) // Cache miss: We must linearize the current automaton
	}

	name := fmt.Sprintf(nameTemplate, "main", nGoroutineStarted)
	meta, existMeta := file.FunctionMeta["main"]
	mainGrFSA := GoroutineFSA{name, meta}

	automaton, existLin := inlinedCache["main"]
	mainGrFSA.Automaton = automaton.Copy()

	if !existMeta || !existLin {
		log.Fatal("Automaton or meta associated to 'main' function not found")
	}

	// Extracts all the GoroutineFSA starting from the "main" function
	// which is the entrypoint for the Go program
	return extractSpawnTree(mainGrFSA, file)
}

// Given an entrypoint (a Goroutine FSA) extracts recursively all the Goroutine spawned during
// the execution of said Goroutine. Before the recursive call the formal args are replaced with
// the actual ones. If A spawns B and B spawns C then extractSpawnTree(A) will return both B, C
// since the latter is in B subtree but also in A subtree.
func extractSpawnTree(gr GoroutineFSA, file meta.FileMetadata) map[string]*GoroutineFSA {
	spawnedGoroutines := make(map[string]*GoroutineFSA)

	gr.Automaton.ForEachTransition(func(from, to int, t fsa.Transition) {
		// We're only interested in the spawn of another Goroutine
		if t.Move != fsa.Spawn {
			return
		}

		nGoroutineStarted++
		spawnedName := fmt.Sprintf(nameTemplate, t.Label, nGoroutineStarted)
		// Retrieves a reference to the metadata of the spawned function
		spawnedMeta, existMeta := file.FunctionMeta[t.Label]
		spawnedGrFSA := GoroutineFSA{spawnedName, spawnedMeta}
		// Retrieves a reference to the linearized automaton of the spawned function
		spawnedLin, existLin := inlinedCache[t.Label]

		// IF the automaton doesn't exist we override the transition with an eps one
		if !existMeta || !existLin {
			newT := fsa.Transition{Move: fsa.Eps, Label: "unknown-function-spawn"}
			gr.Automaton.RemoveTransition(from, to, t)
			gr.Automaton.AddTransition(from, to, newT)
			return
		}

		// Updates the Spawn transition with the full name/id of the spawned Goroutine
		newT := fsa.Transition{Move: fsa.Spawn, Label: spawnedName}
		gr.Automaton.RemoveTransition(from, to, t)
		gr.Automaton.AddTransition(from, to, newT)

		// Get a reference to the list of actual arguments and formal ones
		formalArgs := spawnedMeta.InlineArgs
		actualArgs, _ := t.Payload.([]meta.FuncArg)
		// Get a reference to the channels metadata in the caller scope
		channelInfo := gr.ChanMeta

		// Finds and replace transition with subject a formal parameter and replaces
		// them with the same transition but with a reference to the actual argument
		spawnedGrFSA.Automaton = argumentSubstitution(formalArgs, actualArgs, spawnedLin, channelInfo)

		// Extracts recursively the spawn subtree of our spawned and updates the entries in our agglomerate
		for grName, grFSA := range extractSpawnTree(spawnedGrFSA, file) {
			spawnedGoroutines[grName] = grFSA
		}
	})

	// Adds the current GoroutineFSA to the list before returning
	spawnedGoroutines[gr.Name] = &gr
	return spawnedGoroutines
}

// Given the metadata associated to a function linearize the automaton associated to the latter
// by expanding recursively each function call present: The inlining is performed by copying the
// automaton of the "called" function as subgraph to the automaton of the "caller".
// Before inlining formal arguments are replaced by actual ones.
func linearizeFSA(function meta.FuncMetadata, file meta.FileMetadata, cache map[string]*fsa.FSA) {
	// Makes an independent copy that can be freely modified
	copyAutomaton := function.Automaton.Copy()

	copyAutomaton.ForEachTransition(func(from, to int, t fsa.Transition) {
		if t.Move != fsa.Call { // Ignores all non "Call" type transition
			return
		}

		// Checks if the called function has been parsed, some function call such as
		// append() or make() are not "injected" and no metadata is available
		calledMeta, exist := file.FunctionMeta[t.Label]

		// If the function doesn't exist the transition is overwritten with an eps transition
		if !exist {
			newT := fsa.Transition{Move: fsa.Eps, Label: "unknown-function-call"}
			copyAutomaton.RemoveTransition(from, to, t)
			copyAutomaton.AddTransition(from, to, newT)
			return
		}

		// Cache miss: we linearize the called function and we add it to the cache
		// The update of the cache is done by the recursive call
		if cache[t.Label] == nil {
			linearizeFSA(calledMeta, file, cache)
		}

		// Get a reference to the linearized automaton in cache
		calledFuncAutomaton := cache[t.Label]
		// Get a reference to the list of actual arguments and formal ones
		formalArgs := calledMeta.InlineArgs
		actualArgs, _ := t.Payload.([]meta.FuncArg)
		// Get a reference to the channels metadata in the caller scope
		channelInfo := function.ChanMeta

		// Finds and replace transition with subject a formal parameter and replaces
		// them with the same transition but with a reference to the actual argument
		replaced := argumentSubstitution(formalArgs, actualArgs, calledFuncAutomaton, channelInfo)

		// Expands as a subgraph the called function FSA in place of the transition t
		// this process is really similar to function inlining a technique used in compilers
		// to avoid function call overhead and the allocation of an Activation Record
		inlineAutomata(copyAutomaton, from, to, t, replaced)
	})

	// Adds the fully linearized automaton to the cache
	cache[function.Name] = copyAutomaton
}

// Implements the algorithm to replace formal arguments with actual ones.
// Overrides the transition label but also the payload so that future reference to the channel
// will always be correct and successfull
func argumentSubstitution(formal, actual []meta.FuncArg, automaton *fsa.FSA, chanMeta map[string]meta.ChanMetadata) *fsa.FSA {
	// Makes a copy that can be freely modified
	automatonCopy := automaton.Copy()

	// Bails out at the first discrepancy blocking the execution
	if len(formal) != len(actual) {
		log.Fatalf("Couldn't expand arguments: formal %d but actual %d\n", len(formal), len(actual))
	}

	// Expands the actual arguments with the positional ones
	for _, actualArg := range actual {
		for _, funcArg := range formal {
			// Tries to find a match beetwen the actual argument and the positional argument
			if funcArg.Offset != actualArg.Offset || funcArg.Type != actualArg.Type {
				continue
			}

			// If such match is found then all the transition in the automataCopy that references
			// that "formal" argument are replaced with transition to the "actual" argument
			automatonCopy.ForEachTransition(func(from, to int, t fsa.Transition) {
				if funcArg.Type == meta.Channel && t.Label == funcArg.Name && (t.Move == fsa.Recv || t.Move == fsa.Send) {
					// Creates a new transition that will overwrite the old one
					// (the one that references the formal argument)
					newT := fsa.Transition{
						Move:    t.Move,
						Label:   actualArg.Name,
						Payload: chanMeta[actualArg.Name],
					}

					// Replace the transitions
					automatonCopy.RemoveTransition(from, to, t)
					automatonCopy.AddTransition(from, to, newT)
				}

				// ? Handle funcArg.Type == Function as well
			})
		}
	}

	return automatonCopy
}

// This function expands a graph in place of an transition. Since in our case every
// Automata/Graph has only one initial and final state then we simply copy the other graph
// state by state and transition by transition and then we link the copy to the "from" and "to" states
func inlineAutomata(root *fsa.FSA, from, to int, t fsa.Transition, other *fsa.FSA) {
	// First of all remove the old call transition
	root.RemoveTransition(from, to, t)

	// Count the number of states, in order to extract an offset
	offset := 0
	root.ForEachState(func(_ int) { offset++ })

	// Copies the "other" graph state, applying the offset to each id
	other.ForEachTransition(func(from, to int, t fsa.Transition) {
		root.AddTransition(from+offset, to+offset, t)
	})

	// Links the initial state of "other" FSA with the "root" FSA via eps transition
	tExpansionStart := fsa.Transition{Move: fsa.Eps, Label: "start-call-expansion"}
	root.AddTransition(from, offset, tExpansionStart)

	// Links every final/accepting states of the other FSA with the "root" via eps transition
	for _, item := range other.FinalStates.Values() {
		finalStateId := item.(int)
		tExpansionEnd := fsa.Transition{Move: fsa.Eps, Label: "end-call-expansion"}
		root.AddTransition(finalStateId+offset, to, tExpansionEnd)
	}
}
