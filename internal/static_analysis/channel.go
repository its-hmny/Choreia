// Copyright 2020 Enea Guidi (hmny). All rights reserved.
// This file are distributed under the General Public License v 3.0.
// A copy of abovesaid license can be found in the LICENSE file.

// Package static_analysis declares the types used to represent metedata extracted from the Go source code.
// The source code is transformed to an Abstract Syntax Tree via go/ast module and. Said AST is visited fully
// and all the metadata needed are extractred then returned in a single aggregate struct.
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
func parseSendStmt(stmt *ast.SendStmt, fm *FuncMetadata) {
	chanIdent, isIdent := stmt.Chan.(*ast.Ident)
	if isIdent {
		tSend := fsa.Transition{Move: fsa.Send, Label: chanIdent.Name}
		fm.ScopeAutomata.AddTransition(fsa.Current, fsa.NewState, tSend)
	} else {
		log.Fatalf("Could't find identifier in SendStmt at line: %d\n", stmt.Pos())
	}
}

// This function parses a UnaryExpr statement and saves the Transition(s) data extracted
// in the given FuncMetadata argument. In case of error during execution no error is returned.
// It search for Recv transition (receive from a channel)
func parseRecvStmt(expr *ast.UnaryExpr, fm *FuncMetadata) {
	// Tries to extract the identifier of the expression
	chanIdent, isIdent := expr.X.(*ast.Ident)

	// If an ident isn't found or the token is not "<-" then we return
	// the current isn't a ReceiveStmt
	if !isIdent || expr.Op != token.ARROW {
		return
	}

	// Creates a valid transaction struct
	tRecv := fsa.Transition{Move: fsa.Recv, Label: chanIdent.Name}
	fm.ScopeAutomata.AddTransition(fsa.Current, fsa.NewState, tRecv)
}

// This function parses a SelectStmt statement and saves the Transition(s) data extracted
// in the given FuncMetadata argument. In case of error during execution no error is returned.
func parseSelectStmt(stmt *ast.SelectStmt, fm *FuncMetadata) {
	// Saves a local copy of the current id, all the branch will fork from it
	currentAutomataId := fm.ScopeAutomata.GetLastId()
	// The id of the state in which all the nested scopes will be merged, will converge
	// when -2 is to be considered uninitialized , will be initialized correctly on first iteration
	mergeStateId := fsa.NewState

	for i, bodyStmt := range stmt.Body.List {
		// Convert the bodyStmt to a CommClause one, this is always possible at the moment
		// since we're parsing a "select" statement and this is the only option available
		commClause := bodyStmt.(*ast.CommClause)

		// Generate an eps-transition to represent the fork/branch (the cases in the select)
		// and add it as a transaction from the "branch point" saved before
		startLabel := fmt.Sprintf("select-case-%d-start", i)
		tEpsStart := fsa.Transition{Move: fsa.Eps, Label: startLabel}
		fm.ScopeAutomata.AddTransition(currentAutomataId, fsa.NewState, tEpsStart)

		// Parses the clause (case stmt) before and then parses the nested block/scopes
		ast.Walk(fm, commClause)

		// Generates a transition to return/merge to the "main" scope
		endLabel := fmt.Sprintf("select-case-%d-end", i)
		tEpsEnd := fsa.Transition{Move: fsa.Eps, Label: endLabel}

		if mergeStateId == fsa.NewState {
			// Saves the id, of the merge state for use in next iterations
			fm.ScopeAutomata.AddTransition(fsa.Current, fsa.NewState, tEpsEnd)
			mergeStateId = fm.ScopeAutomata.GetLastId()
		} else {
			fm.ScopeAutomata.AddTransition(fsa.Current, mergeStateId, tEpsEnd)
		}
	}

	// Set the new root of the PartialAutomata, from which all future transition will start
	fm.ScopeAutomata.SetRootId(mergeStateId)
}

// Specific function to extrapolate channel metadata from a DeclStmt statement
// At the moment of writing this should always be possible since only GenDecl
// satisfy the Decl interface however this may change in future releases of Go
func parseDeclStmt(stmt *ast.DeclStmt, fm *FuncMetadata) {
	// Tries to cast the current statement's declaration to a GenDecl.
	genDecl, isGenDecl := stmt.Decl.(*ast.GenDecl)

	if !isGenDecl {
		log.Fatalf("Couldn't get the GenDecl statement from the DeclStmt at line %d\n", stmt.Pos())
	}

	chanMeta := parseGenDecl(genDecl)
	fm.addChannels(chanMeta...)
}

// This function tries to extract metadata about a channel from the GenDecl subtree
// since is possible to declare more than value the function returns a slice of ChanMetadata
// If errors are encountered at any point the function returns nil
func parseGenDecl(genDecl *ast.GenDecl) []ChanMetadata {
	// A Slice containing all the metadata retrieved about the channel declared
	bufferMetadata := []ChanMetadata{}

	// Iterates over the list of ident-value association
	for _, specVal := range genDecl.Specs {
		valueSpec, isValueSpec := specVal.(*ast.ValueSpec)

		if (genDecl.Tok != token.CONST && genDecl.Tok != token.VAR) || !isValueSpec {
			// When the token is VAR or CONST then Specs is a ValueSpec (with a value assigned)
			// this is what we're interested in when looking for channel declaration
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
