package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"strings"

	"github.com/its-hmny/Choreia/pkg/utils"
)

// Only for debugging pourposes will be removed later
type debugVisitor int

func (ip debugVisitor) Visit(node ast.Node) ast.Visitor {
	if node == nil {
		return nil
	}
	fmt.Printf("%s %T\n", strings.Repeat("  ", int(ip)), node)
	return ip + 1
}

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
	flags := parser.DeclarationErrors | parser.AllErrors
	// Parse the file identified by the given path and print the tree to the terminal.
	f, err := parser.ParseFile(fset, "test/basic/channel.go", nil, flags)

	if err != nil {
		log.Fatalf("ParseFile error: %s\n", err)
		return
	}

	// Debug Visitor to print to terminal in a more human readable manner the AST
	var debug debugVisitor
	ast.Walk(debug, f)

	fmt.Printf("\n\nGLOBAL SCOPE PARSER DEBUG PRINT \n")

	utils.ExtractPartialAutomata(f)
}
