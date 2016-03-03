// Package astcontext provides context aware utilities to be used within
// editors.
package astcontext

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/types"
	"sort"
)

// Func represents a declared (*ast.FuncDecl) or an anonymous (*ast.FuncLit) Go
// function
type Func struct {
	// signature of the function
	Signature string `json:"sig" vim:"sig"`

	// position of the "func" keyword
	FuncPos *Position `json:"func" vim:"func"`
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

// signature returns the function signature as a string.
func signature(node ast.Node) string {
	getParams := func(list []*ast.Field) string {
		out := ""
		for i, p := range list {
			for j, n := range p.Names {
				out += n.Name
				if len(p.Names) != j+1 {
					out += ", "
				}
			}

			if len(p.Names) != 0 {
				out += " "
			}

			buf := new(bytes.Buffer)
			types.WriteExpr(buf, p.Type)
			out += buf.String()

			if len(list) != i+1 {
				out += ", "
			}
		}
		return out
	}

	switch x := node.(type) {
	case *ast.FuncDecl:
		recv := ""
		funcName := x.Name.Name
		input := ""
		output := ""
		multiOutput := false

		if x.Type.Params != nil {
			input = getParams(x.Type.Params.List)
		}
		if x.Type.Results != nil {
			output = getParams(x.Type.Results.List)
			multiOutput = len(x.Type.Results.List) > 1
		}
		if x.Recv != nil {
			recv = getParams(x.Recv.List)
		}

		sig := "func "

		if recv != "" {
			sig += fmt.Sprintf("(%s) ", recv)
		}

		sig += fmt.Sprintf("%s", funcName)

		sig += "("
		if input != "" {
			sig += fmt.Sprintf("%s", input)
		}
		sig += ")"

		if output != "" {
			if multiOutput {
				sig += fmt.Sprintf(" (%s)", output)
			} else {
				sig += fmt.Sprintf(" %s", output)
			}
		}

		return sig
	case *ast.FuncLit:
		input := ""
		output := ""
		multiOutput := false

		if x.Type.Params != nil {
			input = getParams(x.Type.Params.List)
		}
		if x.Type.Results != nil {
			output = getParams(x.Type.Results.List)
			multiOutput = len(x.Type.Results.List) > 1
		}

		sig := "func"

		sig += "("
		if input != "" {
			sig += fmt.Sprintf("%s", input)
		}
		sig += ")"

		if output != "" {
			if multiOutput {
				sig += fmt.Sprintf(" (%s)", output)
			} else {
				sig += fmt.Sprintf(" %s", output)
			}
		}

		return sig
	default:
		return "<UNKNOWN>"
	}
}

func (f *Func) String() string {
	// Print according to GNU error messaging format
	// https://www.gnu.org/prep/standards/html_node/Errors.html
	switch x := f.node.(type) {
	case *ast.FuncDecl:
		return fmt.Sprintf("%s:%d:%d %s",
			f.FuncPos.Filename, f.FuncPos.Line, f.FuncPos.Column, x.Name.Name)
	default:
		return fmt.Sprintf("%s:%d:%d %s",
			f.FuncPos.Filename, f.FuncPos.Line, f.FuncPos.Column, "(literal)")
	}
}

// Funcs returns a list of Func's from the parsed source. Func's are sorted
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

			fn.Signature = signature(x)
			funcs = append(funcs, fn)
		case *ast.FuncLit:
			fn := &Func{
				Lbrace:  ToPosition(p.fset.Position(x.Body.Lbrace)),
				Rbrace:  ToPosition(p.fset.Position(x.Body.Rbrace)),
				FuncPos: ToPosition(p.fset.Position(x.Type.Func)),
				node:    x,
			}

			fn.Signature = signature(x)
			funcs = append(funcs, fn)
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
	return f.nextFuncShift(offset, 0)
}

// NextFuncShift returns the nearest next Func for the given offset. Shift
// shifts the index before returning. This is useful to get the second nearest
// next function (shift being 1), third nearest next function (shift being 2),
// etc...
func (f Funcs) NextFuncShift(offset, shift int) (*Func, error) {
	return f.nextFuncShift(offset, shift)
}

// PrevFunc returns the nearest previous *Func for the given offset.
func (f Funcs) PrevFunc(offset int) (*Func, error) {
	return f.prevFuncShift(offset, 0)
}

// PrevFuncShift returns the nearest previous Func for the given offset. Shift
// shifts the index before returning. This is useful to get the second nearest
// previous function (shift being 1), third nearest previous function (shift
// being 2), etc...
func (f Funcs) PrevFuncShift(offset, shift int) (*Func, error) {
	return f.prevFuncShift(offset, shift)
}

// nextFuncShift returns the nearest next function for the given offset and
// shift index. If index is zero it returns the nearest next function. If shift
// is non zero positive number it returns the function shifted by the given
// number. i.e: [a, b, c, d] if the nearest func is b (shift 0), shift with
// value 1 returns c, 2 returns d and anything larger returns an error.
func (f Funcs) nextFuncShift(offset, shift int) (*Func, error) {
	if shift < 0 {
		return nil, errors.New("shift can't be negative")
	}

	// find nearest next function
	nextIndex := sort.Search(len(f), func(i int) bool {
		return f[i].FuncPos.Offset > offset
	})

	if nextIndex >= len(f) {
		return nil, errors.New("no functions found")
	}

	fn := f[nextIndex]

	// if our position is inside the doc, increase the shift by one to pick up
	// the next function. This assumes that people editing a doc of a func want
	// to pick up the next function instead of the current function.
	if fn.Doc != nil && fn.Doc.IsValid() {
		if fn.Doc.Offset <= offset && offset < fn.FuncPos.Offset {
			shift += 1
		}
	}

	if nextIndex+shift >= len(f) {
		return nil, errors.New("no functions found")
	}

	return f[nextIndex+shift], nil
}

// prevFuncShift returns the nearest previous *Func for the given offset and
// shift index. If index is zero it returns the nearest previous function. If
// shift is non zero positive number it returns the function shifted by the
// given number. i.e: [a, b, c, d] if the nearest previous func is c (shift 0),
// shift with value 1 returns b, 2 returns a and anything larger returns an
// error.
func (f Funcs) prevFuncShift(offset, shift int) (*Func, error) {
	if shift < 0 {
		return nil, errors.New("shift can't be negative")
	}

	// start from the reverse to get the prev function
	f.Reserve()

	prevIndex := sort.Search(len(f), func(i int) bool {
		return f[i].FuncPos.Offset < offset
	})

	if prevIndex+shift >= len(f) {
		return nil, errors.New("no functions found")
	}

	return f[prevIndex+shift], nil
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
