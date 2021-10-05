// Copyright Enea Guidi (hmny).

// This package handles the parsing of a given *ast.File which represents
// the content of a Go source file as an Abstract Syntax Tree.

// The only methods avaiable from the outside are ExtractChanMetadata,
// GetSendTransaction and GetRecvTransaction, they return respectively the
// ChannelMetadata struct and a Transaction struct for the latter.
package parser

import (
	"errors"
	"go/ast"
	"go/token"
	"log"
)

// Struct containing the metadata extrapolated about a Channel
// declared inside the program
// name - The name of the variable assigned to that channel
// typing - The type of the channel (int, string, ...)
// async - The type of comunication is avaiable, async for buffered and sync for unbuffered
type ChannelMetadata struct {
	Name   string
	Typing string
	Async  bool
}

// Is possible to declare a variable directly both with the
// := notation (AssignStmt) or with a more verbose syntax
// var chan = make(...) (GenDecl) this method handles both
func ExtractChanMetadata(node ast.Node) []ChannelMetadata {
	// Determines the type of the given expression
	switch statement := node.(type) {
	// Type: chan := make(...)
	case *ast.AssignStmt:
		newChannels, err := parseAssignStmt(statement)
		if err != nil {
			log.Fatal(err)
			return nil
		}
		return newChannels

	// Type: var chan = make(...)
	case *ast.DeclStmt:
		newChannels, err := parseDeclStmt(statement)
		if err != nil {
			log.Fatal(err)
			return nil
		}
		return newChannels
	// Type: var chan = make(...) but in global scope is not wrapped by a DeclStmt
	case *ast.GenDecl:
		newChannels, err := parseGenDecl(statement)
		if err != nil {
			log.Fatal(err)
			return nil
		}
		return newChannels
	}
	return nil
}

// TODO COMMENT This function extrapolates a compliant transaction struct from a send statement
// (obviously invoked on a channel). If at any point an error is encountered the func
// bails out returning an error.
// NOTE: the currentState pointer should not be nil
func GetSelectTransaction(stmt *ast.SelectStmt, currentState *int) []Transaction {
	// Accumulator buffer for the (to be extracted) transaction
	caseTransactions := []Transaction{}

	for _, bodyStmt := range stmt.Body.List {
		// Local copy from which fork the path of each case statement
		localCopyState := *currentState
		commClause := bodyStmt.(*ast.CommClause)

		// The current is the default case of the select statement
		if commClause.Comm == nil {
			continue
		}

		// Else we're evaluating a Send or a Receive from a channel
		switch commStmt := commClause.Comm.(type) {
		case *ast.SendStmt:
			// ! The transaction aren't inserted corretcly (change when possible)
			newTransaction, _ := GetSendTransaction(commStmt, &localCopyState)
			caseTransactions = append(caseTransactions, newTransaction)
		case *ast.AssignStmt, *ast.ExprStmt: // Receive from a channel
			// ! The transaction aren't inserted corretcly (change when possible)
			newTransactions, _ := GetRecvTransaction(commStmt, &localCopyState)
			caseTransactions = append(caseTransactions, newTransactions...)
		}
	}

	// Return the extrapolated transaction
	return caseTransactions
}

// This function extrapolates a compliant transaction struct from a send statement
// (obviously invoked on a channel). If at any point an error is encountered the func
// bails out returning an error.
// NOTE: the currentState pointer should not be nil
func GetSendTransaction(stmt *ast.SendStmt, currentState *int) (Transaction, error) {
	chanIdent, isIdent := stmt.Chan.(*ast.Ident)
	if isIdent {
		transaction := Transaction{Send, chanIdent.Name, Unknown, Unknown}
		// Add state transaction to the automata fragment for the function
		transaction.From = *currentState
		(*currentState)++
		transaction.To = *currentState
		// At last returns the transaction
		return transaction, nil
	}
	return Transaction{}, errors.New("the channel isn't an identifier")
}

