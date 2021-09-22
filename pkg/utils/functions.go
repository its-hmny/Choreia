package utils

import (
	"errors"
	"fmt"
	"go/ast"
	"log"
)

const (
	// Transaction type enum
	Spawn   = 0
	Send    = 1
	Receive = 2
	Call    = 3

	// ArgToExpand type enum
	Function = 0
	Channel  = 1

	Unknown       = -1
	AnonymousFunc = "anonymousFunc"
)

type ArgumentToExpand struct {
	argIndex int
	argName  string
	argType  uint
}

type Transaction struct {
	category  int
	identName string
	from      int
	to        int
}

type FunctionMetadata struct {
	name        string
	inlineArg   []ArgumentToExpand
	transaction []Transaction
}

func GetFunctionMetadata(stmt *ast.FuncDecl) FunctionMetadata {
	// Retrieve function name and arguments
	funcName := stmt.Name.Name
	funcArgs := stmt.Type.Params.List
	// Initial setup of the metadata record
	metadata := FunctionMetadata{funcName, nil, nil}
	fmt.Printf("DEBUG function name %s \n", funcName)

	// The current is an external function
	if stmt.Body == nil {
		return FunctionMetadata{}
	}

	// If the function has arguments we parse them
	if len(funcArgs) > 0 {
		for i, arg := range funcArgs {
			// Extrapolates the argument name and type, we're only
			// interested in channel and function since they must be expanded later
			argName := arg.Names[0].Name
			_, isChannel := arg.Type.(*ast.ChanType)
			_, isFunction := arg.Type.(*ast.FuncType)

			// We're interested only in function and channel passed as arguments
			if isChannel {
				newInlineArg := ArgumentToExpand{i, argName, Channel}
				metadata.inlineArg = append(metadata.inlineArg, newInlineArg)
			}

			// We're interested only in function and channel passed as arguments
			if isFunction {
				newInlineArg := ArgumentToExpand{i, argName, Function}
				metadata.inlineArg = append(metadata.inlineArg, newInlineArg)
			}
		}
	}

	// The function has a return type (TODO is needed)
	/*
		if stmt.Type.Results != nil {
			returnVals := stmt.Type.Results.List
			fmt.Printf("  DEBUG return type %+v \n", returnVals[0])
		}
	*/

	// Parse the body of the function in order to extrapolate transaction and function call
	for _, command := range stmt.Body.List {
		switch blockStmt := command.(type) {
		// Go routine spawn statement
		case *ast.GoStmt:
			spawnTransaction, err := GetSpawnTransaction(blockStmt)
			if err != nil {
				log.Fatalf("%s\n", err)
			}
			fmt.Printf("  DEBUG go spawn %+v\n", spawnTransaction)
		// Send to a channel statement
		case *ast.SendStmt:
			sendTransaction, err := GetSendTransaction(blockStmt)
			if err != nil {
				log.Fatalf("%s\n", err)
			}
			fmt.Printf("  DEBUG send channel %+v\n", sendTransaction)
		// Possibily, receive from channel
		case *ast.ExprStmt, *ast.AssignStmt:
			sendTransactions, _ := GetRecvTransaction(blockStmt)
			if len(sendTransactions) > 0 {
				fmt.Printf("  DEBUG recv channel %+v\n", sendTransactions)
			}

		default:
			fmt.Printf("  DEBUG command %T\n", blockStmt)
		}
	}

	return metadata
}

func GetSpawnTransaction(stmt *ast.GoStmt) (Transaction, error) {
	transaction := Transaction{Spawn, "", Unknown, Unknown}

	// Finds out if the function has been defined globally or we're spawning an anonymous function
	funcIdent, isFuncIdent := stmt.Call.Fun.(*ast.Ident)
	_, isFuncAnonymous := stmt.Call.Fun.(*ast.FuncLit)

	// Populates the metadata accordingly
	if isFuncIdent {
		// The avaiable function name is set but doesn't inherit scopes (and channels variables)
		transaction.identName = funcIdent.Name

		// Checks if some arguments are channels and eventually copies their metadata
		// TODO
		/*
			for _, argExpr := range stmt.Call.Args {
				localVarIdent, isIdentifier := argExpr.(*ast.Ident)

				if !isIdentifier {
					continue
				}

				fmt.Printf("\t Argument: type %T, value: %s\n", localVarIdent, localVarIdent.Name)
			}
		*/
	}

	if isFuncAnonymous {
		// The function name is a fallback one, but inherits scopes from the parent/caller
		transaction.identName = AnonymousFunc

		// TODO Add parent avaiableChan
		// TODO Add parse arguments (different from above)
		// TODO Should parse body of funcLiteral (?)
	}

	if !isFuncIdent && !isFuncAnonymous {
		err := errors.New("GetGoRoutineMetadata: func isn't neither anonymous neither locally defined")
		return Transaction{}, err
	}

	return transaction, nil
}
