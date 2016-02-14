// Package astcontext provides context aware utilities to be used within
// editors.
package astcontext

import (
	"errors"
	"fmt"
	"go/ast"
	"go/token"
	"sort"
)

// Position describes a function position
type Position struct {
	Offset int `json:"offset" vim:"offset"` // offset, starting at 0
	Line   int `json:"line" vim:"line"`     // line number, starting at 1
	Column int `json:"col" vim:"col"`       // column number, starting at 1 (byte count)
}

// Func represents a declared (*ast.FuncDecl) or an anonymous (*ast.FuncLit) Go
// function
type Func struct {
	FuncPos *Position `json:"func" vim:"func"`     // position of the "func" keyword
	Lbrace  *Position `json:"lbrace" vim:"lbrace"` // position of "{"
	Rbrace  *Position `json:"rbrace" vim:"rbrace"` // position of "}"

	// position of the doc comment, only for *ast.FuncDecl
	Doc *Position `json:"doc,omitempty" vim:"doc,omitempty"`

	node    ast.Node // either *ast.FuncDecl or *ast.FuncLit
	literal bool     // literal function or not
}

// ToPosition returns a Position from the given token.Position
func ToPosition(pos token.Position) *Position {
	return &Position{
		Offset: pos.Offset,
		Line:   pos.Line,
		Column: pos.Column,
	}
}

func (pos Position) IsValid() bool { return pos.Line > 0 }

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

// EnclosingFunc returns the enclosing *Func for the given offset
func (p *Parser) EnclosingFunc(offset int) (*Func, error) {
	funcs := p.Funcs()
	return enclosingFunc(funcs, offset)
}

// NextFunc returns the nearest next *Func for the given offset.
func (p *Parser) NextFunc(offset int) (*Func, error) {
	funcs := p.Funcs()
	return nextFunc(funcs, offset)
}

// PrevFunc returns the nearest previous *Func for the given offset.
func (p *Parser) PrevFunc(offset int) (*Func, error) {
	funcs := p.Funcs()
	return prevFunc(funcs, offset)
}

// Funcs returns a list of Func's from the parsed source.  Func's are sorted
// according to the order of Go functions in the given source.
func (p *Parser) Funcs() []*Func {
	return parseFuncs(p.fset, p.file)
}

func parseFuncs(fset *token.FileSet, f ast.Node) []*Func {
	var funcs []*Func

	// Inspect the AST and find all function declarements and literals
	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncDecl:
			fn := &Func{
				Lbrace:  ToPosition(fset.Position(x.Body.Lbrace)),
				Rbrace:  ToPosition(fset.Position(x.Body.Rbrace)),
				FuncPos: ToPosition(fset.Position(x.Type.Func)),
				node:    x,
			}

			if x.Doc != nil {
				fn.Doc = ToPosition(fset.Position(x.Doc.Pos()))
			}

			funcs = append(funcs, fn)
		case *ast.FuncLit:
			funcs = append(funcs, &Func{
				Lbrace:  ToPosition(fset.Position(x.Body.Lbrace)),
				Rbrace:  ToPosition(fset.Position(x.Body.Rbrace)),
				FuncPos: ToPosition(fset.Position(x.Type.Func)),
				node:    x,
				literal: true,
			})
		}
		return true
	})

	return funcs
}

func enclosingFunc(funcs []*Func, offset int) (*Func, error) {
	var encFunc *Func

	// TODO(arslan) this is iterating over all functions. Benchmark it and see
	// if it's worth it to change it with a more effiecent search function. For
	// now this is enough for us.
	for _, fn := range funcs {
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

func nextFunc(funcs []*Func, offset int) (*Func, error) {
	// find nearest next function
	nextIndex := sort.Search(len(funcs), func(i int) bool {
		if funcs[i].literal {
			return false
		}
		return funcs[i].FuncPos.Offset > offset
	})

	if nextIndex == len(funcs) {
		return nil, errors.New("no functions found")
	}
	return funcs[nextIndex], nil
}

func prevFunc(funcs []*Func, offset int) (*Func, error) {
	// start from the reverse to get the prev function
	reverse(byOffset(funcs))

	prevIndex := sort.Search(len(funcs), func(i int) bool {
		if funcs[i].literal {
			return false
		}
		return funcs[i].FuncPos.Offset < offset
	})

	if prevIndex == len(funcs) {
		return nil, errors.New("no functions found")
	}

	return funcs[prevIndex], nil
}

func reverse(s sort.Interface) {
	for start, end := 0, s.Len()-1; start < end; {
		s.Swap(start, end)
		start++
		end--
	}
}

// byOffset implements sort.Interface for []*Func based on
// the Offset field.
type byOffset []*Func

func (b byOffset) Len() int           { return len(b) }
func (b byOffset) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b byOffset) Less(i, j int) bool { return b[i].FuncPos.Offset < b[j].FuncPos.Offset }
