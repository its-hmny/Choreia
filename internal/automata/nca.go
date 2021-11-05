// Copyright Enea Guidi (hmny).

// This package handles the extraction of Partial Nondeterministic Automatas from
// metadata extracted and the handling and subsequent transformation of abovesaid
// NDCA until a single Deterministic Choreography Automata is obtained by them

// This module defines some helper function to transform and work with NDCAs
// (Non-Deterministic Choreography Automatas). Such transformations could be extracting
// the NDCAs from given metadata or removing Eps transition from them (making them DCAs)
package automata

import (
	"fmt"
	"log"

	"github.com/its-hmny/Choreia/internal/meta"
	"github.com/its-hmny/Choreia/internal/types/closure"
	"github.com/its-hmny/Choreia/internal/types/fsa"
)

// This function extracts from the given function metadata a Projection NDCA that
// represents the execution flow of a GoRoutine. When a Spawn transaction is encountered the
// function call itself recursively generating more Projection NDCAs for the spawned GoRoutine.
// NOTE: This function should be called with the metadata of a function that is the entrypoint of one
// or more GoRoutine (the function called on a the spawned routine).
func extractProjectionNDCAs(funcMeta meta.FuncMetadata, fileMeta meta.FileMetadata) []*fsa.FSA {
	// Makes a full indipendent copy of the ScopeAutomata
	localCopy := funcMeta.ScopeAutomata.Copy()
	// List of Projection NDCA extracted from the current recurive call
	extractedNDCAs := []*fsa.FSA{localCopy}

	// ! Debug print, will be removed
	fmt.Printf("Local copy of '%s' ScopeAutomata at %p, other at %p\n", funcMeta.Name, localCopy, funcMeta.ScopeAutomata)

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
					// Recurively call extractProjectionNDCAs on the spawned GoRoutine entrypoint (the function
					// scalled with go keyword), then add the extracted NDCAs to the current list
					extractedNDCAs = append(extractedNDCAs, extractProjectionNDCAs(calledFuncMeta, fileMeta)...)
				} else { // Transforms the transition in an eps-transition (that later will be removed)
					log.Fatalf("Couldn't find function %s spawned as Go Routine\n", t.Label)
				}
			}
		}
	}

	return extractedNDCAs
}

// TODO implement
func subsetConstructionAlgorithm(NDCA *fsa.FSA) fsa.FSA {
	sigmaAlphabet := []fsa.MoveKind{fsa.Recv, fsa.Send, fsa.Spawn}
	firstEpsClosure := getEpsClosure(NDCA, NDCA.GetState(0))
	tSet := []*closure.Closure{firstEpsClosure}
	nIteration := 0

	for nIteration < len(tSet) {
		closure := tSet[nIteration] // Extracts the closure to be evaluated
		closure.ExportAsSVG(fmt.Sprintf("debug/eps-closure-%d.svg", closure.Id))

		for _, possibleMove := range sigmaAlphabet {
			reachedByMove := reachWithMove(possibleMove, closure, NDCA)
			moveEpsClosure := getEpsClosure(NDCA, reachedByMove...)
			moveEpsClosure.Id++
			if !isContained(moveEpsClosure, tSet) {
				tSet = append(tSet, moveEpsClosure)
			}
		}

		nIteration++
	}

	return fsa.FSA{}
}

func getEpsClosure(NDCA *fsa.FSA, states ...fsa.State) *closure.Closure {
	epsClosure := closure.New()

	for _, state := range states {
		epsClosure.Add(state) // A state always belong to its epsClosure

		// For each state reachable with an eps transition from this one we calculate
		// its own epsClosure and we merge the two together
		for destId, transition := range state.TransitionIterator() {
			// Skips non-eps transition or destination already included
			if transition.Move != fsa.Eps || epsClosure.Exist(destId) {
				continue
			}
			reachedState := NDCA.GetState(destId)
			reachedEpsClosure := getEpsClosure(NDCA, reachedState)
			epsClosure.Add(reachedEpsClosure.Iterator()...)
		}
	}

	return epsClosure
}

func reachWithMove(move fsa.MoveKind, closureSet *closure.Closure, NDCA *fsa.FSA) []fsa.State {
	stateList := []fsa.State{}

	for _, state := range closureSet.Iterator() {
		for destId, transition := range state.TransitionIterator() {
			if transition.Move == move {
				stateList = append(stateList, NDCA.GetState(destId))
			}
		}
	}

	return stateList
}

func isContained(item *closure.Closure, set []*closure.Closure) bool {
	for _, element := range set {
		if element.IsEqual(item) {
			return true
		}
	}

	return false
}
