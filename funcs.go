package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
)

// Func represents a declared (*ast.FuncDecl) or an anonymous (*ast.FuncLit) Go
// function
type Func struct {
	funcPos token.Position // position of the "func" keyword
	lbrace  token.Position // position of "{"
	rbrace  token.Position // position of "}"

	node ast.Node // either *ast.FuncDecl or *ast.FuncLit
}

func (f Func) String() string {
	return fmt.Sprintf("%s (%T)", f.funcPos, f.node)
}

// NewFuncs returns a list of Func's from the given src
func NewFuncs(src []byte) ([]Func, error) {
	// Create the AST by parsing src.
	fset := token.NewFileSet() // positions are relative to fset
	f, err := parser.ParseFile(fset, "src.go", src, 0)
	if err != nil {
		return nil, err
	}

	var funcs []Func

	addFunc := func(node ast.Node, fnType *ast.FuncType) {
		funcs = append(funcs, Func{
			lbrace:  fset.Position(node.Pos()),
			rbrace:  fset.Position(node.End()),
			funcPos: fset.Position(fnType.Func),
			node:    node,
		})
	}

	// Inspect the AST and print all identifiers and literals.
	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncDecl:
			addFunc(x, x.Type)
		case *ast.FuncLit:
			addFunc(x, x.Type)
		}

		return true
	})
	return funcs, nil
}
