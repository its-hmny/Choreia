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
func GenerateDCA(fileMeta meta.FileMetadata) fsa.FSA {
	mainFuncMeta, exist := fileMeta.FunctionMeta["main"]

	if !exist {
		log.Fatal("Cannot extract Partial Automata, 'main' function metadata not found")
	}

	// Extracts reursively from the metadata the Projection NDCA, each one of them
	// will be a projection of the final one and will still have eps-transtion
	projectionNDCAs := extractProjectionNDCAs(mainFuncMeta, fileMeta)
	projectionDCAs := make([]*fsa.FSA, len(projectionNDCAs))

	// ! Debug print, will be removed
	fmt.Printf("Successfully extracted %d Projection NCAs\n", len(projectionNDCAs))

	// Removes eps-transition from each Projection NDCA transforming them in
	// equivalent DCA (but we're still working with Projection DCA)
	for i, NCA := range projectionNDCAs {
		if i == 0 { // ! Will be removed
			NCA.ExportAsSVG(fmt.Sprintf("debug/before-eps-removal-%d.svg", i))

			projectionDCAs[i] = subsetConstructionAlgorithm(NCA)

			projectionDCAs[i].ExportAsSVG(fmt.Sprintf("debug/after-eps-removal-%d.svg", i))
		}
	}

	// Takes the deterministic version of the Partial Automatas and merges them
	// in one DCA that will represent the choreography as a whole
	// TODO IMPLEMENT

	return fsa.FSA{}
}
