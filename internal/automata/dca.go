// Copyright Enea Guidi (hmny).

// This package handles the extraction of Partial Nondeterministic Automatas from
// metadata extracted and the handling and subsequent transformation of abovesaid
// NCA until a single Deterministic Choreography Automata is obtained by them

// This module defines some helper function to transform and work with DCAs
// (Deterministic Choreography Automatas). Such transformations could be extracting
// the DCAs from given metadata or merging two or more Partial/Projection DCAs in one
// unified DCAs that represents the Choreography as a whole.
package automata

import (
	"fmt"
	"log"

	"github.com/its-hmny/Choreia/internal/meta"
	"github.com/its-hmny/Choreia/internal/types/fsa"
)

// TODO comment
// TODO comment
// TODO comment
func GenerateDCA(fileMeta meta.FileMetadata) *fsa.FSA {
	mainFuncMeta, exist := fileMeta.FunctionMeta["main"]

	if !exist {
		log.Fatal("Cannot extract Partial Automata, 'main' function metadata not found")
	}

	// Extracts reursively from the metadata the Projection DCAs, each one of them
	// will be a projection of the final one but it has lost all of his eps-transition
	projectionDCAs := extractProjectionDCAs(mainFuncMeta, fileMeta)

	// ! Debug print, will be removed
	fmt.Printf("Successfully extracted %d Projection NCAs\n", len(projectionDCAs))
	for i, DCA := range projectionDCAs {
		DCA.ExportAsSVG(fmt.Sprintf("debug/projectionDCAs-%d.svg", i))
	}

	// Takes the deterministic version of the Partial Automatas and merges them
	// in one DCA that will represent the choreography as a whole
	// TODO implement
	return fsa.New()
}
