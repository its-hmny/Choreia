// Copyright 2020 Enea Guidi (hmny). All rights reserved.
// This file are distributed under the General Public License v 3.0.
// A copy of abovesaid license can be found in the LICENSE file.

// This package implements a Finite State Automata (FSA) data structure and its own API.
// For this specific use cases the implementation is quite simple & basic

// The only method available from the outside are FSA and its API
package fsa

import (
	"fmt"
	"log"

	list "github.com/emirpasic/gods/lists/singlylinkedlist"
	set "github.com/emirpasic/gods/sets/hashset"
	"github.com/goccy/go-graphviz"
	"github.com/goccy/go-graphviz/cgraph"
)

const (
	// State.Id error value
	Unknown = -1
	// FSA.AddTransition() default values for "from" and "to"
	NewState = -2
	Current  = -3
)

// ----------------------------------------------------------------------------------------
// FSA

// A FSA is a  directed multigraph that represents all the possible transition from a state
// to the others. In Choreia is used to represent the execution flow of a single function,
// a whole Goroutine or the full choreography of the whole concurrent system
//
// A struct containing a basic graph implementation that keeps track of the transition that
// occurs subsequently during the execution flow of a function (or scope).
type FSA struct {
	currentId   int                          // The last id generated, the id of the last node
	transitions map[int]map[int][]Transition // Adjacency matrix of transition from edge to edge
	FinalStates *list.List                   // A list containing the ids of the final/accepting states
}

// Generates a new empty FSA and returns a pointer reference to it
func New() *FSA {
	newFsa := FSA{
		currentId:   0,
		FinalStates: list.New(),
		// A FSA has always an initial state
		transitions: map[int]map[int][]Transition{0: nil},
	}

	return &newFsa
}

// This function generates an independent copy of the given FSA and returns it
func (original *FSA) Copy() *FSA {
	localCopy := FSA{
		currentId: original.currentId,
		// Get a copy of the value to enforce two completely independent copies
		FinalStates: list.New(original.FinalStates.Values()...),
		transitions: map[int]map[int][]Transition{0: nil},
	}

	// Iterates over the transition in the original FSA, copying them one by one
	original.ForEachTransition(func(from, to int, t Transition) {
		localCopy.AddTransition(from, to, t)
	})

	return &localCopy
}

// Adds a new Transition to the FSA on which is called.
// The user can specify a special flag for the "to" argument and the "from" one
// (respectively NewState and Current) to create a new node as destination of "t"
// or use the last generated node as starting point of "t" itself
func (fsa *FSA) AddTransition(from, to int, t Transition) {
	// Argument checking
	if from == Unknown || to == Unknown {
		log.Fatal("unknown starting or ending state on AddTransition")
	} else if t.Label == "" {
		log.Fatal("empty labels are not allowed")
	}

	// If the user specified the "Current" flag the starting state used is the latest created
	if from == Current {
		from = fsa.currentId
	}

	// If the user specified the "NewState" flag the destination state is created from scratch
	if to == NewState {
		to = fsa.GetLastId() + 1
		fsa.SetRootId(to)
	}

	// If the "nested" map is nil is initialized just before usage
	if fsa.transitions[from] == nil {
		fsa.transitions[from] = make(map[int][]Transition)
	}

	// Avoids adding duplicated transitions
	for _, prevT := range fsa.transitions[from][to] {
		if prevT.Move == t.Move && prevT.Label == t.Label {
			return
		}
	}

	// Adds the new transition in the adjacency matrix
	fsa.transitions[from][to] = append(fsa.transitions[from][to], t)
}

// Removes a transition "from" and "to" the specified states with a matching Move and Label.
// If no such label that fullfills the match criteria is found then the procedure returns
// without provinding any kind of error
func (fsa *FSA) RemoveTransition(from, to int, t Transition) {
	// Argument checking
	if from == Unknown || to == Unknown {
		log.Fatal("unknown starting or ending state on AddTransition")
	} else if t.Label == "" {
		log.Fatal("empty labels are not allowed")
	}

	// Retrieves the current transition list and creates a new one
	oldList := fsa.transitions[from][to]
	newList := make([]Transition, 0, len(oldList))

	// Puts all the non matching transition in the new list, filtering out only the matching one
	for _, transition := range oldList {
		if t.Label != transition.Label || t.Move != transition.Move {
			newList = append(newList, transition)
		}
	}

	// Overwrites the old list with the new (filtered) one in the adjacency matrix
	fsa.transitions[from][to] = newList
}

