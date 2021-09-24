package utils

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
	name   string
	typing string
	async  bool
}

// Is possible to declare a variable directly both with the
// := notation (AssignStmt) or with a more verbose syntax
// var chan = make(...) (GenDecl) this method handles both
func GetChanMetadata(node ast.Node) []ChannelMetadata {
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

// This function extrapolates a compliant transaction struct from a send statement
// (obviously invoked on a channel). If at any point an error is encountered the func
// bails out returning an error.
// NOTE: the currentState pointer should not be nil
func GetSendTransaction(stmt *ast.SendStmt, currentState *int) (Transaction, error) {
	chanIdent, isIdent := stmt.Chan.(*ast.Ident)
	if isIdent {
		transaction := Transaction{Send, chanIdent.Name, Unknown, Unknown}
		// Add state transaction to the automata fragment for the function
		transaction.from = *currentState
		(*currentState)++
		transaction.to = *currentState
		// At last returns the transaction
		return transaction, nil
	}
	return Transaction{}, errors.New("GetSendTransaction: the channel isn't an identifier")
}

// This function extrapolates a compliant transaction struct from a recv statement.
// At the moment there are only 2 type of stmt that could contain this send operation
// inside it, both cases are handled in this function
// If at any point an error is encountered the func bails out returning an error.
// NOTE: the currentState pointer should not be nil
func GetRecvTransaction(stmt ast.Stmt, currentState *int) ([]Transaction, error) {
	// Buffer in whic all the extrapolated transaction are saved
	parsed := []Transaction{}
	// Based upon the possible expression tyoe extrapolates the data needed
	switch typedStmt := stmt.(type) {
	case *ast.AssignStmt:
		// The assign statement allow for more expression inside it
		for _, rValue := range typedStmt.Rhs {
			transaction := parseRecvExpr(rValue, currentState)
			// The expression isn't a recv from a channel
			if transaction.identName == "" {
				continue
			}
			// If the transaction is valid append it to the slice
			parsed = append(parsed, transaction)
		}
	case *ast.ExprStmt:
		transaction := parseRecvExpr(typedStmt.X, currentState)
		// The expression isn't a recv from a channel
		if transaction.identName == "" {
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
	transaction := Transaction{Receive, chanIdent.Name, Unknown, Unknown}
	// Add state transaction to the automata fragment for the function
	transaction.from = *currentState
	(*currentState)++
	transaction.to = *currentState

	return transaction
}

// Specific function to extrapolate channel metadata from an AssignStmt
func parseAssignStmt(stmt *ast.AssignStmt) ([]ChannelMetadata, error) {
	// A Slice containing all the metadata retrieved about the channel declared
	metadata := []ChannelMetadata{}

	// Check that the number of rvalues (variable assigned) are the same of
	// lvalues (values assignments) in the statement (TODO TEST)
	if len(stmt.Lhs) != len(stmt.Rhs) {
		err := errors.New("parseAssignmentStmt: should receive same number of l_val and r_val")
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
func parseGenDecl(stmt *ast.GenDecl) ([]ChannelMetadata, error) {
	// A Slice containing all the metadata retrieved about the channel declared
	metadata := []ChannelMetadata{}

	for _, specVal := range stmt.Specs {
		// When the token is VAR or CONST then Specs is a ValueSpec
		valueSpec, isValueSpec := specVal.(*ast.ValueSpec)
		if (stmt.Tok == token.CONST || stmt.Tok == token.VAR) && isValueSpec {
			// Check that the number of rvalues (variable assigned) are the same of
			// lvalues (values assignments) in the statement (TODO TEST)
			if len(valueSpec.Values) != len(valueSpec.Names) {
				err := errors.New("parseGenDecl: should receive same number of l_val and r_val")
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
	if !isCallExpr {
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
