package utils

import (
	"fmt"
	"go/ast"
)

var latestUid uint = 1 // Uid = 0 is always taken by main

// Struct containing the metadata avaiable about a Go Routine (main included)
// TODO ADD DOCS
type GoRoutineMetadata struct {
	goRoutineUid uint
	name         string
	avaiableChan []ChannelMetadata
	//transition   []int
}

func GetGoRoutineMetadata(stmt *ast.GoStmt /*, parentMetadata GoRoutineMetadata*/) (GoRoutineMetadata, error) {
	// Retrieves basic info about the newly spawned Go Routine
	funcName := stmt.Call.Fun.(*ast.Ident).Name
	latestUid++ // Generates a new Uid

	// TODO add eventually inherited (by scope or binding) channels
	metadata := GoRoutineMetadata{latestUid, funcName, nil}
	fmt.Printf("Go Routine %s started\n", funcName)

	// Checks if some arguments are channels and eventually copies their metadata
	for _, argExpr := range stmt.Call.Args {
		localVarIdent, isIdentifier := argExpr.(*ast.Ident)

		if !isIdentifier {
			continue
		}

		fmt.Printf("\t Argument: type %T, value: %s\n", localVarIdent, localVarIdent.Name)
	}

	return metadata, nil
}
