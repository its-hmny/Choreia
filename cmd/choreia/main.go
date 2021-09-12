package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"

	"github.com/its-hmny/Choreia/pkg/utils"
)

func inspector(node ast.Node) bool {
	switch typedNode := node.(type) {
	// In case of an assignment we're interested in extrapolating
	// the infos about eventual channel declaration
	case *ast.AssignStmt:
		_, err := utils.ExtrapolateChanDecl(typedNode)
		if err != nil {
			log.Fatalf("ParseAssignment error: %s\n", err)
		}

	// Error handling case
	case *ast.BadDecl, *ast.BadExpr, *ast.BadStmt:
		log.Fatalf("Error at line %d, column %d\n", node.Pos(), node.End())
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
	flags := parser.DeclarationErrors | parser.AllErrors | parser.Trace
	// Parse the file identified by the given path and print the tree to the terminal.
	f, err := parser.ParseFile(fset, "test/basic/channel.go", nil, flags)

	if err != nil {
		fmt.Println(err)
		return
	}

	// With inspect the AST is descended in depth-first order
	ast.Inspect(f, inspector)
}
