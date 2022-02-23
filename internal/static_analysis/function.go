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

const (
	Function ArgType = iota // Possible value of FuncArg.type
	Channel
	anonymousFunc = "anonymousFunc" // Constant to identify anonymous function
)

// ----------------------------------------------------------------------------
// FuncMetadata

// A FuncMetadata contains the metadata available about a Go function
//
// A struct containing all the metadata that the algorithm has been able to
// extrapolate from the function declaration. Only the function declared in the file
// by the user are evaluated (built-in and external functions are ignored)
type FuncMetadata struct {
	Name          string                  // The identifier of the function
	ChanMeta      map[string]ChanMetadata // The channels available inside the function scope
	InlineArgs    map[string]FuncArg      // The argument of the function to be inlined (Callbacks/Functions or Channels)
	ScopeAutomata *fsa.FSA                // A graph representing the transition made inside the function body
}

type FuncArg struct {
	Offset int     // The position of the arg in the function declaration
	Name   string  // The identifier of the argument inside the function
	Type   ArgType // The type of the argument (only Function or Channel)
}

type ArgType int // Enum of the arguments type that we're interested in

// Adds the given metadata about some channel(s) to the FuncMetadata struct
// In case a channel with the same name already exist then the previous association
// is overwritten, this is correct since the channel name is the variable to which
// the channel is assigned and this means that a new assignment was made to that variable
func (fm *FuncMetadata) addChannels(newChanMeta ...ChanMetadata) {
	// Adds or updates the associations
	for _, channel := range newChanMeta {
		// Checks the validity of the current item
		if channel.Name != "" && channel.Type != "" {
			fm.ChanMeta[channel.Name] = channel
		}
	}
}

// In order to satisfy the ast.Visitor interface FuncMetadata implements
// the Visit() method with this function signature. The Visit method takes as
// only argument an ast.Node interface and evaluates all the meaningful cases,
// when the function steps into that it tries to extract metada from the subtree
func (fm FuncMetadata) Visit(node ast.Node) ast.Visitor {
	// Skips empty nodes during descend
	if node == nil {
		return nil
	}

	switch stmt := node.(type) {
	// Handle for-range loops (e.g "for index, item := range list")
	case *ast.RangeStmt:
		parseRangeStmt(stmt, &fm)
		return nil

	// Handles for loop (the classic ones, "for i:= 0; i < 8; i++")
	case *ast.ForStmt:
		parseForStmt(stmt, &fm)
		return nil

	// Handles TypeSwitch statement (e.g "interface.(type)")
	case *ast.TypeSwitchStmt:
		parseTypeSwitchStmt(stmt, &fm)
		return nil

	// Handles switch statement
	case *ast.SwitchStmt:
		parseSwitchStmt(stmt, &fm)
		return nil

	// Handles all cases (If | If-Else | If-ElseIf-Else)
	case *ast.IfStmt:
		parseIfStmt(stmt, &fm)
		return nil

	// Statement to spawn a new Go routine
	case *ast.GoStmt:
		parseGoStmt(stmt, &fm)
		return nil

	// Statement to send or receive from multiple channel without blocking on each one
	case *ast.SelectStmt:
		parseSelectStmt(stmt, &fm)
		return nil

	// Statement to send some data on a channel
	case *ast.SendStmt:
		parseSendStmt(stmt, &fm)
		return nil

	// Statement for binary or unary expression (channel recv, function call)
	case *ast.ExprStmt:
		parseExprStmt(stmt, &fm)
		return nil

	// Statement to assign the value of an expression (chanel recv, channel decl, function call)
	case *ast.AssignStmt:
		parseAssignStmt(stmt, &fm)
		return nil

	// Statement to declare a new variable (channel decl)
	case *ast.DeclStmt:
		parseDeclStmt(stmt, &fm)
		return nil
	}
	return fm
}

// ----------------------------------------------------------------------------
// Function related parsing method

