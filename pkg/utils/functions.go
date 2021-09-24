package utils

import (
	"errors"
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
	name         string
	inlineArg    []ArgumentToExpand
	transactions []Transaction
}

func GetFunctionMetadata(stmt *ast.FuncDecl) FunctionMetadata {
	// Retrieve function name and arguments
	funcName := stmt.Name.Name
	funcArgs := stmt.Type.Params.List
	// Initial setup of the metadata record
	metadata := FunctionMetadata{funcName, nil, nil}

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

	// Set the initial state of the fragment automata generated from the function
	initialState := 0
	transactionList, err := recursiveParseBlockStmt(stmt.Body, &initialState)
	// Error checking
	if err != nil {
		log.Fatalf("%s\n", err)
	}
	// Set the list received in the metadata
	metadata.transactions = transactionList

	return metadata
}

func GetSpawnTransaction(stmt *ast.GoStmt, currentState *int) (Transaction, error) {
	transaction := Transaction{Spawn, "", Unknown, Unknown}

	// Finds out if the function has been defined globally or we're spawning an anonymous function
	funcIdent, isFuncIdent := stmt.Call.Fun.(*ast.Ident)
	_, isFuncAnonymous := stmt.Call.Fun.(*ast.FuncLit)

	// Populates the metadata accordingly
	if isFuncIdent {
		// The avaiable function name is set but doesn't inherit scopes (and channels variables)
		transaction.identName = funcIdent.Name
		// Add state transaction to the automata fragment for the function
		transaction.from = *currentState
		(*currentState)++
		transaction.to = *currentState
	}

	if isFuncAnonymous {
		// The function name is a fallback one, but inherits scopes from the parent/caller
		transaction.identName = AnonymousFunc
		// Add state transaction to the automata fragment for the function
		transaction.from = *currentState
		(*currentState)++
		transaction.to = *currentState
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

func recursiveParseBlockStmt(body *ast.BlockStmt, currentState *int) ([]Transaction, error) {
	transactionList := []Transaction{}
	// Arguments checking
	if currentState == nil {
		return []Transaction{}, errors.New("recursiveParseBlockStmt: passed nil value as currentState")
	}

	// Parse the body of the function in order to extrapolate transaction and function call
	for _, command := range body.List {
		switch blockStmt := command.(type) {
		// Go routine spawn statement
		case *ast.GoStmt:
			spawnTransaction, err := GetSpawnTransaction(blockStmt, currentState)
			if err != nil {
				log.Fatalf("%s\n", err)
			}
			transactionList = append(transactionList, spawnTransaction)
		// Send to a channel statement
		case *ast.SendStmt:
			sendTransaction, err := GetSendTransaction(blockStmt, currentState)
			if err != nil {
				log.Fatalf("%s\n", err)
			}
			transactionList = append(transactionList, sendTransaction)
		// Possibily, receive from channel
		case *ast.ExprStmt, *ast.AssignStmt:
			recvTransactions, _ := GetRecvTransaction(blockStmt, currentState)
			if len(recvTransactions) > 0 {
				transactionList = append(transactionList, recvTransactions...)
			}
		}
	}

	return transactionList, nil
}
