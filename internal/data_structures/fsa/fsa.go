// Copyright Enea Guidi (hmny).

// This package implements a Finite State Automata (FSA) data structure and its own API.
// For this specific use cases the implementation is quite simple & basic

// The only method available from the outside are FSA and its API
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
// FSA

// A FSA is a graph that represents all the possible transition from a state to the others
//
// A struct containing a basic graph implementation that keeps track of the transaction that occur
// subsequently during the execution flow of a function (or scope).
type FSA struct {
	currentId int     // The last id generated, the id of the last node
	states    []State // The list of state inside the graph
}

type State struct {
	Id         int                  // The id of the current state
	IsFinal    bool                 // Flag to indicate that a state si final/accepting
	transition map[int][]Transition // A map to other stateId the Transition data
}

// This function generates a new FSA and returns a pointer reference to it
func New() *FSA {
	return &FSA{
		currentId: 0,
		// Every FSA has already the first (0) state inside
		states: []State{{Id: 0, IsFinal: false, transition: make(map[int][]Transition)}},
	}
}

// This function generates an independent copy of the given FSA and returns it
func (original *FSA) Copy() *FSA {
	fsaCopy := FSA{
		currentId: original.currentId,
		states:    make([]State, 0, len(original.states)),
	}

	for _, currentState := range original.states {
		// Creates a new map, overriding the copied one (since map instance remains intertwined)
		currentState.transition = make(map[int][]Transition)
		// Copies all the entry of the other map into the decoupled one
		for dest, destTransitions := range original.states[currentState.Id].transition {
			tCopy := make([]Transition, len(destTransitions))
			copy(tCopy, destTransitions)
			currentState.transition[dest] = tCopy
		}

		// Then adds the fully copied state to the copy FSA
		fsaCopy.states = append(fsaCopy.states, currentState)
	}

	return &fsaCopy
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
	return 0 // TODO REDO COMPLETELY
}

// Returns the state identified by the given id,if not found returns error
func (fsa *FSA) GetState(stateId int) State {
	found := fsa.states[stateId]
	return found
}

// This function adds a new State to the TransitionGraph generating its
// id incrementally with respects to the previously existent state
func (fsa *FSA) newState() (id int) {
	id = len(fsa.states) // Generates a new id
	fsa.states = append(fsa.states, State{
		Id:         id,
		IsFinal:    false,
		transition: make(map[int][]Transition),
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
	localCopy := fsa.states[from].transition[to]
	fsa.states[from].transition[to] = append(localCopy, t)
}

// TODO COMMENT
func (fsa *FSA) ForEachTransition(callback func(from, to int, t Transition)) {
	for _, fromState := range fsa.states {
		for destState, transitions := range fromState.transition {
			for _, t := range transitions {
				callback(fromState.Id, destState, t)
			}
		}
	}
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
	// Saves the id of the final state
	finalStateId := fsa.GetFinalStateId()

	// Bulk copy of TransitionGraph.states into renderGraph
	for _, state := range fsa.states {
		renderNode, _ := graphRender.CreateNode(fmt.Sprint(state.Id))
		associationMap[state.Id] = renderNode
		renderNode.SetShape(cgraph.CircleShape)
		if finalStateId == state.Id {
			renderNode.SetShape(cgraph.DoubleCircleShape)
		}
	}

	// Bulk copy of the FSA's Transition into renderGraph
	for _, state := range fsa.states {
		for destId, transitions := range state.transition {
			for i, t := range transitions {
				from := associationMap[state.Id]
				to := associationMap[destId]
				edgeId := fmt.Sprintf("%d-%d-%d", state.Id, destId, i)
				renderEdge, _ := graphRender.CreateEdge(edgeId, from, to)
				renderEdge.SetLabel(fmt.Sprint(t))
			}
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
