// Copyright Enea Guidi (hmny).

// This package handles the parsing of a given *ast.File which represents
// the content of a Go source file as an Abstract Syntax Tree.

// TODO COMMENT
package parser

import (
	"errors"
	"fmt"
	"go/ast"
	"log"
)

const (
	// Transaction type enum
	Spawn = "Spawn"
	Send  = "Send"
	Recv  = "Recv"
	Call  = "Call"

	// ArgToExpand type enum
	Function = 0
	Channel  = 1

	// Transaction error values for "from" and "to" fields
	Unknown = -1
	// Constant to identify anonymous function
	AnonymousFunc = "anonymousFunc"
)

type ArgumentToExpand struct {
	ArgIndex int
	ArgName  string
	ArgType  uint
}

type Transaction struct {
	Category  string
	IdentName string
	From      int
	To        int
}

type FunctionMetadata struct {
	Name          string
	ScopeChannels map[string]ChannelMetadata
	InlineArgs    []ArgumentToExpand
	currentState  *int
	Transactions  map[string]Transaction
}

func (fm *FunctionMetadata) addChannels(newChannels ...ChannelMetadata) {
	for _, channel := range newChannels {
		// Checks the validity of the current item
		if channel.Name != "" && channel.Typing != "" {
			fm.ScopeChannels[channel.Name] = channel
		}
	}
}

func (fm *FunctionMetadata) addTransactions(newTransactions ...Transaction) {
	for _, transaction := range newTransactions {
		// Checks the validity of the current item
		if transaction.IdentName != "" && transaction.From != Unknown && transaction.To != Unknown {
			// TODO ADD OPTIMIZED VERSION
			transactionId := fmt.Sprintf("%d-%d", transaction.From, transaction.To)
			fm.Transactions[transactionId] = transaction
		}
	}
}

func (fm FunctionMetadata) Visit(node ast.Node) ast.Visitor {
	// Skips empty nodes during descend
	if node == nil {
		return nil
	}

	switch statement := node.(type) {
	// Go routine spawn statement
	case *ast.GoStmt:
		spawnTransaction, err := getSpawnTransaction(statement, fm.currentState)
		if err != nil {
			log.Fatal(err)
		}
		fm.addTransactions(spawnTransaction)
		return nil
	// Send to a channel statement
	case *ast.SendStmt:
		sendTransaction, err := GetSendTransaction(statement, fm.currentState)
		if err != nil {
			log.Fatal(err)
		}
		fm.addTransactions(sendTransaction)
		return nil
	// Possibily, receive from channel
	case *ast.ExprStmt, *ast.AssignStmt, *ast.DeclStmt:
		recvTransactions, errRecv := GetRecvTransaction(statement, fm.currentState)
		callTransactions, errCall := getCallTransaction(statement, fm.currentState)
		channelsMeta := ExtractChanMetadata(statement)

		if errRecv != nil {
			log.Fatal(errRecv)
		} else if len(recvTransactions) > 0 {
			fm.addTransactions(recvTransactions...)
		}

		if errCall != nil {
			log.Fatal(errCall)
		} else if len(callTransactions) > 0 {
			fm.addTransactions(callTransactions...)
		}

		if len(channelsMeta) > 0 {
			fm.addChannels(channelsMeta...)
		}

		return nil
	}
	return fm
}

