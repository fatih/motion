package main

import (
	"errors"
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

// NewFuncs returns a list of Func's from the given src. The list of Func's are
// sorted according to the order of Go functions in the given src.
func NewFuncs(src []byte) ([]Func, error) {
	fset := token.NewFileSet() // positions are relative to fset
	f, err := parser.ParseFile(fset, "src.go", src, 0)
	if err != nil {
		return nil, err
	}

	var funcs []Func

	// Inspect the AST and find all function declarements and literals
	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncDecl:
			funcs = append(funcs, Func{
				lbrace:  fset.Position(x.Body.Lbrace),
				rbrace:  fset.Position(x.Body.Rbrace),
				funcPos: fset.Position(x.Type.Func),
				node:    x,
			})
		case *ast.FuncLit:
			funcs = append(funcs, Func{
				lbrace:  fset.Position(x.Body.Lbrace),
				rbrace:  fset.Position(x.Body.Rbrace),
				funcPos: fset.Position(x.Type.Func),
				node:    x,
			})
		}
		return true
	})

	return funcs, nil
}

// enclosingFunc returns the enclosing function for the given offset. An error
// is return if no enclosing function is found.
func enclosingFunc(funcs []Func, offset int) (Func, error) {
	for _, fn := range funcs {
		start := fn.lbrace.Offset
		end := fn.rbrace.Offset

		if start <= offset && offset <= end {
			return fn, nil
		}
	}

	return Func{}, errors.New("offset is not enclosing any function")
}
