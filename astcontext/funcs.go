// Package astcontext provides context aware utilities to be used within
// editors.
package astcontext

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
	FuncPos token.Position `json:"funcPos"` // position of the "func" keyword
	Lbrace  token.Position `json:"lbrace"`  // position of "{"
	Rbrace  token.Position `json:"rbrace"`  // position of "}"

	node ast.Node // either *ast.FuncDecl or *ast.FuncLit
}

func (f *Func) String() string {
	return fmt.Sprintf("%s:%d {%d-%d} (%T)", f.FuncPos.Filename,
		f.FuncPos.Offset, f.Lbrace.Offset, f.Rbrace.Offset, f.node)
}

// EnclosingFunc returns the enclosing Func for the given offset from the src.
// Src needs to be a valid Go source file content.
func EnclosingFunc(src []byte, offset int) (*Func, error) {
	funcs, err := ParseFuncs(src)
	if err != nil {
		return nil, err
	}

	return enclosingFunc(funcs, offset)
}

// EnclosingFuncFile returns the enclosing *Func for the given offset from the
// filename. File needs to be a valid Go source file.
func EnclosingFuncFile(filename string, offset int) (*Func, error) {
	funcs, err := ParseFuncsFile(filename)
	if err != nil {
		return nil, err
	}

	return enclosingFunc(funcs, offset)
}

// ParseFuncsFile returns a list of Func's from the given filename. The list of
// Func's are sorted according to the order of Go functions in the given
// filename.
func ParseFuncsFile(filename string) ([]*Func, error) {
	fset := token.NewFileSet() // positions are relative to fset
	f, err := parser.ParseFile(fset, filename, nil, 0)
	if err != nil {
		return nil, err
	}
	return parseFuncs(fset, f), nil
}

// ParseFuncs returns a list of Func's from the given src. The list of Func's
// are sorted according to the order of Go functions in the given src.
func ParseFuncs(src []byte) ([]*Func, error) {
	fset := token.NewFileSet() // positions are relative to fset
	f, err := parser.ParseFile(fset, "src.go", src, 0)
	if err != nil {
		return nil, err
	}

	return parseFuncs(fset, f), nil
}

func parseFuncs(fset *token.FileSet, f ast.Node) []*Func {
	var funcs []*Func

	// Inspect the AST and find all function declarements and literals
	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncDecl:
			funcs = append(funcs, &Func{
				Lbrace:  fset.Position(x.Body.Lbrace),
				Rbrace:  fset.Position(x.Body.Rbrace),
				FuncPos: fset.Position(x.Type.Func),
				node:    x,
			})
		case *ast.FuncLit:
			funcs = append(funcs, &Func{
				Lbrace:  fset.Position(x.Body.Lbrace),
				Rbrace:  fset.Position(x.Body.Rbrace),
				FuncPos: fset.Position(x.Type.Func),
				node:    x,
			})
		}
		return true
	})

	return funcs
}

func enclosingFunc(funcs []*Func, offset int) (*Func, error) {
	var encFunc *Func
	for _, fn := range funcs {
		start := fn.Lbrace.Offset
		end := fn.Rbrace.Offset
		if start <= offset && offset <= end {
			encFunc = fn
		}
	}

	if encFunc == nil {
		return &Func{}, errors.New("no enclosing functions found")
	}

	return encFunc, nil
}
