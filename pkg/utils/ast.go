package utils

import (
	"go/ast"
	"log"
)

// TODO Check if the cursor can be moved since we pass on already visited nodes
func AstInspector(node ast.Node) bool {
	switch statement := node.(type) {
	// In case of an assignment we're interested in extrapolating
	// the infos about eventual channel declaration/assignment
	case *ast.AssignStmt:
		// TODO Use returned ChannelMetadata[]
		_, err := ExtrapolateChanMetadata(statement)
		if err != nil {
			log.Fatal(err)
			return false
		}

	// In this case we're spawning a new goroutine
	case *ast.GoStmt:
		// TODO Use returned GoRoutineMetadata[]
		_, err := GetGoRoutineMetadata(statement)
		if err != nil {
			log.Fatal(err)
			return false
		}

	// Error handling case
	case *ast.BadDecl, *ast.BadExpr, *ast.BadStmt:
		log.Fatalf("Syntax error at line %d, column %d\n", node.Pos(), node.End())
		//case *ast.BasicLit:
		//	fmt.Println()
		//case *ast.BinaryExpr:
		//	fmt.Println()
		//case *ast.BlockStmt:
		//	fmt.Println()
		//case *ast.BranchStmt:
		//	fmt.Println()
		//case *ast.CallExpr:
		//	fmt.Println()
		//case *ast.CaseClause:
		//	fmt.Println()
		// TODO ADD MISSING ONES
	}

	return true
}
