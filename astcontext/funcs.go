// Package astcontext provides context aware utilities to be used within
// editors.
package astcontext

import (
	"errors"
	"fmt"
	"go/ast"
	"sort"
)

// Func represents a declared (*ast.FuncDecl) or an anonymous (*ast.FuncLit) Go
// function
type Func struct {
	FuncPos *Position `json:"func" vim:"func"`     // position of the "func" keyword
	Lbrace  *Position `json:"lbrace" vim:"lbrace"` // position of "{"
	Rbrace  *Position `json:"rbrace" vim:"rbrace"` // position of "}"

	// position of the doc comment, only for *ast.FuncDecl
	Doc *Position `json:"doc,omitempty" vim:"doc,omitempty"`

	node ast.Node // either *ast.FuncDecl or *ast.FuncLit
}

// Funcs represents a list of functions
type Funcs []*Func

// IsDeclaration returns true if the given function is a function declaration
// (*ast.FuncDecl)
func (f *Func) IsDeclaration() bool {
	_, ok := f.node.(*ast.FuncDecl)
	return ok
}

// IsLiteral returns true if the given function is a function literal
// (*ast.FuncLit)
func (f *Func) IsLiteral() bool {
	_, ok := f.node.(*ast.FuncLit)
	return ok
}

func (f *Func) String() string {
	switch x := f.node.(type) {
	case *ast.FuncDecl:
		return fmt.Sprintf("line: %d type: %T name: %s",
			f.FuncPos.Line, f.node, x.Name.Name)
	default:
		return fmt.Sprintf("line: %d type: %T",
			f.FuncPos.Line, f.node)
	}
}

// Funcs returns a list of Func's from the parsed source.  Func's are sorted
// according to the order of Go functions in the given source.
func (p *Parser) Funcs() Funcs {
	var funcs []*Func

	// Inspect the AST and find all function declarements and literals
	ast.Inspect(p.file, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncDecl:
			fn := &Func{
				Lbrace:  ToPosition(p.fset.Position(x.Body.Lbrace)),
				Rbrace:  ToPosition(p.fset.Position(x.Body.Rbrace)),
				FuncPos: ToPosition(p.fset.Position(x.Type.Func)),
				node:    x,
			}

			if x.Doc != nil {
				fn.Doc = ToPosition(p.fset.Position(x.Doc.Pos()))
			}

			funcs = append(funcs, fn)
		case *ast.FuncLit:
			funcs = append(funcs, &Func{
				Lbrace:  ToPosition(p.fset.Position(x.Body.Lbrace)),
				Rbrace:  ToPosition(p.fset.Position(x.Body.Rbrace)),
				FuncPos: ToPosition(p.fset.Position(x.Type.Func)),
				node:    x,
			})
		}
		return true
	})

	return funcs
}

// EnclosingFunc returns the enclosing *Func for the given offset
func (f Funcs) EnclosingFunc(offset int) (*Func, error) {
	var encFunc *Func

	// TODO(arslan) this is iterating over all functions. Benchmark it and see
	// if it's worth it to change it with a more effiecent search function. For
	// now this is enough for us.
	for _, fn := range f {
		// standard function declaration without any docs. Start from the func
		// keyword
		start := fn.FuncPos.Offset

		// has a doc, also include it
		if fn.Doc != nil && fn.Doc.IsValid() {
			start = fn.Doc.Offset
		}

		// one liner, start from the beginning to make it easier
		if fn.FuncPos.Line == fn.Rbrace.Line {
			start = fn.FuncPos.Offset - fn.FuncPos.Column
		}

		end := fn.Rbrace.Offset

		if start <= offset && offset <= end {
			encFunc = fn
		}
	}

	if encFunc == nil {
		return nil, errors.New("no enclosing functions found")
	}

	return encFunc, nil
}

// NextFunc returns the nearest next Func for the given offset.
func (f Funcs) NextFunc(offset int) (*Func, error) {
	// find nearest next function
	nextIndex := sort.Search(len(f), func(i int) bool {
		return f[i].FuncPos.Offset > offset
	})

	if nextIndex == len(f) {
		return nil, errors.New("no functions found")
	}
	return f[nextIndex], nil
}

// PrevFunc returns the nearest previous *Func for the given offset.
func (f Funcs) PrevFunc(offset int) (*Func, error) {
	// start from the reverse to get the prev function
	f.Reserve()

	prevIndex := sort.Search(len(f), func(i int) bool {
		return f[i].FuncPos.Offset < offset
	})

	if prevIndex == len(f) {
		return nil, errors.New("no functions found")
	}

	return f[prevIndex], nil
}

func (f Funcs) Len() int           { return len(f) }
func (f Funcs) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }
func (f Funcs) Less(i, j int) bool { return f[i].FuncPos.Offset < f[j].FuncPos.Offset }

// Reserve reserves the Function data
func (f Funcs) Reserve() {
	for start, end := 0, f.Len()-1; start < end; {
		f.Swap(start, end)
		start++
		end--
	}
}

// Declarations returns a copy of funcs with only Function declarations
func (f Funcs) Declarations() Funcs {
	// NOTE(arslan): we can prepopulate these in the future, but again we need
	// to benchmark first
	var decls []*Func
	for _, fn := range f {
		if fn.IsDeclaration() {
			decls = append(decls, fn)
		}
	}
	return decls
}
