// Copyright Enea Guidi (hmny).

// This package handles the parsing of a given *ast.File which represents
// the content of a Go source file as an Abstract Syntax Tree.

// The only method available from the outside is ParseFuncDecl, ParseGoStmt and ParseCallExpr which
// will add to the given FileMetadata argument the data collected from the parsing phases
package meta

import (
	"fmt"
	"go/ast"

	"github.com/its-hmny/Choreia/internal/types/fsa"
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

type ArgType int

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

// This function parses a FuncDecl statement and saves the data extracted in a
// FuncMetadata struct. In case of error during execution (external or non Go function)
// a zero value of abovesaid struct is returned (no error returned).
func parseFuncDecl(stmt *ast.FuncDecl) FuncMetadata {
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

	// If the current is an external (non Go) function then is skipped since
	// it isn't useful in order to evaluate the choreography of the automon
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
func parseGoStmt(stmt *ast.GoStmt, fm *FuncMetadata) {
	// Determines if GoStmt spawn a Go routine from declared
	// function or an anonymous function is spawned
	funcIdent, isFuncIdent := stmt.Call.Fun.(*ast.Ident)
	_, isFuncAnonymous := stmt.Call.Fun.(*ast.FuncLit)

	// Then extracts the data accordingly
	if isFuncIdent {
		tSpawn := fsa.Transition{Move: fsa.Spawn, Label: funcIdent.Name}

		// Parses the GoStmt arguments looking for channels and saves the "actual" argument to list
		// in the Transition. Later this channels will be inlined during the generation of the automaton
		// ! Remove starting duplicate at line 240
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
		anonFuncName := fmt.Sprintf("%s-%s", anonymousFunc, fm.Name)
		tSpawn := fsa.Transition{Move: fsa.Spawn, Label: anonFuncName}
		fm.ScopeAutomata.AddTransition(fsa.Current, fsa.NewState, tSpawn)
		// ? Add parent availableChan
		// ? Add parse arguments (different from above)
		// ? Should parse body of funcLiteral (?)
	}
}

// This function parses a CallExpr statement and saves the Transition data extracted
// in the given FuncMetadata argument. In case of error during execution no error is returned.
func parseCallExpr(expr *ast.CallExpr, fm *FuncMetadata) {
	// Tries to extract the function name (identifier), else throw an exception
	funcIdent, isIdent := expr.Fun.(*ast.Ident)

	if !isIdent {
		// ? Consider struct.method() syntax as well (*ast.SelectorExpr)
		return
	}

	// Creates a valid transaction struct
	tCall := fsa.Transition{Move: fsa.Call, Label: funcIdent.Name}

	// Parses the CallExpr arguments looking for channels and saves the "actual" argument to list
	// in the Transition. Later this channels will be inlined during the generation of the automaton
	// ! Remove starting duplicate at line 199
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
