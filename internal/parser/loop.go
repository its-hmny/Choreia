// Copyright Enea Guidi (hmny).

// This package handles the parsing of a given *ast.File which represents
// the content of a Go source file as an Abstract Syntax Tree.

// TODO add subsection comment
package parser

import (
	"go/ast"
)

// ----------------------------------------------------------------------------
// Looping/Iteration constructs related parsing method

// TODO COMMENT
//
//
func ParseForStmt(stmt *ast.ForStmt, fm *FuncMetadata) {
	// Parse the init statement at first and the condition (always executed at least one time)
	ast.Walk(fm, stmt.Init)
	ast.Walk(fm, stmt.Cond) // TODO parse BinaryExpr to find transition inside
	// Saves a local copy of the current id, all the branch will fork from it
	forkStateId := fm.PartialAutomata.GetLastId()

	// Generate an eps-transition to represent the fork/branch (the iteration scope in the for loop)
	// and add it as a transaction from the "fork point" saved before
	tEpsStart := Transition{Kind: Eps, IdentName: "for-iteration-start"}
	fm.PartialAutomata.AddTransition(forkStateId, NewNode, tEpsStart)

	// Parses the nested block (and then) the post iteration statement
	ast.Walk(fm, stmt.Body)
	ast.Walk(fm, stmt.Post)

	// Links back the iteration block to the fork state
	tEpsEnd := Transition{Kind: Eps, IdentName: "for-iteration-end"}
	fm.PartialAutomata.AddTransition(Current, forkStateId, tEpsEnd)
	// Links the fork state to a new one (this represents the no-iteration or exit-iteration cases)
	tEpsSkip := Transition{Kind: Eps, IdentName: "for-iteration-skip"}
	fm.PartialAutomata.AddTransition(forkStateId, NewNode, tEpsSkip)
}

// TODO COMMENT
//
//
func ParseRangeStmt(stmt *ast.RangeStmt, fm *FuncMetadata) {
	// Parse the init statement at first and the condition (always executed at least one time)
	iterateeIdent, isIdent := stmt.X.(*ast.Ident)
	// Flag to set if the iteratee is a local channel identifier
	matchFound := false

	// Checks if the iteratee identifier is a locally declared channel, eventually sets a flag
	// this is neede because "ranging" over a channel is equal to receiving multiple time from it
	if isIdent {
		for _, chanMeta := range fm.ChanMeta {
			if chanMeta.Name == iterateeIdent.Name {
				matchFound = true
			}
		}
	}

	// Generate an eps-transition to represent the fork/branch (the iteration block in the loop)
	// and add it as a transaction, if we're using range on a channel then the transition becames
	// a Recv trnasition since on channel this is the default overload of "range" keyword
	if matchFound {
		tEpsStart := Transition{Kind: Recv, IdentName: iterateeIdent.Name}
		fm.PartialAutomata.AddTransition(Current, NewNode, tEpsStart)
	} else {
		tEpsStart := Transition{Kind: Eps, IdentName: "range-iteration-start"}
		fm.PartialAutomata.AddTransition(Current, NewNode, tEpsStart)
	}

	// Saves a local copy of the current id, all the branch will fork from it
	forkStateId := fm.PartialAutomata.GetLastId()
	// Parses the nested block
	ast.Walk(fm, stmt.Body)

	// Links back the iteration block to the fork state
	tEpsEnd := Transition{Kind: Eps, IdentName: "range-iteration-end"}
	fm.PartialAutomata.AddTransition(Current, forkStateId, tEpsEnd)
	// Links the fork state to a new one (this represents the no-iteration or exit-iteration cases)
	tEpsSkip := Transition{Kind: Eps, IdentName: "range-iteration-skip"}
	fm.PartialAutomata.AddTransition(forkStateId, NewNode, tEpsSkip)
}
