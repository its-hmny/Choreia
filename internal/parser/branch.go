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
)

// ----------------------------------------------------------------------------
// Branching/Conditional constructs related parsing method

// This function parses a IfStmt statement and saves the data extracted in a FuncMetadata struct.
// In case of error during execution a zero value of abovesaid struct is returned (no error returned).
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

// This function parses a SwitchStmt statement and saves the data extracted in a FuncMetadata struct.
// In case of error during execution a zero value of abovesaid struct is returned (no error returned).
func ParseSwitchStmt(stmt *ast.SwitchStmt, fm *FuncMetadata) {
	// First parses the init and tag sections, that are always executed before branching
	ast.Walk(fm, stmt.Init)
	ast.Walk(fm, stmt.Tag)

	// Saves a local copy of the current id, all the branch will fork from it
	currentAutomataId := fm.PartialAutomata.GetLastId()
	// The id of the state in which all the nested scopes will be merged, will converge
	// when -2 is to be considered uninitialized , will be initialized correctly on first iteration
	mergeStateId := NewNode

	for i, bodyStmt := range stmt.Body.List {
		// Convert the bodyStmt to a CaseClause one, this is always possible at the moment
		// since we're parsing a "switch" statement and this is the only option avaiable
		caseClauseStmt := bodyStmt.(*ast.CaseClause)

		// Generate an eps-transition to represent the fork/branch (the cases in the select)
		// and add it as a transaction from the "branch point" saved before
		startLabel := fmt.Sprintf("switch-case-%d-start", i)
		tEpsStart := Transition{Kind: Eps, IdentName: startLabel}
		fm.PartialAutomata.AddTransition(currentAutomataId, NewNode, tEpsStart)

		// Parses the clause (case stmt) before and then parses the nested block/scopes
		ast.Walk(fm, caseClauseStmt)

		// Generates a transition to return/merge to the "main" scope
		endLabel := fmt.Sprintf("switch-case-%d-end", i)
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

// This function parses a TypeSwitchStmt statement and saves the data extracted in a FuncMetadata struct.
// In case of error during execution a zero value of abovesaid struct is returned (no error returned).
func ParseTypeSwitchStmt(stmt *ast.TypeSwitchStmt, fm *FuncMetadata) {
	// First parses the init and tag sections, that are always executed before branching
	ast.Walk(fm, stmt.Init)
	ast.Walk(fm, stmt.Assign)

	// Saves a local copy of the current id, all the branch will fork from it
	currentAutomataId := fm.PartialAutomata.GetLastId()
	// The id of the state in which all the nested scopes will be merged, will converge
	// when -2 is to be considered uninitialized , will be initialized correctly on first iteration
	mergeStateId := NewNode

	for i, bodyStmt := range stmt.Body.List {
		// Convert the bodyStmt to a CaseClause one, this is always possible at the moment
		// since we're parsing a "switch" statement and this is the only option avaiable
		caseClauseStmt := bodyStmt.(*ast.CaseClause)

		// Generate an eps-transition to represent the fork/branch (the cases in the select)
		// and add it as a transaction from the "branch point" saved before
		startLabel := fmt.Sprintf("typeswitch-case-%d-start", i)
		tEpsStart := Transition{Kind: Eps, IdentName: startLabel}
		fm.PartialAutomata.AddTransition(currentAutomataId, NewNode, tEpsStart)

		// Parses the clause (case stmt) before and then parses the nested block/scopes
		ast.Walk(fm, caseClauseStmt)

		// Generates a transition to return/merge to the "main" scope
		endLabel := fmt.Sprintf("typeswitch-case-%d-end", i)
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

// ! Refactor the ParseTypeSwitchStmt and ParseSwitchSmtt functions