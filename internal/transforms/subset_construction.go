// Copyright Enea Guidi (hmny).

// TODO COMMENT

// TODO comment
package transforms

import (
	set "github.com/emirpasic/gods/sets/hashset"

	"github.com/its-hmny/Choreia/internal/data_structures/closure"
	"github.com/its-hmny/Choreia/internal/data_structures/fsa"
)

type epsClosure struct {
}

// TODO comment
// TODO comment
// TODO comment
func subsetConstruction(NCA *fsa.FSA) *fsa.FSA {
	DCA := fsa.New() // The deterministic version of the NCA

	// Initialization of the eps-closure of the first state,
	initialClosure := epsClosure.New(NCA, 0)
	//Init the tSet (a set of eps-closure)
	tSet := set.New()
	// Init the nIteration counter that will be used to iterate over tSet
	nIteration := 0

	// We use this trick since the range statement uses a "frozen" version of the variable
	// while we need
	// variable while we need a live value
	for nIteration < len(tSet) {
		closure := tSet[nIteration] // Extracts the closure to be evaluated

		closure.ForEachTransition(func(from, to int, t fsa.Transition) {
			// Extract the states that can be reached from the eps closure with transition t
			// then calculate the aggregate eps-closure of these reachable states
			moveEpsClosure := getReachable(NCA, closure, t)

			// Ignores empty eps-closure when empty this means that the transition function is not defined
			if moveEpsClosure.IsEmpty() {
				return
			}

			// TODO Handle final states in the closure

			// If the eps-closure extracted already exist in tSet (has been already disvocered)
			// then retrieves its twin's id from the map, and use the latter instead of its twin
			exist, twinId := isContained(moveEpsClosure, tSet)

			if !exist {
				// If it's not a member of tSet then it's added to it
				tSet = append(tSet, moveEpsClosure)
				// And it's added a new state (+ transition) to the equivalent DCA
				DCA.AddTransition(idMap[closure.Id], fsa.NewState, t)
				idMap[moveEpsClosure.Id] = DCA.GetLastId()
			} else {
				// Else only a transition its added to the already present state (the twin eps-closure)
				DCA.AddTransition(idMap[closure.Id], idMap[twinId], t)
			}

		})

		nIteration++
	}

	return DCA
}

// TODO comment
// TODO comment
// TODO comment
func getReachable(automata *fsa.FSA, clos *closure.Closure, move fsa.Transition) epsClosure {
	// Init an empty list of states reachable
	reachableStateIds := []int{}

	// Populates the list with all the state reachable from the closure with the given move
	clos.ForEachTransition((func(from, to int, t fsa.Transition) {
		if move.Move == t.Move && move.Label == t.Label {
			reachableStateIds = append(reachableStateIds, to)
		}
	}))

	// Return the epsClosure of the reachable states
	return getEpsClosure(automata, nil, reachableStateIds...)
}
