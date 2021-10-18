// Copyright Enea Guidi (hmny).

// This package handles the parsing of a given *ast.File which represents
// the content of a Go source file as an Abstract Syntax Tree.

// The only method avaiable from the outside is ParseFuncDecl, ParseGoStmt and ParseCallExpr which
// will add to the given FileMetadata argument the data collected from the parsing phases
package parser

import (
	"fmt"
	"go/ast"
)

const (
	AnonymousFunc = "anonymousFunc" // Constant to identify anonymous function

	Function = iota // Possible value of FuncArg.type
	Channel
)

// ----------------------------------------------------------------------------
// FuncMetadata

// A FuncMetadata contains the metadata avaiable about a Go function
//
// A struct containing all the metadata that the algorithm has been able to
// extrapolate from the function declaration. Only the function declared in the file
// by the user are evaluated (built-in and external functions are ignored)
type FuncMetadata struct {
	Name            string                  // The identifier of the function
	ChanMeta        map[string]ChanMetadata // The channels avaiable inside the function scope
	InlineArgs      map[string]FuncArg      // The argument of the function to be inlined (Callbacks/Functions or Channels)
	PartialAutomata *TransitionGraph        // A graph representing the transition made inside the function body
}

type FuncArg struct {
	Offset int    // The position of the arg in the function declaration
	Name   string // The identifier of the argument inside the function
	Type   uint   // The type of the argument (only Function or Channel)
}

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

// In order to satify the ast.Visitor interface FuncMetadata implements
// the Visit() method with this function signature. The Visit method takes as
// only argument an ast.Node interface and evaluates all the meaninggul cases,
// when the function steps into that it tries to extract metada from the subtree
func (fm FuncMetadata) Visit(node ast.Node) ast.Visitor {
	// Skips empty nodes during descend
	if node == nil {
		return nil
	}

	switch stmt := node.(type) {
	case *ast.ForStmt, *ast.RangeStmt:
		fmt.Printf("Meaningful statement reached: %T at position %d\n", stmt, stmt.Pos())

	case *ast.TypeSwitchStmt:
		ParseTypeSwitchStmt(stmt, &fm)
		return nil

	case *ast.SwitchStmt:
		ParseSwitchStmt(stmt, &fm)
		return nil

	// Handles all cases (If | If-Else | If-ElseIf-Else)
	case *ast.IfStmt:
		ParseIfStmt(stmt, &fm)
		return nil

	// Statement to spawn a new Go routine
	case *ast.GoStmt:
		ParseGoStmt(stmt, &fm)
		return nil

	// Statement to send or receive from multiple channel without blocking on each one
	case *ast.SelectStmt:
		ParseSelectStmt(stmt, &fm)
		return nil

	// Statement to send some data on a channel
	case *ast.SendStmt:
		ParseSendStmt(stmt, &fm)
		return nil

	// Statement for binary or unary expression (channel recv, fucntion call)
	case *ast.ExprStmt:
		ParseExprStmt(stmt, &fm)
		return nil

	// Statement to assign the value of an expression (chanel recv, channel decl, function call)
	case *ast.AssignStmt:
		ParseAssignStmt(stmt, &fm)
		return nil

	// Statement to declare a new variable (channel decl)
	case *ast.DeclStmt:
		ParseDeclStmt(stmt, &fm)
		return nil
	}
	return fm
}

// ----------------------------------------------------------------------------
// Function related parsing method

// This function parses a FuncDecl statement and saves the data extracted in a
// FuncMetadata struct. In case of error during execution (external or non Go function)
// a zero value of abovesaid struct is returned (no error returned).
func ParseFuncDecl(stmt *ast.FuncDecl) FuncMetadata {
	// Retrieve function name and arguments
	funcName := stmt.Name.Name
	funcArgs := stmt.Type.Params.List

	// Initial setup of the metadata record
	metadata := FuncMetadata{
		Name:            funcName,
		ChanMeta:        make(map[string]ChanMetadata),
		InlineArgs:      make(map[string]FuncArg),
		PartialAutomata: NewTransitionGraph(),
	}

	// If the current is an external (non Go) function then is skipped since
	// it isn't useful in order to evaluate the choreography of the automa
	if stmt.Body == nil {
		return FuncMetadata{} // Returns zero value of the struct
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

	// At last all the data extracted is returned
	return metadata
}

// This function parses a GoStmt statement and saves the Transition data extracted
//  in the given FuncMetadata argument. In case of error during execution no error is returned.
func ParseGoStmt(stmt *ast.GoStmt, fm *FuncMetadata) {
	// Determines if GoStmt spawn a Go routine from declared
	// function or an anonymous function is spawned
	funcIdent, isFuncIdent := stmt.Call.Fun.(*ast.Ident)
	_, isFuncAnonymous := stmt.Call.Fun.(*ast.FuncLit)

	// Then extracts the data accoringly
	if isFuncIdent {
		tSpawn := Transition{Kind: Spawn, IdentName: funcIdent.Name}
		fm.PartialAutomata.AddTransition(Current, NewNode, tSpawn)
	} else if isFuncAnonymous {
		anonFuncName := fmt.Sprintf("%s-%s", AnonymousFunc, fm.Name)
		tSpawn := Transition{Kind: Spawn, IdentName: anonFuncName}
		fm.PartialAutomata.AddTransition(Current, NewNode, tSpawn)
		// ? Add parent avaiableChan
		// ? Add parse arguments (different from above)
		// ? Should parse body of funcLiteral (?)
	}
}

// This function parses a CallExpr statement and saves the Transition data extracted
// in the given FuncMetadata argument. In case of error during execution no error is returned.
func ParseCallExpr(expr *ast.CallExpr, fm *FuncMetadata) {
	// Tries to extract the function name (identifier), else throw an exception
	funcIdent, isIdent := expr.Fun.(*ast.Ident)

	if !isIdent {
		// ? Consider struct.method() syntax as well (*ast.SelectorExpr)
		return
	}

	// Creates a valid transaction struct
	tCall := Transition{Kind: Call, IdentName: funcIdent.Name}
	fm.PartialAutomata.AddTransition(Current, NewNode, tCall)
}