func NewFunctionMetadata(stmt *ast.FuncDecl) FunctionMetadata {
	// Retrieve function name and arguments
	funcName := stmt.Name.Name
	funcArgs := stmt.Type.Params.List
	// Initial setup of the metadata record
	initialState := 0
	metadata := FunctionMetadata{
		funcName,
		make(map[string]ChannelMetadata),
		nil,
		&initialState,
		make(map[string]Transaction),
	}

	// The current is an external (non Go) function, not useful for us
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
				metadata.InlineArgs = append(metadata.InlineArgs, newInlineArg)
			}

			// We're interested only in function and channel passed as arguments
			if isFunction {
				newInlineArg := ArgumentToExpand{i, argName, Function}
				metadata.InlineArgs = append(metadata.InlineArgs, newInlineArg)
			}
		}
	}

	ast.Walk(metadata, stmt.Body)

	// TODO REMOVE
	for _, t := range metadata.Transactions {
		fmt.Printf("%+v \n", t)
	}
	// Set the initial state of the fragment automata generated from the function
	// initialState := 0
	// transactionList, err := recursiveParseBlockStmt(stmt.Body, &initialState)
	// Error checking
	// if err != nil {
	// log.Fatal(err)
	// }
	// Set the list received in the metadata
	// metadata.Transactions = transactionList

	return metadata
}

func getSpawnTransaction(stmt *ast.GoStmt, currentState *int) (Transaction, error) {
	transaction := Transaction{Spawn, "", Unknown, Unknown}
	// Finds out if the function has been defined globally or we're spawning an anonymous function
	funcIdent, isFuncIdent := stmt.Call.Fun.(*ast.Ident)
	_, isFuncAnonymous := stmt.Call.Fun.(*ast.FuncLit)

	// Populates the metadata accordingly
	if isFuncIdent {
		// The avaiable function name is set but doesn't inherit scopes (and channels variables)
		transaction.IdentName = funcIdent.Name
		// Add state transaction to the automata fragment for the function
		transaction.From = *currentState
		(*currentState)++
		transaction.To = *currentState
	}

	if isFuncAnonymous {
		// The function name is a fallback one, but inherits scopes from the parent/caller
		transaction.IdentName = AnonymousFunc
		// Add state transaction to the automata fragment for the function
		transaction.From = *currentState
		(*currentState)++
		transaction.To = *currentState
		// TODO Add parent avaiableChan
		// TODO Add parse arguments (different from above)
		// TODO Should parse body of funcLiteral (?)
	}

	if !isFuncIdent && !isFuncAnonymous {
		err := errors.New("func isn't neither anonymous neither locally defined")
		return Transaction{}, err
	}

	return transaction, nil
}

// TODO COMMENT
func getCallTransaction(stmt ast.Node, currentState *int) ([]Transaction, error) {
	// Buffer in whic all the extrapolated transaction are saved
	parsed := []Transaction{}
	// Based upon the possible expression tyoe extrapolates the data needed
	switch typedStmt := stmt.(type) {
	case *ast.AssignStmt:
		// The assign statement allow for more expression inside it
		for _, rValue := range typedStmt.Rhs {
			transaction := parseCallExpr(rValue, currentState)
			// The expression isn't a recv from a channel
			if transaction.IdentName == "" {
				continue
			}
			// If the transaction is valid append it to the slice
			parsed = append(parsed, transaction)
		}
	case *ast.ExprStmt:
		transaction := parseCallExpr(typedStmt.X, currentState)
		// The expression isn't a recv from a channel
		if transaction.IdentName == "" {
			return []Transaction{}, nil
		}
		// If the transaction is valid append it to the slice
		parsed = append(parsed, transaction)
	}

	// At last returns the list of transaction extrapolated
	return parsed, nil
}

func parseCallExpr(expr ast.Expr, currentState *int) Transaction {
	// Checks if the given its a unary expression
	callExpr, isCallExpr := expr.(*ast.CallExpr)
	if !isCallExpr {
		return Transaction{}
	}

	// Checks if the nested expression its an identifier (the channel name)
	funcIdent, isIdent := callExpr.Fun.(*ast.Ident)
	if !isIdent {
		return Transaction{}
	}
	// Creates a valid transaction struct
	transaction := Transaction{Call, funcIdent.Name, Unknown, Unknown}
	// Add state transaction to the automata fragment for the function
	transaction.From = *currentState
	(*currentState)++
	transaction.To = *currentState

	return transaction
}
