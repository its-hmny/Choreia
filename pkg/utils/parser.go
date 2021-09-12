package utils

import (
	"errors"
	"fmt"
	"go/ast"
)

// TODO ADD FUNCTIONALITIES
// Determine if it's a channel (bounded or unbounded)
// Extrapolate the type of the channel
func ExtrapolateChanDecl(stmt *ast.AssignStmt) (uint, error) {
	// Simple counter of channel declaration found while parsing this statement
	var channelDeclFound uint = 0
	// Check that the number of rvalues (variable assigned) are the same of
	// lvalues (values assignments) in the statement (TODO TEST)
	if len(stmt.Lhs) != len(stmt.Rhs) {
		return channelDeclFound, errors.New("should receive same number of l_val and r_val")
	}

	// Now iterates over the assignment statements
	for i := range stmt.Lhs {
		lVal, rVal := stmt.Lhs[i], stmt.Rhs[i]

		// Since is posible to only init a channel from "make" function call the
		// function will return if the rvalue isn't a function call expression
		funcCallExpr, isCallExpr := rVal.(*ast.CallExpr)
		if !isCallExpr {
			continue
		}

		// Then we're interested in filtering only the "make" calls used to create
		// a channel (since it can be used to create array and slices as well)
		funcName, isIdentifier := funcCallExpr.Fun.(*ast.Ident)
		channelType, isChannelType := funcCallExpr.Args[0].(*ast.ChanType)
		if isIdentifier && isChannelType && funcName.Name == "make" {
			channelDeclFound++
			// Now we need to extrapolate the avaiable data (TODO)
			fmt.Printf("Channel init found: %s %s %s \n", lVal, funcName, channelType.Value)
		}
	}

	return channelDeclFound, nil
}
