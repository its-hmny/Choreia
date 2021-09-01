package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
)

func inspector(node ast.Node) bool {
	switch nodetype := node.(type) {
	case *ast.ArrayType:
		fmt.Println()
	case *ast.AssignStmt:
		fmt.Println()
	case *ast.BasicLit:
		fmt.Println()
	case *ast.BinaryExpr:
		fmt.Println()
	case *ast.BlockStmt:
		fmt.Println()
	case *ast.BranchStmt:
		fmt.Println()
	case *ast.CallExpr:
		fmt.Println()
	case *ast.CaseClause:
		fmt.Println()
	// TODO ADD MISSING ONES
	default:
		fmt.Println(nodetype)
	}
	return true
}

func main() {
	// Logger setup
	log.SetPrefix("[choreia-go-parser]: ")
	log.SetFlags(0)

	if len(os.Args) < 2 {
		log.Fatal("A path to an existing go source file is neeeded")
	}

	// Positions are relative to fset
	fset := token.NewFileSet()
	flags := parser.Trace | parser.DeclarationErrors | parser.AllErrors
	// Parse the file identified by the given path and print the tree to the terminal.
	f, err := parser.ParseFile(fset, os.Args[1], nil, flags)

	if err != nil {
		fmt.Println(err)
		return
	}

	// With inspect the AST is descended in depth-first order
	ast.Inspect(f, inspector)
}