// This function parses a FuncDecl statement and saves the data extracted in a FuncMetadata struct.
// In case of strange condition (function declared in another module or C function called fromGo code)
// then no metadata are extracted and the execution will resume parsing the global scope.
func parseFuncDecl(stmt *ast.FuncDecl, fm FileMetadata) {
	// Retrieve function name and arguments
	funcName := stmt.Name.Name
	funcArgs := stmt.Type.Params.List

	// Initial setup of the metadata record
	metadata := FuncMetadata{
		Name:          funcName,
		ChanMeta:      make(map[string]ChanMetadata),
		InlineArgs:    make(map[string]FuncArg),
		ScopeAutomata: fsa.New(),
	}

	// Copies the global scope channel in the nested scope of the function.
	// Simple implementation of scope inheritance
	for name, meta := range fm.GlobalChanMeta {
		metadata.ChanMeta[name] = meta
	}

	// If the current is an external (non Go) function then is skipped since
	// it isn't useful in order to evaluate the choreography of the automon
	if stmt.Body == nil {
		return
	}

	// If the function has arguments we search for channels or callback/functions since
	// this are relevant for the Choreography Automata and must be "inlined" later on
	if len(funcArgs) > 0 {
		for i, arg := range funcArgs {
			// Extrapolates the argument name and type
			argName := arg.Names[0].Name
			_, isChannel := arg.Type.(*ast.ChanType)
			_, isFunction := arg.Type.(*ast.FuncType)

			if isChannel {
				// Adds the channel arg as "to be inlined"
				newInlineArg := FuncArg{Offset: i, Name: argName, Type: Channel}
				metadata.InlineArgs[argName] = newInlineArg
			} else if isFunction {
				// Adds the function arg as "to be inlined"
				newInlineArg := FuncArg{Offset: i, Name: argName, Type: Function}
				metadata.InlineArgs[argName] = newInlineArg
			}
		}
	}

	// Upon completion of the "setup" phase then the body of the
	// function is visited through the ast.Walk() function in order to
	// gather additional information about the stmt in the function scope
	ast.Walk(metadata, stmt.Body)

	// Adds an eps transition to a new state
	t := fsa.Transition{Move: fsa.Eps, Label: fmt.Sprintf("func-%s-return", metadata.Name)}
	metadata.ScopeAutomata.AddTransition(fsa.Current, fsa.NewState, t)
	// The newly created state will be the final state of the ScopeAutomata
	metadata.ScopeAutomata.FinalStates.Add(metadata.ScopeAutomata.GetLastId())

	// At last all the data extracted is returned
	fm.FunctionMeta[funcName] = metadata
}

// This function parses a GoStmt statement and saves the transition data extracted
// in the given FuncMetadata argument. In case of error during execution no error is returned.
func parseGoStmt(stmt *ast.GoStmt, fm *FuncMetadata) {
	// Determines if GoStmt spawns a Go routine from declared or anonymous function
	funcIdent, isFuncIdent := stmt.Call.Fun.(*ast.Ident) // Declared function
	_, isFuncAnonymous := stmt.Call.Fun.(*ast.FuncLit)   // Anonymous function

	// Then extracts the data accordingly
	if isFuncIdent {
		tSpawn := fsa.Transition{Move: fsa.Spawn, Label: funcIdent.Name}

		// Parses the GoStmt arguments looking for channels and saves the "actual" argument to list
		// in the Transition. Later this channels will be inlined during the generation of the automaton
		// ! Remove duplicate at line 253
		for i, arg := range stmt.Call.Args {
			argIdent, isIdent := arg.(*ast.Ident)
			if isIdent {
				_, isChannel := fm.ChanMeta[argIdent.Name]
				if isChannel {
					funcArgList, _ := tSpawn.Payload.([]FuncArg)
					newFuncArg := FuncArg{Offset: i, Name: argIdent.Name, Type: Channel}
					tSpawn.Payload = append(funcArgList, newFuncArg)
				}
			}
		}

		// At last add the transition (with the payload) to the ScopeAutomata
		fm.ScopeAutomata.AddTransition(fsa.Current, fsa.NewState, tSpawn)
	} else if isFuncAnonymous {
		// ToDo: This functionality is not yet implemented
		anonFuncName := fmt.Sprintf("%s-%s", anonymousFunc, fm.Name)
		tSpawn := fsa.Transition{Move: fsa.Spawn, Label: anonFuncName}
		fm.ScopeAutomata.AddTransition(fsa.Current, fsa.NewState, tSpawn)
		// ? Add parent ChanMeta (scope inheritance)
		// ? Add parse arguments (different from above)
		// ? Should parse body of funcLiteral
	}
}

// This function parses a CallExpr statement and saves the transition data extracted
// in the given FuncMetadata argument. In case of error during execution no error is returned.
func parseCallExpr(expr *ast.CallExpr, fm *FuncMetadata) {
	// Tries to extract the function name (identifier), else throw an exception
	funcIdent, isIdent := expr.Fun.(*ast.Ident)

	if !isIdent {
		// ? Consider struct.method() syntax as well (*ast.SelectorExpr)
		return
	}

	// Creates a valid transition struct
	tCall := fsa.Transition{Move: fsa.Call, Label: funcIdent.Name}

	// Parses the CallExpr arguments looking for channels and saves the "actual" argument to list
	// in the Transition. Later this channels will be inlined during the generation of the automaton
	// ! Remove duplicate at line 211
	for i, arg := range expr.Args {
		argIdent, isIdent := arg.(*ast.Ident)
		if isIdent {
			_, isChannel := fm.ChanMeta[argIdent.Name]
			if isChannel {
				funcArgList, _ := tCall.Payload.([]FuncArg)
				newFuncArg := FuncArg{Offset: i, Name: argIdent.Name, Type: Channel}
				tCall.Payload = append(funcArgList, newFuncArg)
			}
		}
	}

	// At last add full the transition to the ScopeAutomata of the FuncMetadata
	fm.ScopeAutomata.AddTransition(fsa.Current, fsa.NewState, tCall)
}
