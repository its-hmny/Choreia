// Copyright 2020 Enea Guidi (hmny). All rights reserved.
// This file are distributed under the General Public License v 3.0.
// A copy of abovesaid license can be found in the LICENSE file.

// Package transforms declares the types and functions used to transform and work with some type of FSA.
// Come of the transformation implemented here are standard such as determinization (Subset Construction),
// minimization but more are specifically related to Choreia (GoroutineFSA extraction & Composition)
//
package transforms

import (
	"log"

	"github.com/its-hmny/Choreia/internal/data_structures/fsa"
	meta "github.com/its-hmny/Choreia/internal/static_analysis"
)

var nProjectionExtracted = 0

const nameTemplate = "%s (%d)"

// -------------------------------------------------------------------------------------------
// GoroutineFSA

// A FSA that represents the execution flow of a single Goroutine (identified by its own name)
// this will be used in the future phases as a local view for the whole choreography.
// This means that this FSA will provide an "isolated" view of the choreography from the perspective
// of the Goroutine that takes part in it (almost like a projection of the whole choreography)
type GoroutineFSA struct {
	Name      string   // An identifier for the Automata
	Automaton *fsa.FSA // The FSA itself
}

func ExtractGoroutineFSA(file meta.FileMetadata) map[string]*GoroutineFSA {
	inlinedCache := make(map[string]*fsa.FSA)

	for _, function := range file.FunctionMeta {
		// Cache hit: The current automaton has already been linearized.
		if inlinedCache[function.Name] != nil {
			// This means its function calls in the automaton have been already inlined and the latter
			// contains (as subgraphs) the FSA of the function called (just like compiler inlining)
			continue
		}

		// Cache miss: We must linearize the current automaton
		linearizeFSA(function, file, inlinedCache)
	}

	// ToDO: Complete the second phase
	// Extracts all the GoroutineFSA starting from the "main" function
	// which is the entrypoint for the Go program
	return make(map[string]*GoroutineFSA)
}

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

