// Copyright Enea Guidi (hmny).

// This package implements a Closure (Closure) data structure and its own API.
// For this specific use cases the implementation is quite simple & basic

// The only method available from the outside are Closure and its API
package closure

import (
	"fmt"
	"log"

	"github.com/goccy/go-graphviz"
	"github.com/its-hmny/Choreia/internal/data_structures/fsa"
)

var latestId int = 0

// Closure is an implementation of a Set using the builtin map type.
type Closure struct {
	Id    int               // The UID of the closure
	items map[int]fsa.State // A map that defines which element belong to the closure and which do not
}

// Add will add the provided items to the closure.
func (closure *Closure) IsEmpty() bool {
	return len(closure.items) == 0
}

// Add will add the provided items to the closure.
func (closure *Closure) Add(items ...fsa.State) {
	for _, item := range items {
		closure.items[item.Id] = item
	}
}

// Remove will remove the given items from the closure.
func (closure *Closure) Remove(items ...fsa.State) {
	for _, item := range items {
		delete(closure.items, item.Id)
	}
}

// Contains returns a bool indicating if the given items exists in the closure.
func (closure *Closure) Contains(items ...fsa.State) bool {
	for _, item := range items {
		if _, ok := closure.items[item.Id]; !ok {
			return false
		}
	}

	return true
}

// Exists returns a bool indicating if the given stateId exists in the closure.
func (closure *Closure) Exist(key int) bool {
	_, exist := closure.items[key]
	return exist
}

// IsEqual returns a bool indicating if the given closure is equal to the one provided.
func (closure *Closure) IsEqual(other *Closure) bool {
	if len(closure.items) != len(other.items) {
		return false
	}

	// Checks element by element that each item in other is an item in closure as well
	for _, otherElem := range other.Iterator() {
		if _, exist := closure.items[otherElem.Id]; !exist {
			return false // If an element isn't present then false is returned
		}
	}
	return true
}

// Iterator will return a list of the fsa.State in the closure.
func (closure *Closure) Iterator() []fsa.State {
	flattened := make([]fsa.State, 0, len(closure.items))
	for _, item := range closure.items {
		flattened = append(flattened, item)
	}
	return flattened
}

// Iterator will return a list of possible fsa.Transition possible from the closure.
func (closure *Closure) TransitionIterator() []fsa.Transition {
	list := []fsa.Transition{}
	for _, state := range closure.items {
		for _, t := range state.TransitionIterator() {
			// Ignore eps transition
			if t.Move == fsa.Eps {
				continue
			}

			list = append(list, t)
		}
	}
	return list
}

// ExportAsSVG will export a .svg representation of the closure saved at the given path.
func (closure *Closure) ExportAsSVG(path string) {
	// Creates a GraphViz instance and initializes a Graph instance
	graphvizInstance := graphviz.New()
	graphRender, err := graphvizInstance.Graph()

	if err != nil {
		log.Fatal(err)
	}

	graphRender.SetLayout("neato")

	// Bulk copy of all the item in the Closure into renderGraph
	for id := range closure.items {
		renderNode, _ := graphRender.CreateNode(fmt.Sprint(id))
		renderNode.SetColor("#40c4e6")
	}

	// Creates a .svg export, that saves at the given path
	if err := graphvizInstance.RenderFilename(graphRender, graphviz.SVG, path); err != nil {
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

// New is the constructor for closures. It will pull from a reuseable memory pool if it can.
// Takes a list of items to initialize the closure with.
func New(items ...fsa.State) *Closure {
	closure := Closure{Id: latestId, items: make(map[int]fsa.State)}

	for _, item := range items {
		closure.items[item.Id] = item
	}

	latestId++
	return &closure
}
