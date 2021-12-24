// Copyright Enea Guidi (hmny).

// TODO COMMENT

// TODO COMMENT
package transforms

import (
	"github.com/its-hmny/Choreia/internal/data_structures/fsa"
	meta "github.com/its-hmny/Choreia/internal/static_analysis"
)

// TODO comment
// TODO comment
// TODO comment
func GenerateDCA(fileMeta meta.FileMetadata) *fsa.FSA {
	// Takes the deterministic version of the Partial Automaton and merges them
	// in one DCA that will represent the choreography as a whole
	// TODO implement
	return fsa.New()
}
