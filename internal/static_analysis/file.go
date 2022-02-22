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
	"go/ast"
	"log"
)

// ----------------------------------------------------------------------------
// FileMetadata

// A FileMetadata contains the metadata available about a Go source file
//
// A struct containing all the metadata that the Visitor has been able to
// gather from the parsed file. The data are structured hierarchically:
// Module -> File -> Function -> Channels
type FileMetadata struct {
	GlobalChanMeta map[string]ChanMetadata // The channel declared in the global scope
	FunctionMeta   map[string]FuncMetadata // The top-level function declared in the file
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

// In order to satisfy the ast.Visitor interface FileMetadata implements
// the Visit() method with this function signature. The Visit method takes as
// only argument an ast.Node interface and evaluates all the meaningful cases,
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
	// Obviously we want to extrapolate data about the declared function (and their action)
	case *ast.FuncDecl:
		parseFuncDecl(stmt, fm)
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
// interesting such as global channel or function declaration it saves the metadata available
func parseAstFile(file *ast.File) FileMetadata {
	// Initializes the FileMetadata struct
	metadata := FileMetadata{
		GlobalChanMeta: map[string]ChanMetadata{},
		FunctionMeta:   map[string]FuncMetadata{},
	}
	// With Walk() descends the AST in depth-first order
	ast.Walk(metadata, file)
	// Returns the collected data
	return metadata
}