/*
! This all need to be refactored
// Extracts the Projection Automata (or "local view") for the given FuncMeta.
// The ScopeAutomata of said function is used as entry point of the Projection CA.
// Every call to another function is inlined in the local view.
// When a new GoRoutine is spawned during the execution flow a new local view is generated.
// Both Call and Spawn operation require expansion of formal arguments with actual ones
func GetLocalViews(function meta.FuncMetadata, file meta.FileMetadata) map[string]*GoroutineFSA {
	// Creates the Projection Automata for the current GoRoutine
	current := GoroutineFSA{
		Name:      fmt.Sprintf(nameTemplate, function.Name, nProjectionExtracted),
		Automaton: function.ScopeAutomata.Copy(), // Makes a full independent copy of the funcMeta.ScopeAutomata
	}

	nProjectionExtracted++
	extractedList := map[string]*GoroutineFSA{current.Name: &current}

	// Iterates over each transition in the ScopeAutomata
	function.ScopeAutomata.ForEachTransition(func(from, to int, t fsa.Transition) {
		switch t.Move {
		case fsa.Call:
			inlineCallTransition(file, current.Automaton, from, to, t)
		case fsa.Spawn:
			for key, pAutomata := range extractSpawnTransition(file, current.Automaton, from, to, t) {
				extractedList[key] = pAutomata
			}
		}
	})

	return extractedList
}

// Handles the Call transition in a local view (or Projection Automata), the call transition is removed
// and the Scope Automata of the called function is inlined at her place (once the actual arguments are expanded)
func inlineCallTransition(file meta.FileMetadata, root *fsa.FSA, from, to int, t fsa.Transition) {
	// Tries to retrieve the called function metadata from the file
	calledFunc, hasMeta := file.FunctionMeta[t.Label]

	// Could not find metadata (probably built-in or external function such as "make" or "append")
	// the call transition is replaced with an eps-transition
	if !hasMeta {
		newT := fsa.Transition{Move: fsa.Eps, Label: "unknown-function-call"}
		root.RemoveTransition(from, to, t)
		root.AddTransition(from, to, newT)
		return
	}

	// Expands positional argument (Channel e Function/Callback) with the actual arguments provided
	// by the caller (Find and Replace over the called ScopeAutomata) then inline the called ScopeAutomata
	// into the root one
	calledScopeAutomata := replaceActualArgs(t, calledFunc)
	inlineAutomata(root, from, to, t, calledScopeAutomata)
}

// Hansles the Spawn transition in a local view (or Projection Automata), the local view of the newly spawned
// is extracted with eventually the his "child" Goroutine and then the transition is updated with a reference
// to the ProjectionAutomata struct of the spawned Goroutine
func extractSpawnTransition(file meta.FileMetadata, root *fsa.FSA, from, to int, t fsa.Transition) map[string]*GoroutineFSA {
	// Tries to retrieve the called function metadata from the file
	calledFunc, hasMeta := file.FunctionMeta[t.Label]

	// Could not find metadata (probably built-in or external function such as "make" or "append")
	// the Choreography cannot be determined correctly with a missing actor so the program fails
	if !hasMeta {
		log.Fatalf("Couldn't find function %s spawned as Go Routine\n", t.Label)
	}

	// Expands positional arguments with the actual ones
	calledFunc.ScopeAutomata = replaceActualArgs(t, calledFunc)
	// Precalculates the key used for retrieval of the first local view extracted
	calledFuncKey := fmt.Sprintf(nameTemplate, calledFunc.Name, nProjectionExtracted)

	// Recursively call getProjectionAutomata on the spawned GoRoutine entrypoint
	//(the function called with go keyword), then returns the extracted Projection Automata
	// the first is always the spawned one, the others can be function spawned by the one spawned by us
	newLocalAutomata := GetLocalViews(calledFunc, file)
	// Retrieves a reference to the first local view
	localViewRef := newLocalAutomata[calledFuncKey]

	// Overrides the older transition with additional data (reference to the spawned ProjectionAutomata)
	newT := fsa.Transition{Move: t.Move, Label: localViewRef.Name, Payload: localViewRef}
	root.RemoveTransition(from, to, t)
	root.AddTransition(from, to, newT)

	return newLocalAutomata
}

// Expand positional argument (Channel e Function/Callback) with the actual ones so that any
// Transition reference the actual channel at runtime. The function uses the offset of the arguments
// and their types to determine which positional argument has to be replaced with the "actual" ones
func replaceActualArgs(t fsa.Transition, calledFunc meta.FuncMetadata) *fsa.FSA {
	// Copies the ScopeAutomata of the called function
	calledAutomataCp := calledFunc.ScopeAutomata.Copy()
	// Retrieves the actual arguments from the Call transition
	expandArgs, _ := t.Payload.([]meta.FuncArg)

	// Bails out at the first discrepancy returning a non-expanded copy
	if len(expandArgs) != len(calledFunc.InlineArgs) {
		log.Fatalf("Could not expand actual arguments: requested %d but given %d\n", len(calledFunc.InlineArgs), len(expandArgs))
	}

	// Expands the actual arguments with the positional ones
	for _, actualArg := range expandArgs {
		for _, funcArg := range calledFunc.InlineArgs {
			// Tries to find a match beetwen the actual argument and the positional argument
			if funcArg.Offset != actualArg.Offset || funcArg.Type != actualArg.Type {
				continue
			}

			// If such match is found then all the transition in the localCopy that references that
			// positional argument are replaced with transition to the actual argument
			calledAutomataCp.ForEachTransition(func(from, to int, t fsa.Transition) {
				if funcArg.Type == meta.Channel && t.Label == funcArg.Name && (t.Move == fsa.Recv || t.Move == fsa.Send) {
					newT := fsa.Transition{Move: t.Move, Label: actualArg.Name}
					// Replace the transitions in the copied ScopeAutomata
					fmt.Println(from, to, newT)
					calledAutomataCp.RemoveTransition(from, to, t)
					calledAutomataCp.AddTransition(from, to, newT)
				}
				// ? Handle funcArg.Type == Function as well
			})
		}
	}

	return calledAutomataCp
}*/
