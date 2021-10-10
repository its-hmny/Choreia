// Copyright Enea Guidi (hmny).

// This package handles the parsing of a given *ast.File which represents
// the content of a Go source file as an Abstract Syntax Tree,

// The only method avaiable from the outside is ExtractFileMetadata which as the name
// suggest will return a FileMetadata struct containing some info needed by the caller
// for further uses.
package parser

import (
	"go/ast"
	"log"
)

// A struct containing all the metadata that the algorithm
// has been able to extrapolate from the parsed file
type FileMetadata struct {
	// The channel declared and avaiable in the global scope
	GlobalChan map[string]ChannelMetadata
	// The top-level function declared in the file
	FuncDecl map[string]FunctionMetadata
}

// Adds the given metadata about some channels to the fileMetadata struct
// In case a channel with the same name already exist then the previous association
// is overwritten, this is correct since the channel name is the variable to which
// the channel is assigned and this means that a new assignment was made to that variable
func (fm *FileMetadata) addChannelMeta(channelMetas ...ChannelMetadata) {
	// Adds/updates the associations
	for _, channel := range channelMetas {
		// Checks the validity of the current item
		if channel.Name != "" && channel.Typing != "" {
			fm.GlobalChan[channel.Name] = channel
		}
	}
}

// Adds the given metadata about a function to the fileMetadata struct
// In case of a function with the same name then the previous association
// is overwritten although this should not happen since it's not possible to
// use the same function name more than one times (except for overloading that is ignored)
func (fm *FileMetadata) addFunctionMeta(functionMetas ...FunctionMetadata) {
	// Adds the metadata association to the map
	for _, function := range functionMetas {
		// Checks the validity of the current item
		if function.Name != "" {
			fm.FuncDecl[function.Name] = function
		}
	}
}

// In order for fileMetadata to be used in the ast.Walk() method, it must implement
// the Visitor interface and subsequently have a Visit() method with this signature
func (fm FileMetadata) Visit(node ast.Node) ast.Visitor {
	// Skips empty nodes during descend
	if node == nil {
		return nil
	}

	switch stmt := node.(type) {
	// In this case we're interested in extrapolating info about global channel declaration
	case *ast.GenDecl:
		newChannels := ExtractChanMetadata(stmt)
		fm.addChannelMeta(newChannels...)
		return nil
	// Obvoiusly we want to extrapolate data about the declared function (and their action)
	case *ast.FuncDecl:
		newFunction := NewFunctionMetadata(stmt)
		fm.addFunctionMeta(newFunction)
		return nil
	// Error handling case
	case *ast.BadDecl, *ast.BadExpr, *ast.BadStmt:
		log.Fatalf("Syntax error from position %d to %d\n", node.Pos(), node.End())
		return nil
	}

	return fm
}

// This function handles the extraction of metadata about the given file, it simply
// receives an *ast.File as input and call ast.Walk on it. Whenever it encounters something
// interesting such as global channel or function declaration it saves the metadata avaiable
func ExtractFileMetadata(file *ast.File) FileMetadata {
	// Intializes the file metadata struct in which all the data
	// avaiable and useful will be stored
	metadata := FileMetadata{
		map[string]ChannelMetadata{},
		map[string]FunctionMetadata{},
	}

	// With Walk() descends the AST in depth-first order
	ast.Walk(metadata, file)

	return metadata
}
