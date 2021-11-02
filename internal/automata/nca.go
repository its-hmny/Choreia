// Copyright Enea Guidi (hmny).

// This package handles the extraction of Partial Nondeterministic Automatas from
// metadata extracted and the handling and subsequent transformation of abovesaid
// NCA until a single Deterministic Choreography Automata is obtained by them

// This module defines some helper function to transform and work with NCAs
// (Nondeterministic Choreography Automatas). Such transformations could be extracting
// the NCAs from given metadata or removing Eps transition from them (making them DCAs)
package automata

import (
	"log"

	"github.com/its-hmny/Choreia/internal/meta"
	"github.com/its-hmny/Choreia/internal/types/fsa"
)

// This function extracts from the given function metadata a Partial Nondeterministic CA that
// represents the execution flow of that GoRoutine. When a Spawn transaction is encountered the
// function call itself recursively generating more Partial NCAs for the spawned GoRoutine.
// NOTE: This function should be called with the metadata of a function that is the entrypoint of one
// or more GoRoutine (the function called on a the spawned routine).
func extractPartialNCAs(funcMeta meta.FuncMetadata, fileMeta meta.FileMetadata) []ChoregoraphyAutomata {
	// List of Partial/Projection NCA extracted from the current recurive call
	localCopy := *funcMeta.ScopeAutomata // TODO Add deep copy of SccopeAutoamata
	extractedNCAs := []ChoregoraphyAutomata{&localCopy}

	// ! refactor this code

	// Executes a function on each Transition (edge) of the Graph
	for _, state := range localCopy.StateIterator() {
		for to, t := range state.TransitionIterator() {
			if t.Move == fsa.Call {
				calleeMeta, hasMeta := fileMeta.FunctionMeta[t.Label]
				if hasMeta { // Expands in place the ScopeAutomata of the called function
					localCopy.ExpandInPlace(state.Id, to, *calleeMeta.ScopeAutomata)
				} else { // Transforms the transition in an eps-transition (that later will be removed)
					newT := fsa.Transition{Move: fsa.Eps, Label: "unknown-fuction-call"}
					localCopy.AddTransition(state.Id, to, newT) // Overwrites the current one
				}
			} else if t.Move == fsa.Spawn {
				calledFuncMeta, hasMeta := fileMeta.FunctionMeta[t.Label]
				if hasMeta {
					// Recurively call extractPartialNCAs on the spawned GoRoutine entrypoint (the function
					// scalled with go keyword), then add the extracted NCAs to the current list
					extractedNCAs = append(extractedNCAs, extractPartialNCAs(calledFuncMeta, fileMeta)...)
				} else { // Transforms the transition in an eps-transition (that later will be removed)
					log.Fatalf("Couldn't find function %s spawned as Go Routine\n", t.Label)
				}
			}
		}
	}

	return extractedNCAs
}

// TODO implement
func removeEpsTransitions(NCA ChoregoraphyAutomata) ChoregoraphyAutomata {
	return nil
}
