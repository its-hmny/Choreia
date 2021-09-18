package utils

import (
	"errors"
	"fmt"
	"go/ast"
)

const ANONYMOUS_IDENT = "anonymous"

var latestUid uint = 0 // Uid = 0 is always taken by main

// Struct containing the metadata avaiable about a Go Routine (main included)
// TODO ADD DOCS
type GoRoutineMetadata struct {
	goRoutineUid uint
	name         string
	// avaiableChan []ChannelMetadata
	// transition   []int
}

type FunctionMetadata struct {
}

func GetGoRoutineMetadata(stmt *ast.GoStmt /*, parentMetadata GoRoutineMetadata*/) (GoRoutineMetadata, error) {
	latestUid += 1 // Generates a new Uid
	metadata := GoRoutineMetadata{goRoutineUid: latestUid}

	// Finds out if the function has been defined globally or we're spawning an anonymous function
	funcIdent, isFuncIdent := stmt.Call.Fun.(*ast.Ident)
	_, isFuncAnonymous := stmt.Call.Fun.(*ast.FuncLit)

	// Populates the metadata accordingly
	if isFuncIdent {
		// The avaiable function name is set but doesn't inherit scopes (and channels variables)
		metadata.name = funcIdent.Name
		fmt.Printf("Go Routine %s-%d started\n", metadata.name, metadata.goRoutineUid)

		// Checks if some arguments are channels and eventually copies their metadata
		for _, argExpr := range stmt.Call.Args {
			localVarIdent, isIdentifier := argExpr.(*ast.Ident)

			if !isIdentifier {
				continue
			}

			fmt.Printf("\t Argument: type %T, value: %s\n", localVarIdent, localVarIdent.Name)
		}
	}

	if isFuncAnonymous {
		// The function name is a fallback one, but inherits scopes from the parent/caller
		metadata.name = ANONYMOUS_IDENT
		fmt.Printf("Go Routine %s-%d started\n", metadata.name, metadata.goRoutineUid)

		// TODO Add parent avaiableChan
		// TODO Add parse arguments (different from above)
		// TODO Should parse body of funcLiteral (?)
	}

	if !isFuncIdent && !isFuncAnonymous {
		err := errors.New("GetGoRoutineMetadata: func isn't neither anonymous neither locally defined")
		return metadata, err
	}

	return metadata, nil
}
