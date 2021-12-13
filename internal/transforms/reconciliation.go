// Copyright Enea Guidi (hmny).

// TODO COMMENT

// TODO COMMENT
package transforms

import (
	"fmt"
	"log"

	"github.com/goccy/go-graphviz"
	"github.com/its-hmny/Choreia/internal/data_structures/fsa"
	meta "github.com/its-hmny/Choreia/internal/static_analysis"
)

// TODO comment
// TODO comment
// TODO comment
func GenerateDCA(fileMeta meta.FileMetadata) *fsa.FSA {
	mainFuncMeta, exist := fileMeta.FunctionMeta["main"]

	if !exist {
		log.Fatal("Cannot extract Partial Automata, 'main' function metadata not found")
	}

	// Extracts recursively from the metadata the Projection DCAs, each one of them
	// will be a projection of the final one but it has lost all of his eps-transition
	localAutomata := getProjectionAutomata(mainFuncMeta, fileMeta)

	// ! Debug print and export, will be removed
	fmt.Printf("Successfully extracted %d Projection NCAs\n", len(localAutomata))
	for _, DCA := range localAutomata {
		filename := fmt.Sprintf("debug/%s.svg", DCA.name)
		DCA.automata.Export(filename, graphviz.SVG)
	}

	// Takes the deterministic version of the Partial Automaton and merges them
	// in one DCA that will represent the choreography as a whole
	// TODO implement
	return fsa.New()
}
