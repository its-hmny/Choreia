// Copyright Enea Guidi (hmny).

// This package handles the parsing of a given *ast.File which represents
// the content of a Go source file as an Abstract Syntax Tree.

// The only method avaiable from the outside is ParseGenDecl, ParseDeclStmt, ParseSendStmt,
// ParseRecvStmt and ParseSelectStmt which will add to the given FileMetadata argument the
// data collected from the parsing phases
package parser

import (
	"fmt"
	"go/ast"
	"go/token"
	"log"
)

// ----------------------------------------------------------------------------
// ChanMetadata

// A ChanMetadata contains the metadata avaiable about a Go channel
//
// A struct containing all the metadata that the algorithm has been able to
// extrapolate from a channel declaration or assignment. Only the channel declared
// in the file by the user are evaluated (channel returned from external functions are ignored)
type ChanMetadata struct {
	Name  string
	Type  string
	Async bool
}

// ----------------------------------------------------------------------------
// Channel related parsing method

// This function parses a SendStmt statement and saves the Transition(s) data extracted
// in the given FuncMetadata argument. In case of error during execution no error is returned.
func ParseSendStmt(stmt *ast.SendStmt, fm *FuncMetadata) {
	chanIdent, isIdent := stmt.Chan.(*ast.Ident)
	if isIdent {
		tSend := Transition{Kind: Send, IdentName: chanIdent.Name}
		fm.PartialAutomata.AddTransition(Current, NewNode, tSend)
	} else {
		log.Fatalf("Could't find identifier in SendStmt at line: %d\n", stmt.Pos())
	}
}

// This function parses a UnaryExpr statement and saves the Transition(s) data extracted
// in the given FuncMetadata argument. In case of error during execution no error is returned.
// It search for Recv transition (receive from a channel)
func ParseRecvStmt(expr *ast.UnaryExpr, fm *FuncMetadata) {
	// Tries to extract the identifier of the expression
	chanIdent, isIdent := expr.X.(*ast.Ident)

	// If an ident isn't found or the token is not "<-" then we return
	// the current isn't a ReceiveStmt
	if !isIdent || expr.Op != token.ARROW {
		return
	}

	// Creates a valid transaction struct
	tRecv := Transition{Kind: Recv, IdentName: chanIdent.Name}
	fm.PartialAutomata.AddTransition(Current, NewNode, tRecv)
}

// This function parses a SelectStmt statement and saves the Transition(s) data extracted
// in the given FuncMetadata argument. In case of error during execution no error is returned.
func ParseSelectStmt(stmt *ast.SelectStmt, fm *FuncMetadata) {
	// Saves a local copy of the current id, all the branch will fork from it
	currentAutomataId := fm.PartialAutomata.GetLastId()
	// The id of the state in which all the nested scopes will be merged, will converge
	// when -2 is to be considered uninitialized , will be initialized correctly on first iteration
	mergeStateId := NewNode

	for i, bodyStmt := range stmt.Body.List {
		// Convert the bodyStmt to a CommClause one, this is always possible at the moment
		// since we're parsing a "select" statement and this is the only option avaiable
		commClause := bodyStmt.(*ast.CommClause)

		// Generate an eps-transition to represent the fork/branch (the cases in the select)
		// and add it as a transaction from the "branch point" saved before
		startLabel := fmt.Sprintf("select-case-%d-start", i)
		tEpsStart := Transition{Kind: Eps, IdentName: startLabel}
		fm.PartialAutomata.AddTransition(currentAutomataId, NewNode, tEpsStart)

		// Parses the clause (case stmt) before and then parses the nested block/scopes
		ast.Walk(fm, commClause)

		// Generates a transition to return/merge to the "main" scope
		endLabel := fmt.Sprintf("select-case-%d-end", i)
		tEpsEnd := Transition{Kind: Eps, IdentName: endLabel}

		if mergeStateId == NewNode {
			// Saves the id, of the merge state for use in next iterations
			fm.PartialAutomata.AddTransition(Current, NewNode, tEpsEnd)
			mergeStateId = fm.PartialAutomata.GetLastId()
		} else {
			fm.PartialAutomata.AddTransition(Current, mergeStateId, tEpsEnd)
		}
	}

	// Set the new root of the PartialAutomata, from which all future transition will start
	fm.PartialAutomata.SetRootId(mergeStateId)
}

// Specific function to extrapolate channel metadata from a DeclStmt statement
// At the moment of writing this should always be possible since only GenDecl
// satisfy the Decl interface however this may change in future releases of Go
func ParseDeclStmt(stmt *ast.DeclStmt, fm *FuncMetadata) {
	// Tries to cast the current statement's declaration to a GenDecl.
	genDecl, isGenDecl := stmt.Decl.(*ast.GenDecl)

	if !isGenDecl {
		log.Fatalf("Couldn't get the GenDecl statement fron the DeclStmt at line %d\n", stmt.Pos())
	}

	chanMeta := ParseGenDecl(genDecl)
	fm.addChannels(chanMeta...)
}

// This function tries to extract metadata about a channel from the GenDecl subtree
// since is possible to declare more than value the function returns a slice of ChanMetadata
// If errors are encountered at any point the function returns nil
func ParseGenDecl(genDecl *ast.GenDecl) []ChanMetadata {
	// A Slice containing all the metadata retrieved about the channel declared
	bufferMetadata := []ChanMetadata{}

	// Iterates over the list of ident-value association
	for _, specVal := range genDecl.Specs {
		valueSpec, isValueSpec := specVal.(*ast.ValueSpec)

		if (genDecl.Tok != token.CONST && genDecl.Tok != token.VAR) || !isValueSpec {
			// When the token is VAR or CONST then Specs is a ValueSpec (with a value assigned)
			// this is what we're interested in when looking for channel declaration
			return nil
		} else if len(valueSpec.Values) != len(valueSpec.Names) {
			// Check that the number of rvalues and lvalues are the same
			return nil
		}

		// Now iterates over the assignment statements
		for i := range valueSpec.Values {
			lVal, rVal := valueSpec.Names[i], valueSpec.Values[i]
			callExpr, isCallExpr := rVal.(*ast.CallExpr)
			// If the Rhs expression is a function call then is possible is a "make call"
			if isCallExpr {
				newChan := parseMakeCall(callExpr, lVal.Name)
				bufferMetadata = append(bufferMetadata, newChan)
			}
		}
	}

	return bufferMetadata
}

// This function tries to parse a "make" function call in order to extract metadata
// about the initialized channel, if at any point errors are encountered then the
// function returns the zero value of the ChanMetadata struct
func parseMakeCall(callExpr *ast.CallExpr, chanName string) ChanMetadata {
	// Tries to extract the function name (identifier), else return a zero value
	funcIdent, isIdent := callExpr.Fun.(*ast.Ident)

	if !isIdent {
		return ChanMetadata{}
	}

	// If we're considering a make function call we ignore the Transition and try
	// to extract some data about an eventual channel declared
	if funcIdent.Name == "make" {
		// If the first argument is a ChanType we're initializing a channel
		channelTypeExpr, isChannelType := callExpr.Args[0].(*ast.ChanType)
		if isChannelType {
			// Extrapolates all the metadata needed about the chan
			channelType := channelTypeExpr.Value.(*ast.Ident).Name
			isChannelBuffered := len(callExpr.Args) > 1
			// The name is empty and has to be set from the caller function
			return ChanMetadata{Name: chanName, Type: channelType, Async: isChannelBuffered}
		}
	}

	return ChanMetadata{}
}
