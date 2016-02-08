// Package astcontext provides context aware utilities to be used within
// editors.
package astcontext

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
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

// EnclosingFunc returns the enclosing Func for the given offset from the src.
// Src needs to be a valid Go source file content.
func EnclosingFunc(src []byte, offset int) Func {
	funcs, err := ParseFuncs(src)
	if err != nil {
		return Func{}
	}

	encFunc := Func{}
	for _, fn := range funcs {
		start := fn.lbrace.Offset
		end := fn.rbrace.Offset
		if start <= offset && offset <= end {
			encFunc = fn
		}
	}

	return encFunc
}

// EnclosingFunc returns the enclosing Func for the given offset from the
// filename. File needs to be a valid Go source file.
func EnclosingFuncFile(filename string, offset int) Func {
	src, err := ioutil.ReadFile(filename)
	if err != nil {
		return Func{}
	}
	return EnclosingFunc(src, offset)
}

// ParseFuncs returns a list of Func's from the given src. The list of Func's
// are sorted according to the order of Go functions in the given src.
func ParseFuncs(src []byte) ([]Func, error) {
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
