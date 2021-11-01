// Copyright Enea Guidi (hmny).

// This package handles the parsing of a given *ast.File which represents
// the content of a Go source file as an Abstract Syntax Tree.

// The only method avaiable from the outside is ParseFile which, as the name suggest, will return
// a FileMetadata struct containing some info needed by the caller for further uses.
package meta

import (
	"fmt"
	"go/ast"
	"log"
)

// ----------------------------------------------------------------------------
// FileMetadata

// A FileMetadata contains the metadata avaiable about a Go source file
//
// A struct containing all the metadata that the algorithm has been able to
// extrapolate from the parsed file.
type FileMetadata struct {
	// The channel declared and avaiable in the global scope
	GlobalChanMeta map[string]ChanMetadata
	// The top-level function declared in the file
	FunctionMeta map[string]FuncMetadata
}

// Adds the given metadata about some channel(s) to the FileMetadata struct
// In case a channel with the same name already exist then the previous association
// is overwritten, this is correct since the channel name is the variable to which
// the channel is assigned and this means that a new assignment was made to that variable
func (fm *FileMetadata) addChannelMeta(newChanMeta ...ChanMetadata) {
	// Adds/updates the associations
	for _, channel := range newChanMeta {
		// Checks the validity of the current item
		if channel.Name != "" && channel.Type != "" {
			fm.GlobalChanMeta[channel.Name] = channel
		}
	}
}

// Adds the given metadata about a function(s) to the FileMetadata struct
// In case of a function with the same name then the previous association
// is overwritten although this should not happen since it's not possible to
// use the same function name with different signature (overloading isn't allowed)
func (fm *FileMetadata) addFunctionMeta(functionMetas ...FuncMetadata) {
	// Adds the metadata association to the map
	for _, function := range functionMetas {
		// Checks the validity of the current item
		if function.Name != "" {
			fm.FunctionMeta[function.Name] = function
		}
	}
}

// In order to satify the ast.Visitor interface FileMetadata implements
// the Visit() method with this function signature. The Visit method takes as
// only argument an ast.Node interface and evaluates all the meaninggul cases,
// when the function steps into that it tries to extract metada from the subtree
func (fm FileMetadata) Visit(node ast.Node) ast.Visitor {
	// Skips empty nodes during descend
	if node == nil {
		return nil
	}

	switch stmt := node.(type) {
	// In this case we're interested in extrapolating info about global channel declaration
	case *ast.GenDecl:
		newChannels := parseGenDecl(stmt)
		fm.addChannelMeta(newChannels...)
		return nil
	// Obvoiusly we want to extrapolate data about the declared function (and their action)
	case *ast.FuncDecl:
		fmt.Printf("\nParsing function '%s'\n", stmt.Name.Name)
		newFunction := parseFuncDecl(stmt)
		fm.addFunctionMeta(newFunction)
		return nil
	// Error handling case
	case *ast.BadDecl, *ast.BadExpr, *ast.BadStmt:
		log.Fatalf("Syntax error from position %d to %d\n", node.Pos(), node.End())
		return nil
	}

	return fm
}

// ----------------------------------------------------------------------------
// File related parsing method

// This function handles the extraction of metadata about the given file, it simply
// receives an *ast.File as input and call ast.Walk on it. Whenever it encounters something
// interesting such as global channel or function declaration it saves the metadata avaiable
func parseAstFile(file *ast.File) FileMetadata {
	// Intializes the FileMetadata struct
	metadata := FileMetadata{
		GlobalChanMeta: map[string]ChanMetadata{},
		FunctionMeta:   map[string]FuncMetadata{},
	}
	// With Walk() descends the AST in depth-first order
	ast.Walk(metadata, file)
	// Returns the collected data
	return metadata
}
