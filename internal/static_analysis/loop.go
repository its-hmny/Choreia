// Copyright 2020 Enea Guidi (hmny). All rights reserved.
// This file are distributed under the General Public License v 3.0.
// A copy of abovesaid license can be found in the LICENSE file.

// Package static_analysis declares the types used to represent metedata extracted from the Go source code.
// The source code is transformed to an Abstract Syntax Tree via go/ast module and. Said AST is visited fully
// and all the metadata needed are extractred then returned in a single aggregate struct.
//
package static_analysis

import (
	"go/ast"

	"github.com/its-hmny/Choreia/internal/data_structures/fsa"
)

// ----------------------------------------------------------------------------
// Looping/Iteration constructs related parsing method

// This function parses a ForStmt statement and saves the Transition(s) data extracted
// in the given FuncMetadata argument. In case of error during execution no error is returned.
func parseForStmt(stmt *ast.ForStmt, fm *FuncMetadata) {
	// Parse the init statement at first and the condition (always executed at least one time)
	ast.Walk(fm, stmt.Init)
	ast.Walk(fm, stmt.Cond) // ? parse BinaryExpr to find transition inside
	// Saves a local copy of the current id, all the branch will fork from it
	forkStateId := fm.ScopeAutomata.GetLastId()

	// Generate an eps-transition to represent the fork/branch (the iteration scope in the for loop)
	// and add it as a transaction from the "fork point" saved before
	tEpsStart := fsa.Transition{Move: fsa.Eps, Label: "for-iteration-start"}
	fm.ScopeAutomata.AddTransition(forkStateId, fsa.NewState, tEpsStart)

	// Parses the nested block (and then) the post iteration statement
	ast.Walk(fm, stmt.Body)
	ast.Walk(fm, stmt.Post)

	// Links back the iteration block to the fork state
	tEpsEnd := fsa.Transition{Move: fsa.Eps, Label: "for-iteration-end"}
	fm.ScopeAutomata.AddTransition(fsa.Current, forkStateId, tEpsEnd)
	// Links the fork state to a new one (this represents the no-iteration or exit-iteration cases)
	tEpsSkip := fsa.Transition{Move: fsa.Eps, Label: "for-iteration-skip"}
	fm.ScopeAutomata.AddTransition(forkStateId, fsa.NewState, tEpsSkip)
}

// This function parses a RangeStmt statement and saves the data extracted in a FuncMetadata struct.
// In case of error during execution no error is returned. If the identifier on which we're iterating
// is a channel then the range function behaves as a for loop in which we're receiving from the channel
// before each iteration, else (if we're iterating on a map or list) an eps-transition is used instead
func parseRangeStmt(stmt *ast.RangeStmt, fm *FuncMetadata) {
	// Parse the init statement at first and the condition (always executed at least one time)
	iterateeIdent, isIdent := stmt.X.(*ast.Ident)
	// Flag to set if the iteratee is a local channel identifier
	matchFound := false

	// Checks if the iteratee identifier is a locally declared channel, eventually sets a flag
	// this is needs because "ranging" over a channel is equal to receiving multiple time from it
	if isIdent {
		for _, chanMeta := range fm.ChanMeta { // ? add support for global channel
			if chanMeta.Name == iterateeIdent.Name {
				matchFound = true
			}
		}
		for argName := range fm.InlineArgs {
			if argName == iterateeIdent.Name {
				matchFound = true
			}
		}
	}

	// Saves a local copy of the current id, all the branch will fork from it
	forkStateId := fm.ScopeAutomata.GetLastId()

	// Generate an eps-transition to represent the fork/branch (the iteration block in the loop)
	// and add it as a transaction, if we're using range on a channel then the transition became
	// a Recv transition since on channel this is the default overload of "range" keyword
	if matchFound {
		channelMeta := fm.ChanMeta[iterateeIdent.Name]
		tRecvStart := fsa.Transition{Move: fsa.Recv, Label: iterateeIdent.Name, Payload: channelMeta}
		fm.ScopeAutomata.AddTransition(fsa.Current, fsa.NewState, tRecvStart)
	} else {
		tEpsStart := fsa.Transition{Move: fsa.Eps, Label: "range-iteration-start"}
		fm.ScopeAutomata.AddTransition(fsa.Current, fsa.NewState, tEpsStart)
	}

	// Parses the nested block
	ast.Walk(fm, stmt.Body)

	// Links back the iteration block to the fork state
	tEpsEnd := fsa.Transition{Move: fsa.Eps, Label: "range-iteration-end"}
	fm.ScopeAutomata.AddTransition(fsa.Current, forkStateId, tEpsEnd)
	// Links the fork state to a new one (this represents the no-iteration or exit-iteration cases)
	tEpsSkip := fsa.Transition{Move: fsa.Eps, Label: "range-iteration-skip"}
	fm.ScopeAutomata.AddTransition(forkStateId, fsa.NewState, tEpsSkip)
}
