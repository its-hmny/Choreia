package main

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
)

func ParseAssignment(stmt *ast.AssignStmt) (bool, error) {
	if len(stmt.Lhs) != len(stmt.Rhs) {
		return false, errors.New("should receive same number of l_val and r_val")
	}

	for i, _ := range stmt.Lhs {
		lVal, rVal := stmt.Lhs[i], stmt.Rhs[i]
		tmp, ok := rVal.(*ast.CallExpr)

		// TODO ADD FUNCTIONALITIES
		// Check that the function returned is a "make"
		// Determine if it's a channel (bounded or unbounded)
		// Extrapolate the type of the channel
		if ok {
			fmt.Printf("%s %s \t", lVal, tmp.Args[0])
		}

		// TODO debug pourposes, can remove
		if stmt.Tok == token.DEFINE {
			fmt.Printf("DEFINE %s %T \n", lVal, stmt.Rhs[i])
		} else if stmt.Tok == token.ASSIGN {
			fmt.Printf("ASSIGN %s %T \n", lVal, stmt.Rhs[i])
		}
	}
	return true, nil
}

func inspector(node ast.Node) bool {
	switch typedNode := node.(type) {
	// In case of an assignment we're interested in extrapolating
	// the infos about eventual channel declaration
	case *ast.AssignStmt:
		_, err := ParseAssignment(typedNode)
		if err != nil {
			log.Fatal(err)
		}

	// Error handling case
	case *ast.BadDecl, *ast.BadExpr, *ast.BadStmt:
		log.Fatalf("Error at line %d, column %d", node.Pos(), node.End())
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

func main() {
	// Logger setup
	log.SetPrefix("[choreia-go-parser]: ")
	log.SetFlags(0)

	//if len(os.Args) < 2 {
	//	log.Fatal("A path to an existing go source file is neeeded")
	//}

	// Positions are relative to fset
	fset := token.NewFileSet()
	flags := parser.Trace | parser.DeclarationErrors | parser.AllErrors
	// Parse the file identified by the given path and print the tree to the terminal.
	f, err := parser.ParseFile(fset, "test/basic/channel.go", nil, flags)

	if err != nil {
		fmt.Println(err)
		return
	}

	// With inspect the AST is descended in depth-first order
	ast.Inspect(f, inspector)
}
