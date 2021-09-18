package utils

import (
	"fmt"
	"go/ast"
)

const (
	Spawn   uint = 0
	Send         = 1
	Receive      = 2
	Call         = 3
)

type TransactionPayload interface {
	// TODO
}

type Transaction struct {
	category uint
	buffer   TransactionPayload
	from     uint
	to       uint
}

type FunctionMetadata struct {
	name        string
	transaction []Transaction
}

func GetFunctionMetadata(stmt *ast.FuncDecl) FunctionMetadata {
	funcName := stmt.Name.Name
	funcArgs := stmt.Type.Params.List
	fmt.Printf("DEBUG function name %s \n", funcName)

	// The function has arguments
	if len(funcArgs) > 0 {
		for _, arg := range funcArgs {
			// Extrapolates the argument name and type
			argName := arg.Names[0].Name
			chanArg, isChannel := arg.Type.(*ast.ChanType)
			funcArg, isFunction := arg.Type.(*ast.FuncType)

			// We're interested only in function and channel passed as arguments
			if isChannel {
				chanType := chanArg.Value.(*ast.Ident).Name
				fmt.Printf("  DEBUG channel argument (%s,%s) \n", argName, chanType)
			}

			// We're interested only in function and channel passed as arguments
			if isFunction {
				// TODO
				fmt.Printf("  DEBUG function argument (%s,%+v) \n", argName, funcArg)
			}
		}
	}

	// The function has a return type
	if stmt.Type.Results != nil {
		returnVals := stmt.Type.Results.List
		fmt.Printf("  DEBUG return type %+v \n", returnVals[0])
	}
	return FunctionMetadata{"", nil}
}

/*
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

func GetGoRoutineMetadata(stmt *ast.GoStmts) (GoRoutineMetadata, error) {
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
}*/
