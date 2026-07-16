package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

const gormStoreDirectory = "apps/api/internal/adapters/gormstore"

func main() {
	paths := commandPaths(os.Args[1:])
	fileSet := token.NewFileSet()
	violations := 0
	for _, path := range paths {
		if !isGORMStoreGoFile(path) {
			continue
		}
		file, err := parser.ParseFile(fileSet, path, nil, 0)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: parse Go source: %v\n", path, err)
			os.Exit(2)
		}
		ast.Inspect(file, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok || len(call.Args) == 0 {
				return true
			}
			selector, ok := call.Fun.(*ast.SelectorExpr)
			if !ok || !checkedMethod(selector.Sel.Name) {
				return true
			}
			literal, ok := call.Args[0].(*ast.BasicLit)
			if !ok || literal.Kind != token.STRING {
				return true
			}
			position := fileSet.Position(literal.Pos())
			fmt.Fprintf(os.Stderr, "%s: string query fragment passed to %s; use structured GORM clauses or typed repository helpers\n", position, selector.Sel.Name)
			violations++
			return true
		})
	}
	if violations > 0 {
		os.Exit(1)
	}
}

func commandPaths(args []string) []string {
	for index, arg := range args {
		if arg == "--" {
			return args[index+1:]
		}
	}
	return args
}

func isGORMStoreGoFile(path string) bool {
	cleaned := filepath.ToSlash(filepath.Clean(path))
	return filepath.ToSlash(filepath.Dir(cleaned)) == gormStoreDirectory && strings.HasSuffix(cleaned, ".go")
}

func checkedMethod(name string) bool {
	switch name {
	case "Where", "Order", "Joins":
		return true
	default:
		return false
	}
}
