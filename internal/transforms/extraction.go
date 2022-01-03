// Copyright Enea Guidi (hmny).

// TODO COMMENT

// TODO comment
package transforms

import (
	"fmt"
	"log"

	"github.com/its-hmny/Choreia/internal/data_structures/fsa"
	meta "github.com/its-hmny/Choreia/internal/static_analysis"
)

var nProjectionExtracted = 0

// -------------------------------------------------------------------------------------------
// ProjectionAutomata

// A FSA that represents the execution flow of a single Goroutine (identified by its own name)
type ProjectionAutomata struct {
	Name     string   // An identifier for the Automata
	Automata *fsa.FSA // The FSA itself
}

// Extracts the Projection (or Local) Choreography Automata for the given FuncMeta. The ScopeAutomata of
// said function is used as entry point of the Projection CA. Every call to another function is expanded inline
// in Local Automata (channel and callbacks passed as argument are expanded as well). When a new GoRoutine is
// spawned during the execution flow a new Local Automata is generated. All the CA are transformed to a
// deterministic version with the Subset Construction Algorithm
func GetLocalViews(function meta.FuncMetadata, file meta.FileMetadata) []*ProjectionAutomata {
	// Creates the Projection Automata for the current GoRoutine
	current := ProjectionAutomata{
		Name:     fmt.Sprintf("Goroutine %d (%s)", nProjectionExtracted, function.Name),
		Automata: function.ScopeAutomata.Copy(), // Makes a full independent copy of the funcMeta.ScopeAutomata
	}

	nProjectionExtracted++
	extractedList := []*ProjectionAutomata{&current}

	// Iterates over each transition in the ScopeAutomata
	function.ScopeAutomata.ForEachTransition(func(from, to int, t fsa.Transition) {
		switch t.Move {
		case fsa.Call:
			inlineCallTransition(file, current.Automata, from, to, t)
		case fsa.Spawn:
			newLocalAutomata := extractSpawnTransition(file, current.Automata, from, to, t)
			extractedList = append(extractedList, newLocalAutomata...)
		}
	})

	// extractedList[0].Automata = subsetConstruction(current.Automata)
	return extractedList
}

// TODO Comment
// TODO Comment
// TODO Comment
func inlineCallTransition(file meta.FileMetadata, root *fsa.FSA, from, to int, t fsa.Transition) {
	// Tries to retrieve the called function metadata from the file
	calledFunc, hasMeta := file.FunctionMeta[t.Label]

	// Could not find metadata (probably built-in or external function such as "make" or "append")
	// the call transition is replaced with an eps-transition
	if !hasMeta {
		newT := fsa.Transition{Move: fsa.Eps, Label: "unknown-function-call"}
		root.RemoveTransition(from, to, t)
		root.AddTransition(from, to, newT)
		return
	}

	// Expands positional argument (Channel e Function/Callback) with the actual arguments provided
	// by the caller (Find and Replace over the called ScopeAutomata) then inline the called ScopeAutomata
	// into the root one
	calledScopeAutomata := replaceActualArgs(t, calledFunc)
	inlineAutomata(root, from, to, t, calledScopeAutomata)
}

// TODO Comment
// TODO Comment
// TODO Comment
func extractSpawnTransition(file meta.FileMetadata, root *fsa.FSA, from, to int, t fsa.Transition) []*ProjectionAutomata {
	// Tries to retrieve the called function metadata from the file
	calledFunc, hasMeta := file.FunctionMeta[t.Label]

	// Could not find metadata (probably built-in or external function such as "make" or "append")
	// the Choreography cannot be determined correctly with a missing actor so the program fails
	if !hasMeta {
		log.Fatalf("Couldn't find function %s spawned as Go Routine\n", t.Label)
	}

	// Expands positional arguments with the actual ones
	calledFunc.ScopeAutomata = replaceActualArgs(t, calledFunc)

	// Recursively call getProjectionAutomata on the spawned GoRoutine entrypoint
	//(the function called with go keyword), then returns the extracted Projection Automata
	// the first is always the spawned one, the others can be function spawned by the one spawned by us
	newLocalAutomata := GetLocalViews(calledFunc, file)

	// Overrides the older transition with additional data (reference to the spawned ProjectionAutomata)
	newT := fsa.Transition{Move: t.Move, Label: t.Label, Payload: newLocalAutomata[0]}
	root.RemoveTransition(from, to, t)
	root.AddTransition(from, to, newT)

	return newLocalAutomata
}

// Expand positional argument (Channel e Function/Callback) with the actual ones so that any
// Transition reference the actual channel at runtime. The function uses the offset of the arguments
// and their types to determine which positional argument has to be replaced with the "actual" ones
func replaceActualArgs(t fsa.Transition, calledFunc meta.FuncMetadata) *fsa.FSA {
	// Copies the ScopeAutomata of the called function
	calledAutomataCp := calledFunc.ScopeAutomata.Copy()
	// Retrieves the actual arguments from the Call transition
	expandArgs, isList := t.Payload.([]meta.FuncArg)

	// Bails out at the first discrepancy returning a non-expanded copy
	if !isList || len(expandArgs) <= 0 || len(calledFunc.InlineArgs) <= 0 {
		// TODO REMOVE => return calledAutomataCp
		log.Fatal("Could not expand actual arguments: requested %d but given %d\n", len(calledFunc.InlineArgs), len(expandArgs))
	}

	// Expands the actual arguments with the positional ones
	for _, actualArg := range expandArgs {
		for _, funcArg := range calledFunc.InlineArgs {
			// Tries to find a match beetwen the actual argument and the positional argument
			if funcArg.Offset != actualArg.Offset || funcArg.Type != actualArg.Type {
				continue
			}

			// If such match is found then all the transition in the localCopy that references that
			// positional argument are replaced with transition to the actual argument
			calledAutomataCp.ForEachTransition(func(from, to int, t fsa.Transition) {
				if funcArg.Type == meta.Channel && (t.Move == fsa.Recv || t.Move == fsa.Send) {
					newT := fsa.Transition{Move: t.Move, Label: actualArg.Name}
					// Replace the transitions in the copied ScopeAutomata
					calledAutomataCp.RemoveTransition(from, to, t)
					calledAutomataCp.AddTransition(from, to, newT)
				}
				// ? Handle funcArg.Type == Function as well
			})
		}
	}

	return calledAutomataCp
}

// This function expands a graph in place of an transition. Since in our case every
// Automata/Graph has only one initial and final state then we simply copy the other graph
// state by state and transition by transition and then we link the copy to the "from" and "to" states
// TODO COMMENTS
func inlineAutomata(root *fsa.FSA, from, to int, t fsa.Transition, other *fsa.FSA) {
	// First of all remove the old call transition
	root.RemoveTransition(from, to, t)

	// Count the number of states, in order to extract an offset
	offset := 0
	root.ForEachState(func(_ int) { offset++ })

	// Copies the "other" graph state, applying the offset to each id
	other.ForEachTransition(func(from, to int, t fsa.Transition) {
		root.AddTransition(from+offset, to+offset, t)
	})

	// Links the initial state of "other" FSA with the "root" FSA via eps transition
	tExpansionStart := fsa.Transition{Move: fsa.Eps, Label: "start-call-expansion"}
	root.AddTransition(from, offset, tExpansionStart)

	// Links every final/accepting states of the other FSA with the "root" via eps transition
	for _, item := range other.FinalStates.Values() {
		finalStateId := item.(int)
		tExpansionEnd := fsa.Transition{Move: fsa.Eps, Label: "end-call-expansion"}
		root.AddTransition(finalStateId+offset, to, tExpansionEnd)
	}

}
