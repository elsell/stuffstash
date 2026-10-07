package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path"
	"regexp"
	"strconv"
)

var sqlStatement = regexp.MustCompile(`(?is)\b(SELECT|INSERT\s+INTO|UPDATE\s+.+?\s+SET|DELETE\s+FROM|CREATE\s+TABLE|ALTER\s+TABLE|DROP\s+TABLE)\s+`)

func main() {
	files := os.Args[1:]
	if len(files) > 0 && files[0] == "--" {
		files = files[1:]
	}
	positions := token.NewFileSet()
	failed := false
	for _, name := range files {
		file, err := parser.ParseFile(positions, name, nil, 0)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		diagnostic := diagnosticLiterals(file)
		ast.Inspect(file, func(n ast.Node) bool {
			value, ok := n.(*ast.BasicLit)
			if !ok || value.Kind != token.STRING || diagnostic[value] {
				return true
			}
			text, err := strconv.Unquote(value.Value)
			if err == nil && sqlStatement.MatchString(text) {
				fmt.Fprintf(os.Stderr, "%s: raw SQL string; use structured GORM clauses behind repository adapters\n", positions.Position(value.Pos()))
				failed = true
			}
			return true
		})
	}
	if failed {
		os.Exit(1)
	}
}

func diagnosticLiterals(file *ast.File) map[*ast.BasicLit]bool {
	imports := map[string]string{}
	for _, item := range file.Imports {
		name, err := strconv.Unquote(item.Path.Value)
		if err != nil {
			continue
		}
		alias := path.Base(name)
		if item.Name != nil {
			alias = item.Name.Name
		}
		imports[alias] = name
	}
	result := map[*ast.BasicLit]bool{}
	ast.Inspect(file, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		selector, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		qualifier, ok := selector.X.(*ast.Ident)
		if !ok || qualifier.Obj != nil {
			return true
		}
		index := -1
		switch imports[qualifier.Name] + "." + selector.Sel.Name {
		case "errors.New", "fmt.Errorf":
			index = 0
		case "github.com/stuffstash/stuff-stash/cli/internal/ports.Failure":
			index = 1
		}
		if index >= 0 && len(call.Args) > index {
			if literal, ok := call.Args[index].(*ast.BasicLit); ok && literal.Kind == token.STRING {
				result[literal] = true
			}
		}
		return true
	})
	return result
}
