package utils

import (
	"errors"
	"fmt"
	"go/ast"
)

// Struct containing the needed metadata about a Channel
// TODO ADD DOCS
type ChannelMetadata struct {
	name   string
	typing string
	async  bool
}

// TODO Maybe remove if not used
func (meta *ChannelMetadata) String() string {
	if meta != nil {
		return fmt.Sprintf("ChannelMetadata { name: %s, type: %s, async: %t }", meta.name, meta.typing, meta.async)
	}
	return "<nil>"
}

// TODO ADD DOCS
func ExtrapolateChanMetadata(stmt *ast.AssignStmt) ([]ChannelMetadata, error) {
	// A Slice containing all the metadata retrieved about the channel declared
	metadata := make([]ChannelMetadata, 0)

	// Check that the number of rvalues (variable assigned) are the same of
	// lvalues (values assignments) in the statement (TODO TEST)
	if len(stmt.Lhs) != len(stmt.Rhs) {
		return nil, errors.New("ExtrapolateChanMetadata: should receive same number of l_val and r_val")
	}

	// Now iterates over the assignment statements
	for i := range stmt.Lhs {
		lVal, rVal := stmt.Lhs[i], stmt.Rhs[i]

		// Since is posible to init a channel only with a "make" function call
		// we just want to consider Rhs expression that are function call
		funcCallExpr, isCallExpr := rVal.(*ast.CallExpr)
		// If the Rhs expression isn't a function call we skip the current iteration
		if !isCallExpr {
			continue
		}

		// Then we're interested in filtering only the "make" call used to create
		// a channel (since it can be used to create array and slices as well)
		funcName, isIdentifier := funcCallExpr.Fun.(*ast.Ident)
		channelTypeExpr, isChannelType := funcCallExpr.Args[0].(*ast.ChanType)

		// If every condition matches we're in front of a new channel declaration
		if isIdentifier && isChannelType && funcName.Name == "make" {
			// Extrapolates all the metadata needed
			channelName := lVal.(*ast.Ident).Name
			channelType := channelTypeExpr.Value.(*ast.Ident).Name
			isChannelBuffered := len(funcCallExpr.Args) > 1

			// Creates a new metadata struct that is then added to the previosly declared slice
			newChannelMetadata := ChannelMetadata{channelType, channelName, isChannelBuffered}
			metadata = append(metadata, newChannelMetadata)
		}
	}

	return metadata, nil
}
