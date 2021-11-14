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

// This function extracts from the given function metadata a Projection DCA that
// represents the execution flow of a GoRoutine. When a Spawn transaction is encountered the
// function call itself recursively generating more Projection DCAs for the spawned GoRoutine.
// NOTE: This function should be called with the metadata of a function that is the entrypoint of one
// or more GoRoutine (the function called on a the spawned routine).
func extractProjectionDCAs(funcMeta meta.FuncMetadata, fileMeta meta.FileMetadata) []*fsa.FSA {
	// Makes a full indipendent copy of the ScopeAutomata
	localCopy := funcMeta.ScopeAutomata.Copy()
	// List of Projection DCA extracted from the current recursive call,
	// the first position is reserved to the currently evaluated ScopeAutomata
	// but it will be inserted at last
	extractedNDCAs := []*fsa.FSA{nil}

	// Executes the following  on each Transition (edge) of the Graph
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
					newNDCAs := extractProjectionDCAs(calledFuncMeta, fileMeta)
					extractedNDCAs = append(extractedNDCAs, newNDCAs...)
					// Overrides the older transtion with additional data
					newT := fsa.Transition{Move: fsa.Spawn, Label: t.Label, Payload: newNDCAs[0]}
					localCopy.AddTransition(state.Id, to, newT)
				} else {
					// Exit with errror since we cannot determine the final Choreography correctly
					log.Fatalf("Couldn't find function %s spawned as Go Routine\n", t.Label)
				}
			}
		}
	}

	// Sets the first place in the slice to the deterministic version of the
	// currently evaluated ScopeAutomata that now represents a Projection Automata
	// (the flow of a whole GoRoutine and not only a scope)
	extractedNDCAs[0] = getDeterministicForm(localCopy)

	return extractedNDCAs
}

// Removes eps-transition from a given NDCA (Non-Deterministic Choreography Automata) transforming it
// to an equivalent deterministic form, obtaining, in fact, a DCA (Deterministic Choreography Automata)
// Abovesaid DCA is then returned to the caller, the 2 instance are completely sepratated
func getDeterministicForm(NDCA *fsa.FSA) *fsa.FSA {
	DCA := fsa.New()           // The deterministic DCA
	idMap := make(map[int]int) // To map the id of the closures to the id of the FSA's states

	// Initialization of some basic fields, such as the eps-closure of the first state,
	// the tSet (a set of eps-slosure) and the nIteration counter that will be used to iterate
	// over tSet without using the "range" construct that uses a "frozen" copy of the iteratee
	// and doesn't allow for everchanging value
	firstEpsClosure := getEpsClosure(NDCA, nil, NDCA.GetState(0))
	tSet := []*closure.Closure{firstEpsClosure}
	nIteration := 0

	// This type of iteration is necessary since the "range" one will not work
	// on a struct that changes inside a loop, some element are not guaranteed to be iterated
	for nIteration < len(tSet) {
		closure := tSet[nIteration] // Extracts the closure to be evaluated
		// ! Debug only, will be removed later
		closure.ExportAsSVG(fmt.Sprintf("debug/eps-closure--%p-%d.svg", NDCA, closure.Id))

		for _, possibleTransition := range closure.TransitionIterator() {
			// Extract the states that can be reached from the eps closure with transition t
			// then calculate the aggregate eps-closure of these reachable states
			reachedByMove := reachWithMove(possibleTransition, closure, NDCA)
			moveEpsClosure := getEpsClosure(NDCA, nil, reachedByMove...)

			// Ignores the error state (empty eps-closure), from which is not possible to escape
			// this state doen't provide any information about the automata and "bloats" the representation
			if moveEpsClosure.IsEmpty() {
				continue
			}

			// If the eps-closure extracted already exist in tSet (has been already found)
			// then retrieves its twin's id from the map, and use the last instead of the id assigned
			exist, twinId := isContained(moveEpsClosure, tSet)

			if !exist {
				// If it's not a member of tSet then it's added to it
				tSet = append(tSet, moveEpsClosure)
				// And it's added a new state (+ transition) to the equivalent DCA
				DCA.AddTransition(idMap[closure.Id], fsa.NewState, possibleTransition)
				idMap[moveEpsClosure.Id] = DCA.GetLastId()
			} else {
				// Else only a transition its added to the already present state (the twin eps-closure)
				DCA.AddTransition(idMap[closure.Id], idMap[twinId], possibleTransition)
			}
		}

		nIteration++
	}

	return DCA
}

// Given one (or more states) and the FSA to which said states belong to, extracts the aggregate eps-closure
// of the states in list, recursively, and returns it to the caller. The "prevClosure" argument is used internally
// by the function to avoid cyclcing on states already visited as well as avoiding recursive infinite loop on
// cyclic transition (from x to x itself). This argument should be nil when calling the function from outside
func getEpsClosure(NDCA *fsa.FSA, prevClosure *closure.Closure, states ...fsa.State) *closure.Closure {
	// If a the prevClosure is nil then an aggragate one is created and used
	if prevClosure == nil {
		prevClosure = closure.New()
	}

	for _, state := range states {
		prevClosure.Add(state) // A state always belong to its epsClosure

		// For each state reachable with an eps transition from this one we calculate
		// its own epsClosure and we merge the two together
		for destId, transition := range state.TransitionIterator() {
			// Skips non-eps transition or destination already included
			if transition.Move != fsa.Eps || prevClosure.Exist(destId) {
				continue
			}
			// Get the eps-closure of the eps-reached state and adds it to the "aggregate" closure
			reachedState := NDCA.GetState(destId)
			reachedEpsClosure := getEpsClosure(NDCA, prevClosure, reachedState)
			prevClosure.Add(reachedEpsClosure.Iterator()...)
		}
	}

	return prevClosure
}

// Given a transition t, a closure set and the finite state automata to which said
// closure and transition belong returns the list of reachable states with that transition
// from the closure
func reachWithMove(t fsa.Transition, closureSet *closure.Closure, automata *fsa.FSA) []fsa.State {
	stateList := []fsa.State{}

	for _, state := range closureSet.Iterator() {
		for destId, transition := range state.TransitionIterator() {
			if transition.Move == t.Move && transition.Label == t.Label {
				stateList = append(stateList, automata.GetState(destId))
			}
		}
	}

	return stateList
}

// Simply checks that the given "item" closure exist in the list provided
// if a match is found, returns the id of the "twin" closure else return -1
func isContained(item *closure.Closure, set []*closure.Closure) (bool, int) {
	for _, element := range set {
		if element.IsEqual(item) {
			return true, element.Id
		}
	}

	return false, -1
}
