// Copyright Enea Guidi (hmny).

// TODO COMMENT

// TODO comment
package transforms

import (
	list "github.com/emirpasic/gods/lists/singlylinkedlist"
	set "github.com/emirpasic/gods/sets/hashset"

	"github.com/its-hmny/Choreia/internal/data_structures/fsa"
)

// TODO COMMENT
// TODO COMMENT
// TODO COMMENT
func newEpsClosure(automata *fsa.FSA, states *set.Set) *set.Set {
	// A set to keep track of all the states already reached
	reachedStates := set.New(states.Values()...)

	automata.ForEachTransition(func(from, to int, t fsa.Transition) {
		// If the current is a eps transition starting from one of the already reachable states
		// we add the destination of said transition to the eps closure
		if t.Move == fsa.Eps && reachedStates.Contains(from) {
			reachedStates.Add(to)
		}

	})

	// If we've reached more states than the previous call we search recursively
	if reachedStates.Size() > states.Size() {
		return newEpsClosure(automata, reachedStates)
	}

	// Else we found all the states reachable and we return the full closure
	return reachedStates
}

// TODO comment
// TODO comment
// TODO comment
func subsetConstruction(NCA *fsa.FSA) *fsa.FSA {
	DCA := fsa.New() // The deterministic version of the NCA

	// Initialization of the eps-closure of the first state,
	initialClosure := newEpsClosure(NCA, set.New(0))
	//Init the tSet (a set of eps-closure)
	tSet := list.New(initialClosure)
	// Init the nIteration counter that will be used to iterate over tSet
	nIteration := 0

	// We use this trick since the range statement uses a "frozen" version of the variable
	// while we need a "live" value
	for nIteration < tSet.Size() {
		item, _ := tSet.Get(nIteration) // Extracts the closure to be evaluated
		closure := item.(*set.Set)

		NCA.ForEachTransition(func(from, to int, t fsa.Transition) {
			if !closure.Contains(from) { // Skips the transition that don't start from within the closure
				return
			}

			// Extracts the states that can be reached from the eps closure with transition t
			// then calculates the aggregate eps-closure of these reachable states
			moveEpsClosure := getReachable(NCA, closure, t)

			// Ignores empty eps-closure when empty this means that the transition function is not defined
			if moveEpsClosure.Size() <= 0 {
				return
			}

			// TODO Handle final states in the closure

			// If the eps-closure extracted already exist in tSet (has been already disvocered)
			// then retrieves its twin's id from the map, and use the latter instead of its twin
			twinIndex, twinId := tSet.Find(func(_ int, item interface{}) bool {
				c := item.(*set.Set)
				// Simple tricK: If A is contained in B and viceversa then A equals B
				isAContained := c.Contains(moveEpsClosure.Values()...)
				isBContained := moveEpsClosure.Contains(c.Values()...)
				return isAContained && isBContained
			})

			// If a twin closure exist its index is used to link the states with t, else a new state is used
			if twinId == nil {
				tSet.Add(moveEpsClosure)
				DCA.AddTransition(nIteration, fsa.NewState, t)
			} else {
				DCA.AddTransition(nIteration, twinIndex, t)
			}
		})

		nIteration++
	}

	return DCA
}

// TODO comment
// TODO comment
// TODO comment
func getReachable(automata *fsa.FSA, clos *set.Set, move fsa.Transition) *set.Set {
	// Init an empty list of states reachable
	tReachable := set.New()

	automata.ForEachTransition(func(from, to int, t fsa.Transition) {
		if move.Move == t.Move && move.Label == t.Label && clos.Contains(from) {
			tReachable.Add(to)
		}
	})

	// Return the epsClosure of the reachable states
	return newEpsClosure(automata, tReachable)
}