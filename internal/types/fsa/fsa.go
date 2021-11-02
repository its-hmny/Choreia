// Copyright Enea Guidi (hmny).

// This package implements a Finite State Automata (FSA) data structure and its own API.
// For this specific use cases the implementation is quite simple & basic

// The only method avaiable from the outside are Transitiongraph and its API
package fsa

import (
	"fmt"
	"log"

	"github.com/goccy/go-graphviz"
	"github.com/goccy/go-graphviz/cgraph"
)

const (
	// FSA.AddTransition() default values for "from" and "to"
	NewState = -2
	Current  = -3
	// State.Id error value
	Unknown = -1
)

// ----------------------------------------------------------------------------
// TransitionGraph

// A FSA is a graph that represents all the possible transition from a state to the others
//
// A struct containing a basic graph implementation that keeps track of the transaction that occur
// subsequently during the execution flow of a function (or scope).
type FSA struct {
	currentId int     // The last id generated, the id of the last node
	states    []State // The list of state inside the graph
}

type State struct {
	Id         int                // The id of the current state
	transition map[int]Transition // A map to other stateId the Transition data
}

// This function generates a new FSA and returns a pointer reference to it
func NewFSA() *FSA {
	return &FSA{
		currentId: 0,
		states: []State{
			// Every FSA has already the first (0) state inside
			{Id: 0, transition: make(map[int]Transition)},
		},
	}
}

// This function generates an indipendet copy of the given FSA and returns it
func (original *FSA) Copy() *FSA {
	copy := FSA{
		currentId: original.currentId,
		states:    make([]State, 0, len(original.states)),
	}

	for _, currenState := range original.states {
		// Creates a new map, overriding the copied one (since map instance remains intertwined)
		currenState.transition = make(map[int]Transition)
		// Copies all the entry of the other map into the decoupled one
		for dest, t := range original.states[currenState.Id].transition {
			currenState.transition[dest] = t
		}
		// Then adds the fully copied state to the copy FSA
		copy.states = append(copy.states, currenState)
	}

	return &copy
}

// Returns the latest generated id
func (fsa *FSA) GetLastId() int {
	return len(fsa.states) - 1
}

// Set a new rootId, a rootId is the id of the State (state) from which all future transition
// will start when an id isn't specified, this is used since when merging multiple subgraph is
// needed that the merge state will become the one from which create new transition even if it
// is not the last created state
func (fsa *FSA) SetRootId(newRootId int) {
	fsa.currentId = newRootId
}

// Returns the id of the final state, if such state is not present returns Unknown
func (fsa *FSA) GetFinalStateId() int {
	for _, currentState := range fsa.states {
		// The final state is the one for which there aren't any outcoming transitions
		if len(currentState.transition) == 0 {
			return currentState.Id
		}
	}

	return Unknown
}

// This function adds a new State to the TransitionGraph generating its
// id incrementally with respects to the previusly existent state
func (fsa *FSA) newState() (id int) {
	id = len(fsa.states) // Generates a new id
	fsa.states = append(fsa.states, State{
		Id:         id,
		transition: make(map[int]Transition),
	})
	return id
}

// This function adds a new Transition and its payload the user can specify the
// from and to states or eventually can use some special value such as Current
// for "from" that connect the new state to the latest or newStates for "to"
// that automatically instantiate a new state as destination of the transition
// NOTE: If a transition already exist then is overwritten
func (fsa *FSA) AddTransition(from, to int, t Transition) {
	// Input checking
	if from == Unknown || to == Unknown {
		log.Fatal("unknown starting or ending state on AddTransition")
	} else if t.Label == "" {
		return
	}

	// The user can omit a specific starting state, in this case the
	// latest added state id is used as starting point of the transition
	if from == Current {
		from = fsa.currentId
	}

	// The user can omit the ending state of the new transition, in this
	// case a new state is created and the transition is linked to that one
	if to == NewState {
		to = fsa.newState()
		fsa.SetRootId(to)
	}

	// ! Debug print, will be removed later
	fmt.Printf("BP__ %d -> %d \t %+v\n", from, to, t)

	// Creates/assign the new transition
	fsa.states[from].transition[to] = t
}

// This function expands a graph in place of an transition. Since in our case every
// Automata/Graph has only one initial and final state then we simply copy the other graph
// state by state and transition by transition and then we link the copy to the "from" and "to" states
func (fsa *FSA) ExpandInPlace(from, to int, other FSA) {
	// First of all remove the old call transition (since it will be expanded)
	delete(fsa.states[from].transition, to)
	// Calculate the offset from which the ids of the "other" will be padded with
	offset := len(fsa.states)

	// Copies the "other" graph state, applying the offset to each id
	for _, cpState := range other.states {
		// Creates a new transition map for and copies all the transitions to it
		newStateTrans := make(map[int]Transition)
		for dest, t := range cpState.transition {
			newDest := offset + dest
			newStateTrans[newDest] = t
		}
		// Then creates a new state and adds it to the destination graph
		newState := State{Id: offset + cpState.Id, transition: newStateTrans}
		fsa.states = append(fsa.states, newState)
	}

	// Eps transition to mark/link the start of the expanded/copied graph
	tExpansionStart := Transition{Move: Eps, Label: "start-call-expansion"}
	tExpansionEnd := Transition{Move: Eps, Label: "end-call-expansion"}
	// Initial and final state of the freshly copied graph
	copyInitialStateId, copyFinalStateId := offset, other.GetFinalStateId()+offset
	// The "linking" transition are now added to the graph, completing the expansion
	fsa.AddTransition(from, copyInitialStateId, tExpansionStart)
	fsa.AddTransition(copyFinalStateId, to, tExpansionEnd)
}

// Returns an iterable representation of the outcoming transition for the given State
func (s *State) TransitionIterator() map[int]Transition {
	return s.transition
}

// Returns an iterable representation of the states for the given Graph
func (fsa *FSA) StateIterator() []State {
	return fsa.states
}

// This function exports a .png image of the current state of the Graph, it copies state by state
// and then transition by transition the graph upon which is called, and then saves the GraphViz copy as
// a .png image file to the provided path
func (fsa *FSA) ExportAsSVG(imagePath string) {
	// Creates a GraphViz instance and initializes a Graph instance
	graphvizInstance := graphviz.New()
	graphRender, err := graphvizInstance.Graph()

	if err != nil {
		log.Fatal(err)
	}

	// Initializes a map that will map the TransitionGraph state's id to a cgraph.Node pointer
	// (a copy of the state that will be rendered). This will be used to render the edges later on
	associationMap := make(map[int]*cgraph.Node)

	// Bulk copy of TransitionGraph.states into renderGraph
	for _, state := range fsa.states {
		renderNode, _ := graphRender.CreateNode(fmt.Sprint(state.Id))
		associationMap[state.Id] = renderNode
	}

	// Bulk copy of the FSA's Transition into renderGraph
	for _, state := range fsa.states {
		for destId, transition := range state.transition {
			from := associationMap[state.Id]
			to := associationMap[destId]
			edgeId := fmt.Sprintf("%d-%d", state.Id, destId)
			renderEdge, _ := graphRender.CreateEdge(edgeId, from, to)
			renderEdge.SetLabel(fmt.Sprint(transition))
		}
	}

	// Creates a .png export, that saves in current working directory
	if err := graphvizInstance.RenderFilename(graphRender, graphviz.SVG, imagePath); err != nil {
		log.Fatal(err)
	}

	// Cleanup function that closes both the Graph and GraphViz instances
	defer func() {
		if err := graphRender.Close(); err != nil {
			log.Fatal(err)
		}
		graphvizInstance.Close()
	}()
}
