package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"log"

	"github.com/its-hmny/Choreia/pkg/utils"
)

func main() {
	// Logger setup
	log.SetPrefix("[Choreia]: ")
	log.SetFlags(0)

	// Command line argument checking
	//if len(os.Args) < 2 {
	//	log.Fatal("A path to an existing go source file is neeeded")
	//}

	// Positions are relative to fset
	fset := token.NewFileSet()
	// Parser mode flags, we want all every error possible and a trace printed on the stdout
	flags := parser.DeclarationErrors | parser.AllErrors | parser.Trace
	// Parse the file identified by the given path and print the tree to the terminal.
	f, err := parser.ParseFile(fset, "test/basic/channel.go", nil, flags)

	if err != nil {
		log.Fatalf("ParseFile error: %s\n", err)
		return
	}

	// With inspect the AST is descended in depth-first order
	ast.Inspect(f, utils.AstInspector)
}
