// Copyright Enea Guidi (hmny).

// This package implements a Graph data structure and its own API.
// For this specific use cases the implementation is quite simple & basic

// The only method avaiable from the outside are Transitiongraph and its API
package graph

import (
	"fmt"
	"log"
)

const (
	// Transaction type enum
	Call  = "Call"
	Eps   = "Epsilon"
	Recv  = "Recv"
	Send  = "Send"
	Spawn = "Spawn"

	// Graph edge error value
	Unknown = -1
	// Graph edge default values
	NewNode = -2
	Current = -3
)

// ----------------------------------------------------------------------------
// TransitionGraph

// A TransitionGraph is a graph that represents all the possible transition from a node to the others
//
// A struct containing a basic graph implementation that keeps track of the transaction that occur
// subsequently during the execution flow of a function (or scope).
type TransitionGraph struct {
	currentId int    // The id of the state from which the new transition (or edge) will start when an id is not specified
	Nodes     []Node // The list of node inside the graph
}

type Node struct {
	Id    int                // The id of the current node
	Edges map[int]Transition // A map to other nodeId the Transition data
}

type Transition struct {
	Kind      string // The type of Transition (Call, Eps, Recv, Send, Spawn)
	IdentName string // The identifier (variable name) on which the action is being executed
}

// This function generates a new TransitionGraph and returns a ref to it
func NewTransitionGraph() *TransitionGraph {
	return &TransitionGraph{
		currentId: 0,
		Nodes: []Node{
			{Id: 0, Edges: make(map[int]Transition)},
		},
	}
}

// Returns the last id generated
func (g *TransitionGraph) GetLastId() int {
	return len(g.Nodes) - 1
}

// Returns the id of the final state, if such state is not present, returns the last generated id
func (g *TransitionGraph) GetFinalStateId() int {
	for _, node := range g.Nodes {
		// The final state is the one for which there aren't any outcoming edges
		if len(node.Edges) == 0 {
			return node.Id
		}
	}

	return g.GetLastId()
}

// Set a new rootId, a rootId is the id of the state (node) from which all future transition will start
// when an id isn't specified, this is used since when merging multiple subgraph is needed that the merge state
// will become the one from which create new transition even if it is not the last created node
func (g *TransitionGraph) SetRootId(newRootId int) {
	g.currentId = newRootId
}

// This function adds a new Node to the TransitionGraph generating its
// id incrementally with respects to the previusly existent nodes
func (g *TransitionGraph) newNode() (id int) {
	id = g.GetLastId() + 1
	g.Nodes = append(g.Nodes, Node{
		Id:    id,
		Edges: make(map[int]Transition),
	})
	return id
}

// This function adds a new Edge and its payload the user can specify the
// from and to nodes or eventually can use some special value such as Current
// for "from" that connect the new node to the latest or NewNode for "to"
// that automatically instantiate a new node as destination of the edge
func (g *TransitionGraph) AddTransition(from, to int, t Transition) {
	// Input checking
	if from == Unknown || to == Unknown {
		log.Fatal("unknown starting or ending node on AddEdge")
	} else if t.IdentName == "" {
		return
	}

	// The user can omit a specific starting node, in this case the
	// latest added node id is used as starting point of the edge
	if from == Current {
		from = g.currentId
	}

	// The user can omit the ending node of the new edge, in this
	// case a new node is created and the edge is linked to that one
	if to == NewNode {
		to = g.newNode()
		g.SetRootId(to)
	}

	fmt.Printf("BP__ %d -> %d \t %+v\n", from, to, t)

	// Creates/assign the new edge
	g.Nodes[from].Edges[to] = t
}

// This function expands a graph in place of an edge/transition. Since in our case every
// Automata/Graph has only one initial and final state then we simply copy the other graph
// node by node and edge by edge and then we link the copy to the "from" and "to" nodes
func (g *TransitionGraph) ExpandInPlace(from, to int, other TransitionGraph) {
	// First of all remove the old call transition (since it will be expanded)
	delete(g.Nodes[from].Edges, to)
	// Calculate the offset from which the ids of the "other" will be padded with
	offset := len(g.Nodes)

	// Copies the "other" graph node, applying the offset to each id
	for _, cpNode := range other.Nodes {
		// Creates a new edge map for and copies all the edges to it
		newNodeEdges := make(map[int]Transition)
		for dest, t := range cpNode.Edges {
			newDest := offset + dest
			newNodeEdges[newDest] = t
		}
		// Then creates a new node and adds it to the destination graph
		newNode := Node{Id: offset + cpNode.Id, Edges: newNodeEdges}
		g.Nodes = append(g.Nodes, newNode)
	}

	// Eps transition to mark/link the start of the expanded/copied graph
	tExpansionStart := Transition{Kind: Eps, IdentName: "start-call-expansion"}
	tExpansionEnd := Transition{Kind: Eps, IdentName: "end-call-expansion"}
	// Initial and final state of the freshly copied graph
	copyInitialStateId, copyFinalStateId := offset, other.GetFinalStateId()
	// The "linking" transition are now added to the graph, completing the expansion
	g.AddTransition(from, copyInitialStateId, tExpansionStart)
	g.AddTransition(copyFinalStateId, to, tExpansionEnd)
}

// This function provides a simple mechanism to execute a function on each
// edge of the graph. For convenience a ref to the transition is passed so that the
// callback function can possibly change the transition if needed
func (g *TransitionGraph) ForEachTransition(toExecute func(from, to int, t *Transition)) {
	for _, node := range g.Nodes {
		for dest, edge := range node.Edges {
			toExecute(node.Id, dest, &edge)
		}
	}
}
