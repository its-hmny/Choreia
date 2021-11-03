// Copyright Enea Guidi (hmny).

// This package implements a Closure (Closure) data structure and its own API.
// For this specific use cases the implementation is quite simple & basic

// The only method avaiable from the outside are Closure and its API
package closure

import (
	"fmt"
	"log"

	"github.com/goccy/go-graphviz"
	"github.com/its-hmny/Choreia/internal/types/fsa"
)

// Closure is an implementation of a Set using the builtin map type.
type Closure struct {
	items map[int]fsa.State
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

// Iteator will return a list of the fsa.State in the closure.
func (closure *Closure) Iterator() []fsa.State {
	flattened := make([]fsa.State, 0, len(closure.items))
	for _, item := range closure.items {
		flattened = append(flattened, item)
	}
	return flattened
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
	closure := Closure{items: make(map[int]fsa.State)}
	for _, item := range items {
		closure.items[item.Id] = item
	}

	return &closure
}
