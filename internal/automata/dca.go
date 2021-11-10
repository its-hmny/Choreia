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

	"github.com/Workiva/go-datastructures/list"
	"github.com/its-hmny/Choreia/internal/meta"
	"github.com/its-hmny/Choreia/internal/types/fsa"
)

type SimulatedFSA struct {
	fsa          *fsa.FSA
	currentState int
}

// TODO comment
// TODO comment
// TODO comment
func GenerateDCA(fileMeta meta.FileMetadata) *fsa.FSA {
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
		NCA.ExportAsSVG(fmt.Sprintf("debug/before-eps-removal-%d.svg", i))
		projectionDCAs[i] = subsetConstructionAlgorithm(NCA)
		projectionDCAs[i].ExportAsSVG(fmt.Sprintf("debug/after-eps-removal-%d.svg", i))
	}

	// Takes the deterministic version of the Partial Automatas and merges them
	// in one DCA that will represent the choreography as a whole
	return reconciliationAlgorithm(projectionNDCAs)
}

func reconciliationAlgorithm(NDCAs []*fsa.FSA) *fsa.FSA {
	finalDCA := fsa.New()
	activeNDCAs := list.Empty.Add(SimulatedFSA{NDCAs[0], 0})
	blockedNDCAs := list.Empty

	for !activeNDCAs.IsEmpty() {
		// Get a reference to a NDCA and before removing it from the list
		item, exist := activeNDCAs.Get(0)
		activeNDCAs, _ = activeNDCAs.Remove(0)

		// Some work with the fsa
		simNDCA, isSimulatedFSA := item.(SimulatedFSA)
		currentSimState := simNDCA.fsa.GetState(simNDCA.currentState)

		if !exist || !isSimulatedFSA {
			log.Fatal("Error while removing NDCA from list")
		}

		for destId, transition := range currentSimState.TransitionIterator() {
			// ! Will be removed
			fmt.Printf("Encountered %s\n", transition)

			// ? How to behave with "recursive" transition (from x to x)
			if destId == currentSimState.Id {
				continue
			}

			switch transition.Move {
			case fsa.Eps:
				// Inserts the current back at the top of the list
				activeNDCAs = activeNDCAs.Add(SimulatedFSA{simNDCA.fsa, destId})
			case fsa.Spawn:
				// Inserts the spawnee at the bottom of the list
				spawneeFSA := transition.Payload.(*fsa.FSA)
				activeNDCAs, _ = activeNDCAs.Insert(SimulatedFSA{spawneeFSA, 0}, activeNDCAs.Length())
				// But the current keep is previous position at the top of the list
				activeNDCAs = activeNDCAs.Add(SimulatedFSA{simNDCA.fsa, destId})
			case fsa.Send:
			case fsa.Recv:
			default:
				log.Fatal("Unrecognized transition move", destId)
			}

		}
	}

	if !blockedNDCAs.IsEmpty() {
		log.Fatal("Some NDCA(s) are still waiting, there's some problem in your code")
	}

	return finalDCA
}