// Returns the id of the last state generated
func (fsa *FSA) GetLastId() int {
	stateSet := set.New()

	// Populates the state set (duplicate ids are avoided)
	for from, outgoing := range fsa.transitions {
		stateSet.Add(from)
		for to := range outgoing {
			stateSet.Add(to)
		}
	}
	// ! Maybe will be better to use something like stateSet.Max()
	// ! to get the biggest id available
	return stateSet.Size() - 1
}

// Sets the state identified by the given id as the new root of the FSA, this means that the next
// transition added with the "Current" flag will start from this node, this is valid until a new
// state is generated with the NewState flag which, in that case, will override the current root id
func (fsa *FSA) SetRootId(newRootId int) {
	fsa.currentId = newRootId
}

// Allows functional iteration over each transition currently available in the FSA.
// The callback of the user can change and interact with FSA but the changes made will
// not be available in this method since it considers a "frozen" version of the adjency matrix
func (fsa *FSA) ForEachTransition(callback func(from, to int, t Transition)) {
	// Iterates over each state in the adjacency matrix
	for from, outgointTransitions := range fsa.transitions {
		// Iterates over each outgoing transitions for the abovesaid state
		for to, parallelTransitions := range outgointTransitions {
			// Iterates over each parallel transition (with same start and ending state)
			for _, t := range parallelTransitions {
				callback(from, to, t)
			}
		}
	}
}

// Allows functional iteration over each state currently available in the FSA.
// The callback of the user can change and interact with FSA but the changes made will
// not be available in this method since it considers a "frozen" version of the adjency matrix
func (fsa *FSA) ForEachState(callback func(id int)) {
	stateSet := set.New()

	// Populates the state set (duplicate ids are avoided)
	for from, outgoing := range fsa.transitions {
		stateSet.Add(from)
		for to := range outgoing {
			stateSet.Add(to)
		}
	}

	// Iterate on the set with only unique values
	for stateId := range stateSet.Values() {
		callback(stateId)
	}
}

// Exports the referenced FSA to a given path and in the given format/encoding.
// Some supported encoding/format are: SVG, PNG, DOT, etc... The funcion doesn't
// do any check about the given path and wil straight up fail if the path is invalid
// or it will overwrite the current file saved at that location
func (fsa *FSA) Export(outputFile string, format graphviz.Format) {
	// Creates a GraphViz instance and initializes a Graph render object
	gvInstance := graphviz.New()
	graph, graphErr := gvInstance.Graph()

	// Cleanup function that closes both the Graph and GraphViz instances
	// in case of any error during execution or after the execution completed successfully
	defer func() {
		if err := graph.Close(); err != nil {
			log.Fatal(err)
		}
		gvInstance.Close()
	}()

	if graphErr != nil {
		log.Fatal(graphErr)
	}

	// A simple conversion map to keep track of the cross references
	// (FSA => graphviz.Graph) between states and nodes
	state2node := make(map[int]*cgraph.Node)

	// Bulk copy of states from the FSA to the graphviz Graph (as nodes)
	fsa.ForEachState(func(stateId int) {
		// Creates a cgraph.Node from the current stateId
		node, nodeErr := graph.CreateNode(fmt.Sprint(stateId))
		node.SetShape(cgraph.CircleShape) // Default shape

		if nodeErr != nil {
			log.Fatal(nodeErr)
		}

		// If the current state is final state then changes the shape
		if fsa.FinalStates.Contains(stateId) {
			node.SetShape(cgraph.DoubleCircleShape)
		}

		// At last updates the association map with the new entries
		state2node[stateId] = node
	})

	// Bulk copy of transitions from the FSA to the graphviz Graph (as edges)
	fsa.ForEachState(func(startId int) {
		for destId, parallelT := range fsa.transitions[startId] {
			// Retrieves the references to the graphviz.Graph nodes
			fromRef, toRef := state2node[startId], state2node[destId]
			// Creates a uid for the current edge from the tuple (from, to, t)
			edgeId := fmt.Sprintf("%d-%d", startId, destId)
			edgeLabel := ""

			// Since Graphviz doesn't support parallel edges we implement it ourselves
			// by "squashing" all parallel transitions into one singe label "\n" separated
			for _, t := range parallelT {
				edgeLabel += fmt.Sprintf("\n%s", t)
			}

			// Creates the edge and sets its label
			edge, edgeErr := graph.CreateEdge(edgeId, fromRef, toRef)
			edge.SetLabel(edgeLabel)

			if edgeErr != nil {
				log.Fatal(edgeErr)
			}
		}
	})

	// Creates an export in the format requested at the given path
	exportErr := gvInstance.RenderFilename(graph, format, outputFile)

	if exportErr != nil {
		log.Fatal(exportErr)
	}
}
