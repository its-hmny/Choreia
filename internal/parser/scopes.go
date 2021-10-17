// Copyright Enea Guidi (hmny).

// This package handles the parsing of a given *ast.File which represents
// the content of a Go source file as an Abstract Syntax Tree.

// The only method avaiable from the outside is ParseGenDecl, ParseDeclStmt, ParseSendStmt,
// ParseRecvStmt and ParseSelectStmt which will add to the given FileMetadata argument the
// data collected from the parsing phases
package parser

import (
	"go/ast"
)

// ----------------------------------------------------------------------------
// Nested scopes related parsing method

// TODO COMMENT
//
//
func ParseIfStmt(stmt *ast.IfStmt, fm *FuncMetadata) {
	// First parses the init statement that is always executed before branching
	ast.Walk(fm, stmt.Init)

	// Saves a local copy of the current id, all the branch will fork from it
	branchingStateId := fm.PartialAutomata.GetLastId()

	// Generate an eps-transition to represent the creation of a new nested scope
	tEpsIfStart := Transition{Kind: Eps, IdentName: "if-block-start"}
	fm.PartialAutomata.AddTransition(branchingStateId, NewNode, tEpsIfStart)
	// Then parses both the condition and the nested scope (if-then)
	ast.Walk(fm, stmt.Cond)
	ast.Walk(fm, stmt.Body)
	// Generates a transition to return/merge to the "main" scope
	tEpsIfEnd := Transition{Kind: Eps, IdentName: "if-block-end"}
	fm.PartialAutomata.AddTransition(Current, NewNode, tEpsIfEnd)

	// Saves the id of the latest created states (the one in which the 2+ scopes will be merged)
	mergeStateId := fm.PartialAutomata.GetLastId()

	// If an else block is specified then its parsed on its own branch
	tEpsElseStart := Transition{Kind: Eps, IdentName: "else-block-start"}
	fm.PartialAutomata.AddTransition(branchingStateId, NewNode, tEpsElseStart)
	// Parses the else block
	ast.Walk(fm, stmt.Else)
	// Links the else-block-end to the same destination as the if-block-end
	tEpsElseEnd := Transition{Kind: Eps, IdentName: "else-block-end"}
	fm.PartialAutomata.AddTransition(Current, mergeStateId, tEpsElseEnd)

	// Set the new root of the PartialAutomata, from which all future transition will start
	fm.PartialAutomata.SetRootId(mergeStateId)
}
