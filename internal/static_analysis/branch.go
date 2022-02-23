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

	"github.com/its-hmny/Choreia/internal/data_structures/fsa"
)

// ----------------------------------------------------------------------------
// Branching/Conditional constructs related parsing method

// This function parses a IfStmt statement and saves the data extracted in a FuncMetadata struct.
// In case of error during execution a zero value of abovesaid struct is returned (no error returned).
func parseIfStmt(stmt *ast.IfStmt, fm *FuncMetadata) {
	// First parses the init statement that is always executed before branching
	ast.Walk(fm, stmt.Init)

	// Saves a local copy of the current id.
	// All the branches in this statement will fork from it
	branchingStateId := fm.ScopeAutomata.GetLastId()

	// Generate an eps-transition to represent the creation of a new nested scope/branch
	tEpsIfStart := fsa.Transition{Move: fsa.Eps, Label: "if-block-start"}
	fm.ScopeAutomata.AddTransition(branchingStateId, fsa.NewState, tEpsIfStart)
	// Then parses both the condition and the nested scope (if-then)
	ast.Walk(fm, stmt.Cond)
	ast.Walk(fm, stmt.Body)
	// Generates a transition to return/merge to the "main" scope
	tEpsIfEnd := fsa.Transition{Move: fsa.Eps, Label: "if-block-end"}
	fm.ScopeAutomata.AddTransition(fsa.Current, fsa.NewState, tEpsIfEnd)

	// Saves the id of the latest created state
	// All the branches in this statement will converge to this
	mergeStateId := fm.ScopeAutomata.GetLastId()

	// If an else block is specified then its parsed on its own branch (2 equal branches are created)
	if stmt.Else != nil {
		tEpsElseStart := fsa.Transition{Move: fsa.Eps, Label: "else-block-start"}
		fm.ScopeAutomata.AddTransition(branchingStateId, fsa.NewState, tEpsElseStart)
		// Parses the else block
		ast.Walk(fm, stmt.Else)
		// Links the else-block-end to the same destination as the if-block-end
		tEpsElseEnd := fsa.Transition{Move: fsa.Eps, Label: "else-block-end"}
		fm.ScopeAutomata.AddTransition(fsa.Current, mergeStateId, tEpsElseEnd)
	} else {
		// If an else block isn't provided the we will have a "main" branch and the "alternative"
		// execution flow (the one in which also the if-then block is executed as well)
		tEpsIfSkip := fsa.Transition{Move: fsa.Eps, Label: "if-block-skip"}
		fm.ScopeAutomata.AddTransition(branchingStateId, mergeStateId, tEpsIfSkip)
	}

	// Set the new root of the Automaton, from which all future transition will start
	fm.ScopeAutomata.SetRootId(mergeStateId)
}

// This function parses a SwitchStmt statement and saves the data extracted in a FuncMetadata struct.
// In case of error during execution a zero value of abovesaid struct is returned (no error returned).
// ! Refactor the parseTypeSwitchStmt and parseSwitchStmt functions since they're almost equal
func parseSwitchStmt(stmt *ast.SwitchStmt, fm *FuncMetadata) {
	// First parses the init and tag sections, that are always executed before branching
	ast.Walk(fm, stmt.Init)
	ast.Walk(fm, stmt.Tag)

	// Saves the id of the latest created state
	// All the branches in this statement will converge to this
	currentAutomataId := fm.ScopeAutomata.GetLastId()
	// All the branches in this statement will converge to this state
	// The first branch to be parsed will be the one to initialize the variable with a valid id
	mergeStateId := fsa.Unknown

	for i, bodyStmt := range stmt.Body.List {
		// Convert the Stmt to a CaseClause one, this is always possible at the moment.
		// Since we're parsing a "switch" statement and this is the only option available
		caseClauseStmt := bodyStmt.(*ast.CaseClause)

		// Generate an eps-transition to represent the fork/branch (the cases in the select)
		// and add it as a transition from the "branching point" saved before
		startLabel := fmt.Sprintf("switch-case-%d-start", i)
		tEpsStart := fsa.Transition{Move: fsa.Eps, Label: startLabel}
		fm.ScopeAutomata.AddTransition(currentAutomataId, fsa.NewState, tEpsStart)

		// Parses the ClauseCase statement, then parses the nested block/scopes
		ast.Walk(fm, caseClauseStmt)

		// Generates a transition to return/merge to the main scope
		endLabel := fmt.Sprintf("switch-case-%d-end", i)
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

// This function parses a TypeSwitchStmt statement and saves the data extracted in a FuncMetadata struct.
// In case of error during execution a zero value of abovesaid struct is returned (no error returned).
// ! Refactor the parseTypeSwitchStmt and parseSwitchStmt functions since they're almost equal
func parseTypeSwitchStmt(stmt *ast.TypeSwitchStmt, fm *FuncMetadata) {
	// First parses the init and tag sections, that are always executed before branching
	ast.Walk(fm, stmt.Init)
	ast.Walk(fm, stmt.Assign)

	// Saves the id of the latest created state
	// All the branches in this statement will converge to this
	currentAutomataId := fm.ScopeAutomata.GetLastId()
	// All the branches in this statement will converge to this state
	// The first branch to be parsed will be the one to initialize the variable with a valid id
	mergeStateId := fsa.Unknown

	for i, bodyStmt := range stmt.Body.List {
		// Convert the Stmt to a CaseClause one, this is always possible at the moment.
		// Since we're parsing a "switch" statement and this is the only option available
		caseClauseStmt := bodyStmt.(*ast.CaseClause)

		// Generate an eps-transition to represent the fork/branch (the cases in the select)
		// and add it as a transition from the "branching point" saved before
		startLabel := fmt.Sprintf("typeswitch-case-%d-start", i)
		tEpsStart := fsa.Transition{Move: fsa.Eps, Label: startLabel}
		fm.ScopeAutomata.AddTransition(currentAutomataId, fsa.NewState, tEpsStart)

		// Parses the ClauseCase statement, then parses the nested block/scopes
		ast.Walk(fm, caseClauseStmt)

		// Generates a transition to return/merge to the main scope
		endLabel := fmt.Sprintf("typeswitch-case-%d-end", i)
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
