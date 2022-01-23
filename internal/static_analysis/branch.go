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

	"github.com/its-hmny/Choreia/internal/data_structures/fsa"
)

// ----------------------------------------------------------------------------
// Branching/Conditional constructs related parsing method

// This function parses a IfStmt statement and saves the data extracted in a FuncMetadata struct.
// In case of error during execution a zero value of abovesaid struct is returned (no error returned).
func parseIfStmt(stmt *ast.IfStmt, fm *FuncMetadata) {
	// First parses the init statement that is always executed before branching
	ast.Walk(fm, stmt.Init)

	// Saves a local copy of the current id, all the branch will fork from it
	branchingStateId := fm.ScopeAutomata.GetLastId()

	// Generate an eps-transition to represent the creation of a new nested scope
	tEpsIfStart := fsa.Transition{Move: fsa.Eps, Label: "if-block-start"}
	fm.ScopeAutomata.AddTransition(branchingStateId, fsa.NewState, tEpsIfStart)
	// Then parses both the condition and the nested scope (if-then)
	ast.Walk(fm, stmt.Cond)
	ast.Walk(fm, stmt.Body)
	// Generates a transition to return/merge to the "main" scope
	tEpsIfEnd := fsa.Transition{Move: fsa.Eps, Label: "if-block-end"}
	fm.ScopeAutomata.AddTransition(fsa.Current, fsa.NewState, tEpsIfEnd)

	// Saves the id of the latest created states (the one in which the 2+ scopes will be merged)
	mergeStateId := fm.ScopeAutomata.GetLastId()

	// If an else block is specified then its parsed on its own branch
	if stmt.Else != nil {
		tEpsElseStart := fsa.Transition{Move: fsa.Eps, Label: "else-block-start"}
		fm.ScopeAutomata.AddTransition(branchingStateId, fsa.NewState, tEpsElseStart)
		// Parses the else block
		ast.Walk(fm, stmt.Else)
		// Links the else-block-end to the same destination as the if-block-end
		tEpsElseEnd := fsa.Transition{Move: fsa.Eps, Label: "else-block-end"}
		fm.ScopeAutomata.AddTransition(fsa.Current, mergeStateId, tEpsElseEnd)
	} else { // Else the if statement can be skipped the execution flow stays on the main branch
		tEpsIfSkip := fsa.Transition{Move: fsa.Eps, Label: "if-block-skip"}
		fm.ScopeAutomata.AddTransition(branchingStateId, mergeStateId, tEpsIfSkip)
	}

	// Set the new root of the PartialAutomata, from which all future transition will start
	fm.ScopeAutomata.SetRootId(mergeStateId)
}

// This function parses a SwitchStmt statement and saves the data extracted in a FuncMetadata struct.
// In case of error during execution a zero value of abovesaid struct is returned (no error returned).
func parseSwitchStmt(stmt *ast.SwitchStmt, fm *FuncMetadata) {
	// First parses the init and tag sections, that are always executed before branching
	ast.Walk(fm, stmt.Init)
	ast.Walk(fm, stmt.Tag)

	// Saves a local copy of the current id, all the branch will fork from it
	currentAutomataId := fm.ScopeAutomata.GetLastId()
	// The id of the state in which all the nested scopes will be merged, will converge
	// when -2 is to be considered uninitialized , will be initialized correctly on first iteration
	mergeStateId := fsa.NewState

	for i, bodyStmt := range stmt.Body.List {
		// Convert the bodyStmt to a CaseClause one, this is always possible at the moment
		// since we're parsing a "switch" statement and this is the only option available
		caseClauseStmt := bodyStmt.(*ast.CaseClause)

		// Generate an eps-transition to represent the fork/branch (the cases in the select)
		// and add it as a transaction from the "branch point" saved before
		startLabel := fmt.Sprintf("switch-case-%d-start", i)
		tEpsStart := fsa.Transition{Move: fsa.Eps, Label: startLabel}
		fm.ScopeAutomata.AddTransition(currentAutomataId, fsa.NewState, tEpsStart)

		// Parses the clause (case stmt) before and then parses the nested block/scopes
		ast.Walk(fm, caseClauseStmt)

		// Generates a transition to return/merge to the "main" scope
		endLabel := fmt.Sprintf("switch-case-%d-end", i)
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

// This function parses a TypeSwitchStmt statement and saves the data extracted in a FuncMetadata struct.
// In case of error during execution a zero value of abovesaid struct is returned (no error returned).
func parseTypeSwitchStmt(stmt *ast.TypeSwitchStmt, fm *FuncMetadata) {
	// First parses the init and tag sections, that are always executed before branching
	ast.Walk(fm, stmt.Init)
	ast.Walk(fm, stmt.Assign)

	// Saves a local copy of the current id, all the branch will fork from it
	currentAutomataId := fm.ScopeAutomata.GetLastId()
	// The id of the state in which all the nested scopes will be merged, will converge
	// when -2 is to be considered uninitialized , will be initialized correctly on first iteration
	mergeStateId := fsa.NewState

	for i, bodyStmt := range stmt.Body.List {
		// Convert the bodyStmt to a CaseClause one, this is always possible at the moment
		// since we're parsing a "switch" statement and this is the only option available
		caseClauseStmt := bodyStmt.(*ast.CaseClause)

		// Generate an eps-transition to represent the fork/branch (the cases in the select)
		// and add it as a transaction from the "branch point" saved before
		startLabel := fmt.Sprintf("typeswitch-case-%d-start", i)
		tEpsStart := fsa.Transition{Move: fsa.Eps, Label: startLabel}
		fm.ScopeAutomata.AddTransition(currentAutomataId, fsa.NewState, tEpsStart)

		// Parses the clause (case stmt) before and then parses the nested block/scopes
		ast.Walk(fm, caseClauseStmt)

		// Generates a transition to return/merge to the "main" scope
		endLabel := fmt.Sprintf("typeswitch-case-%d-end", i)
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

// ! Refactor the ParseTypeSwitchStmt and ParseSwitchStmt functions
