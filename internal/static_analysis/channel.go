// Copyright 2020 Enea Guidi (hmny). All rights reserved.
// This file are distributed under the General Public License v 3.0.
// A copy of abovesaid license can be found in the LICENSE file.

// Package static_analysis declares the types used to represent metadata extracted from the Go source.
// The source code is transformed to an Abstract Syntax Tree via go/ast module.
// Said AST is visited through the Visitor pattern all the metadata available are extractred
// and agglomerated in a single comprehensive struct.
//
package static_analysis

import (
	"fmt"
	"go/ast"
	"go/token"
	"log"

	"github.com/its-hmny/Choreia/internal/data_structures/fsa"
)

// ----------------------------------------------------------------------------
// ChanMetadata

// A ChanMetadata contains the metadata available about a Go channel
//
// A struct containing all the metadata that the Visitor algorithm has been able to extrapolate.
// This kind of date are derived both from channel declaration and assignment.
// Only the channel declared in the file are evaluated (channel returned from function call or
// imported from another module are ignored)
type ChanMetadata struct {
	Name  string // The name of the channel
	Type  string // The type of message the channel supports (int, string, interface{}, ...)
	Async bool   // Is the channel unbuffered (synchronous) or buffered (asynchronous)
}

// ----------------------------------------------------------------------------
// Channel related parsing method

// This function parses a SendStmt statement and saves the transition(s) extracted
// in the given FuncMetadata argument. In case of error the whole execution is stopped.
func parseSendStmt(stmt *ast.SendStmt, fm *FuncMetadata) {
	chanIdent, isIdent := stmt.Chan.(*ast.Ident)
	if isIdent {
		channelMeta := fm.ChanMeta[chanIdent.Name]
		tSend := fsa.Transition{Move: fsa.Send, Label: chanIdent.Name, Payload: channelMeta}
		fm.ScopeAutomata.AddTransition(fsa.Current, fsa.NewState, tSend)
	} else {
		log.Fatalf("Could't find identifier in SendStmt at line: %d\n", stmt.Pos())
	}
}

// This function parses a UnaryExpr statement and saves the Transition(s) data extracted
// in the given FuncMetadata argument. In case of error during execution no error is returned.
func parseRecvStmt(expr *ast.UnaryExpr, fm *FuncMetadata) {
	// Tries to extract the identifier of the expression
	chanIdent, isIdent := expr.X.(*ast.Ident)

	// If an ident isn't found or the token is not "<-" then we return.
	// This is means the current op we're parsing isn't a ReceiveStmt
	if !isIdent || expr.Op != token.ARROW {
		return
	}

	// Retrieves the channel metadata and initializes a valid transition
	channelMeta := fm.ChanMeta[chanIdent.Name]
	tRecv := fsa.Transition{Move: fsa.Recv, Label: chanIdent.Name, Payload: channelMeta}
	fm.ScopeAutomata.AddTransition(fsa.Current, fsa.NewState, tRecv)
}

// This function parses a SelectStmt statement and saves the Transition(s) data extracted
// in the given FuncMetadata argument. In case of error during execution no error is returned.
func parseSelectStmt(stmt *ast.SelectStmt, fm *FuncMetadata) {
	// Saves a local copy of the current id, all the branch will fork from it
	currentAutomataId := fm.ScopeAutomata.GetLastId()
	// The id of the state in which all the nested scopes will converge.
	// It will be initialized correctly after the first iteration
	mergeStateId := fsa.Unknown

	for i, bodyStmt := range stmt.Body.List {
		// Convert the bodyStmt to a CommClause one, this is always possible at the moment
		// since we're parsing a "select" statement and this is the only option available
		commClause := bodyStmt.(*ast.CommClause)

		// Generate an eps-transition to represent the fork/branch (the cases in the select)
		// and add it as a transition from the "branching point" saved before
		startLabel := fmt.Sprintf("select-case-%d-start", i)
		tEpsStart := fsa.Transition{Move: fsa.Eps, Label: startLabel}
		fm.ScopeAutomata.AddTransition(currentAutomataId, fsa.NewState, tEpsStart)

		// Parses the CaseClause, then parses the nested block/scopes
		ast.Walk(fm, commClause)

		// Generates a transition to return/merge to the "main" scope
		endLabel := fmt.Sprintf("select-case-%d-end", i)
		tEpsEnd := fsa.Transition{Move: fsa.Eps, Label: endLabel}

		if mergeStateId == fsa.Unknown {
			// Saves the id, of the merge state for use in next iterations
			fm.ScopeAutomata.AddTransition(fsa.Current, fsa.NewState, tEpsEnd)
			mergeStateId = fm.ScopeAutomata.GetLastId()
		} else {
			fm.ScopeAutomata.AddTransition(fsa.Current, mergeStateId, tEpsEnd)
		}
	}

	// Set the new root of the Automaton, from which all future transition will start
	fm.ScopeAutomata.SetRootId(mergeStateId)
}

// Specific function to extrapolate channel metadata from a DeclStmt statement.
// At the moment of writing this should always be possible since only GenDecl
// satisfies the Decl interface however this may change in future releases of Go
func parseDeclStmt(stmt *ast.DeclStmt, fm *FuncMetadata) {
	// Tries to cast the current statement's declaration to a GenDecl.
	genDecl, isGenDecl := stmt.Decl.(*ast.GenDecl)

	if !isGenDecl { // This should never happen
		log.Fatalf("Couldn't get the GenDecl statement from the DeclStmt at line %d\n", stmt.Pos())
	}

	chanMeta := parseGenDecl(genDecl)
	fm.addChannels(chanMeta...)
}

// This function tries to extract metadata about a channel from the GenDecl subtree.
// Since is possible to declare more variables in a single GenDecl statement the function
// returns a slice of ChanMetadata. If errors are encountered at any point the function returns nil
func parseGenDecl(genDecl *ast.GenDecl) []ChanMetadata {
	// Initializes the slice where al the data extracted will be aggregated
	bufferMetadata := []ChanMetadata{}

	// Iterates over the list of Ident <-> Value association
	for _, specVal := range genDecl.Specs {
		valueSpec, isValueSpec := specVal.(*ast.ValueSpec)

		if (genDecl.Tok != token.CONST && genDecl.Tok != token.VAR) || !isValueSpec {
			// When the token is VAR or CONST then Specs is a ValueSpec (with a value assigned).
			// This isn't what we're interested in when looking for channel declaration
			return nil
		}

		// Now iterates over the assignment statements
		for i := range valueSpec.Values {
			// Get the right-hand-side and left-hand-side of the expression
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
// about the initialized channel. If at any point errors are encountered then the
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