// This function extrapolates a compliant transaction struct from a recv statement.
// At the moment there are only 2 type of stmt that could contain this send operation
// inside it, both cases are handled in this function
// If at any point an error is encountered the func bails out returning an error.
// NOTE: the currentState pointer should not be nil
func GetRecvTransaction(stmt ast.Node, currentState *int) ([]Transaction, error) {
	// Buffer in whic all the extrapolated transaction are saved
	parsed := []Transaction{}
	// Based upon the possible expression tyoe extrapolates the data needed
	switch typedStmt := stmt.(type) {
	case *ast.AssignStmt:
		// The assign statement allow for more expression inside it
		for _, rValue := range typedStmt.Rhs {
			transaction := parseRecvExpr(rValue, currentState)
			// The expression isn't a recv from a channel
			if transaction.IdentName == "" {
				continue
			}
			// If the transaction is valid append it to the slice
			parsed = append(parsed, transaction)
		}
	case *ast.ExprStmt:
		transaction := parseRecvExpr(typedStmt.X, currentState)
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

// This function takes a Expr interface and tries to extrapolate the transaction
// data from the send operation (if existent) else return an invalid transaction
// (with identName equal to "")
func parseRecvExpr(expr ast.Expr, currentState *int) Transaction {
	// Checks if the given its a unary expression
	unaryExpr, isUnaryExpr := expr.(*ast.UnaryExpr)
	if !isUnaryExpr || unaryExpr.Op != token.ARROW {
		return Transaction{}
	}

	// Checks if the nested expression its an identifier (the channel name)
	chanIdent, isIdent := unaryExpr.X.(*ast.Ident)
	if !isIdent {
		return Transaction{}
	}
	// Creates a valid transaction struct
	transaction := Transaction{Recv, chanIdent.Name, Unknown, Unknown}
	// Add state transaction to the automata fragment for the function
	transaction.From = *currentState
	(*currentState)++
	transaction.To = *currentState

	return transaction
}

// Specific function to extrapolate channel metadata from an AssignStmt
func parseAssignStmt(stmt *ast.AssignStmt) ([]ChannelMetadata, error) {
	// A Slice containing all the metadata retrieved about the channel declared
	metadata := []ChannelMetadata{}

	// Check that the number of rvalues (variable assigned) are the same of
	// lvalues (values assignments) in the statement (TODO TEST)
	if len(stmt.Lhs) != len(stmt.Rhs) {
		err := errors.New("should receive same number of l_val and r_val")
		return nil, err
	}

	// Now iterates over the assignment statements
	for i := range stmt.Lhs {
		lVal, rVal := stmt.Lhs[i], stmt.Rhs[i]
		// Extrapolates all the metadata needed
		channelName := lVal.(*ast.Ident).Name

		newMetadata, err := extractMetadata(channelName, rVal)
		if err == nil {
			metadata = append(metadata, newMetadata)
		}

	}

	return metadata, nil
}

// Specific function to extrapolate channel metadata from a GenDecl statement
func parseDeclStmt(stmt *ast.DeclStmt) ([]ChannelMetadata, error) {
	// Tries to cast the current statement's declaration to a GenDecl. At the moment
	// of writing this should always be possible since only GenDecl satisfy the Decl interface
	// however this may change in future releases og Go
	genDecl, isGenDecl := stmt.Decl.(*ast.GenDecl)

	if !isGenDecl {
		return nil, errors.New("the declaration in the statement isn't a GenDecl")
	}

	return parseGenDecl(genDecl)
}

func parseGenDecl(genDecl *ast.GenDecl) ([]ChannelMetadata, error) {
	// A Slice containing all the metadata retrieved about the channel declared
	metadata := []ChannelMetadata{}

	for _, specVal := range genDecl.Specs {
		// When the token is VAR or CONST then Specs is a ValueSpec
		valueSpec, isValueSpec := specVal.(*ast.ValueSpec)
		if (genDecl.Tok == token.CONST || genDecl.Tok == token.VAR) && isValueSpec {
			// Check that the number of rvalues (variable assigned) are the same of
			// lvalues (values assignments) in the statement (TODO TEST)
			if len(valueSpec.Values) != len(valueSpec.Names) {
				err := errors.New("should receive same number of l_val and r_val")
				return nil, err
			}

			// Now iterates over the assignment statements
			for i := range valueSpec.Values {
				lVal, rVal := valueSpec.Names[i], valueSpec.Values[i]
				newMetadata, err := extractMetadata(lVal.Name, rVal)
				if err == nil {
					metadata = append(metadata, newMetadata)
				}

			}
		}
	}

	return metadata, nil
}

// Takes the name of the lVal and the rVal that is a generic expression, then
// if the rVal is a "make" call that specifically initializes a channel then
// it creates the respective ChannelMetadata record, else returns error
// NOTE: this function is shared between parseAssignStmt e parseGenDecl
// TODO CHECK IT WITH make([]int, 0) to create a list
func extractMetadata(chanName string, rVal ast.Expr) (ChannelMetadata, error) {
	// Since is posible to init a channel only with a "make" function call
	// we just want to consider Rhs expression that are function call
	callExpr, isCallExpr := rVal.(*ast.CallExpr)
	// If the Rhs expression isn't a function call we skip the current iteration
	// we skip as well if the function called has no arguments since we're only
	// interested in this case in "make" call that always has at least one param
	if !isCallExpr || len(callExpr.Args) <= 0 {
		return ChannelMetadata{}, errors.New("cannot parse, not a func")
	}

	// Then we're interested in filtering only the "make" call used to create
	// a channel (since it can be used to create array and slices as well)
	funcName, isIdentifier := callExpr.Fun.(*ast.Ident)
	channelTypeExpr, isChannelType := callExpr.Args[0].(*ast.ChanType)

	// If every condition matches we're in front of a new channel declaration
	if isIdentifier && isChannelType && funcName.Name == "make" {
		// Extrapolates all the metadata needed
		channelType := channelTypeExpr.Value.(*ast.Ident).Name
		isChannelBuffered := len(callExpr.Args) > 1

		// Creates a new metadata struct that is then added to the previosly declared slice
		return ChannelMetadata{chanName, channelType, isChannelBuffered}, nil
	}

	// Should not reach here
	return ChannelMetadata{}, errors.New("cannot parse, probably not a make func")
}
