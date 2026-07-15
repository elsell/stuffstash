package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
)

var forbiddenMethods = map[string]struct{}{
	"CreateAsset":  {},
	"ArchiveAsset": {},
	"RestoreAsset": {},
}

func main() {
	files := token.NewFileSet()
	failed := false
	for _, path := range os.Args[1:] {
		if path == "--" {
			continue
		}
		parsed, err := parser.ParseFile(files, path, nil, 0)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: parse Go source: %v\n", path, err)
			os.Exit(2)
		}
		for _, declaration := range parsed.Decls {
			function, ok := declaration.(*ast.FuncDecl)
			if !ok || function.Recv == nil || len(function.Recv.List) != 1 {
				continue
			}
			if _, forbidden := forbiddenMethods[function.Name.Name]; !forbidden {
				continue
			}
			if receiverType(function.Recv.List[0].Type) != "App" && receiverType(function.Recv.List[0].Type) != "Service" {
				continue
			}
			position := files.Position(function.Pos())
			fmt.Fprintf(os.Stderr, "%s:%d: asset-only %s.%s compatibility facade is not allowed\n", path, position.Line, receiverType(function.Recv.List[0].Type), function.Name.Name)
			failed = true
		}
	}
	if failed {
		fmt.Fprintln(os.Stderr, "asset application commands must expose operation-aware results")
		os.Exit(1)
	}
}

func receiverType(expression ast.Expr) string {
	if pointer, ok := expression.(*ast.StarExpr); ok {
		expression = pointer.X
	}
	identifier, _ := expression.(*ast.Ident)
	if identifier == nil {
		return ""
	}
	return identifier.Name
}
