// Copyright Enea Guidi (hmny).

// This package handles the parsing of a given *ast.File which represents
// the content of a Go source file as an Abstract Syntax Tree.

// The only method avaiable from the outside are Transitiongraph and its API
package parser

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

/*
? Can be useful later
// Neighbors : returns a list of node IDs that are linked to this node
func (g *Graph) Neighbors(id int) []int {
	neighbors := []int{}
	for _, node := range g.nodes {
		for edge := range node.edges {
			if node.id == id {
				neighbors = append(neighbors, edge)
			}
			if edge == id {
				neighbors = append(neighbors, node.id)
			}
		}
	}
	return neighbors
}

// Nodes : returns a list of node IDs
func (g *Graph) Nodes() []int {
	nodes := make([]int, len(g.nodes))
	for i := range g.nodes {
		nodes[i] = i
	}
	return nodes
}

// Edges : returns a list of edges with weights
func (g *Graph) Edges() [][3]int {
	edges := make([][3]int, 0, len(g.nodes))
	for i := range g.nodes {
		for k, v := range g.nodes[i].edges {
			edges = append(edges, [3]int{i, k, int(v)})
		}
	}
	return edges
}
*/
