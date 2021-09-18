package utils

import (
	"fmt"
	"go/ast"
	"log"
)

// A struct containing all the metadata that the algorithm
// has been able to extrapolate from the parsed file
type fileMetadata struct {
	globalChan map[string]ChannelMetadata
	funcDecl   map[string]FunctionMetadata
}

// Adds the given metadata about some channels to the fileMetadata struct
// In case a channel with the same name already exist then the previous association
// is overwritten, this is correct since the channel name is the variable to which
// the channel is assigned and this means that a new assignment was made to that variable
func (fm *fileMetadata) AddChannels(newChannels []ChannelMetadata) {
	if len(newChannels) <= 0 {
		return
	}

	// Adds/updates the associations
	for _, channel := range newChannels {
		fm.globalChan[channel.name] = channel
	}
}

// In order for fileMetadata to be used in the ast.Walk() method, it must implement
// the Visitor interface and subsequently have a Visit() method with this signature
func (fm fileMetadata) Visit(node ast.Node) ast.Visitor {
	// Skips empty nodes during descend
	if node == nil {
		return nil
	}

	switch stmt := node.(type) {
	// In this case we're interested in extrapolating info about channels declaration
	case *ast.AssignStmt, *ast.GenDecl:
		newChannels := GetChanMetadata(stmt)
		fm.AddChannels(newChannels)
		return nil
	// Error handling case
	case *ast.BadDecl, *ast.BadExpr, *ast.BadStmt:
		log.Fatalf("Syntax error at line %d, column %d\n", node.Pos(), node.End())
		return nil
	}

	return fm
}

// TODO COMMENT
func ExtractPartialAutomata(file *ast.File) {
	// Intializes the file metadata struct in which all the data
	// avaiable and useful will be stored
	metadata := fileMetadata{
		map[string]ChannelMetadata{},
		map[string]FunctionMetadata{},
	}

	// With Walk() descends the AST in depth-first order
	ast.Walk(metadata, file)

	fmt.Printf("File metadata: %+v \n", metadata)
}
