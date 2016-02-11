// Package astcontext provides context aware utilities to be used within
// editors.
package astcontext

import (
	"errors"
	"fmt"
	"go/ast"
	"go/token"
)

// Func represents a declared (*ast.FuncDecl) or an anonymous (*ast.FuncLit) Go
// function
type Func struct {
	FuncPos token.Position `json:"funcPos"` // position of the "func" keyword
	Lbrace  token.Position `json:"lbrace"`  // position of "{"
	Rbrace  token.Position `json:"rbrace"`  // position of "}"
	Doc     token.Position `json:"docPos"`  // position of the doc comment, only for *ast.FuncDecl

	node ast.Node // either *ast.FuncDecl or *ast.FuncLit
}

func (f *Func) String() string {
	if f.Doc.IsValid() {
		return fmt.Sprintf("{'func': {'line': '%d', 'col':'%d'}, 'doc': {'line': '%d', 'col':'%d'}, 'lbrace': {'line': '%d', 'col':'%d'}, 'rbrace': {'line': '%d', 'col':'%d'}}",
			f.FuncPos.Line, f.FuncPos.Column,
			f.Doc.Line, f.Doc.Column,
			f.Lbrace.Line, f.Lbrace.Column,
			f.Rbrace.Line, f.Rbrace.Column,
		)
	}

	return fmt.Sprintf("{'func': {'line': '%d', 'col':'%d'}, 'lbrace': {'line': '%d', 'col':'%d'}, 'rbrace': {'line': '%d', 'col':'%d'}}",
		f.FuncPos.Line, f.FuncPos.Column, f.Lbrace.Line, f.Lbrace.Column, f.Rbrace.Line, f.Rbrace.Column)
}

// EnclosingFunc returns the enclosing *Func for the given offset
func (p *Parser) EnclosingFunc(offset int) (*Func, error) {
	funcs, err := p.Funcs()
	if err != nil {
		return nil, err
	}

	return enclosingFunc(funcs, offset)
}

// Funcs returns a list of Func's from the parsed source.  Func's are sorted
// according to the order of Go functions in the given source.
func (p *Parser) Funcs() ([]*Func, error) {
	if p.err != nil {
		return nil, p.err
	}

	return parseFuncs(p.fset, p.file), nil
}

func parseFuncs(fset *token.FileSet, f ast.Node) []*Func {
	var funcs []*Func

	// Inspect the AST and find all function declarements and literals
	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncDecl:
			fn := &Func{
				Lbrace:  fset.Position(x.Body.Lbrace),
				Rbrace:  fset.Position(x.Body.Rbrace),
				FuncPos: fset.Position(x.Type.Func),
				node:    x,
			}

			if x.Doc != nil {
				fn.Doc = fset.Position(x.Doc.Pos())
			}

			funcs = append(funcs, fn)
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
		start := fn.FuncPos.Offset
		if fn.Doc.IsValid() {
			start = fn.Doc.Offset
		}

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
