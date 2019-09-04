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

// FuncSignature defines the function signature
type FuncSignature struct {
	// Full signature representation
	Full string `json:"full" vim:"full"`

	// Receiver representation. Empty for non methods.
	Recv string `json:"recv" vim:"recv"`

	// Name of the function. Empty for function literals.
	Name string `json:"name" vim:"name"`

	// Input arguments of the function, if present
	In string `json:"in" vim:"in"`

	// Output argument of the function, if present
	Out string `json:"out" vim:"out"`
}

func (s *FuncSignature) String() string { return s.Full }

// Func represents a declared (*ast.FuncDecl) or an anonymous (*ast.FuncLit) Go
// function
type Func struct {
	// Signature of the function
	Signature *FuncSignature `json:"sig" vim:"sig"`

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

// NewFuncSignature returns a function signature from the given node. Node should
// be of type *ast.FuncDecl or *ast.FuncLit
func NewFuncSignature(node ast.Node) *FuncSignature {
	getParams := func(list []*ast.Field) string {
		var named bool
		buf := new(bytes.Buffer)
		for i, p := range list {
			for j, n := range p.Names {
				buf.WriteString(n.Name)
				if len(p.Names) != j+1 {
					buf.WriteString(", ")
				}
			}

			if len(p.Names) != 0 {
				named = true
				buf.WriteString(" ")
			}

			types.WriteExpr(buf, p.Type)

			if len(list) != i+1 {
				named = true
				buf.WriteString(", ")
			}
		}
		if named {
			return fmt.Sprintf("(%s)", buf.String())
		}
		return buf.String()
	}

	switch x := node.(type) {
	case *ast.FuncDecl:
		sig := &FuncSignature{
			Name: x.Name.Name,
		}

		if x.Type.Params != nil {
			sig.In = getParams(x.Type.Params.List)
		}
		if x.Type.Results != nil {
			sig.Out = getParams(x.Type.Results.List)
		}
		if x.Recv != nil {
			sig.Recv = getParams(x.Recv.List)
		}

		full := "func "

		if sig.Recv != "" {
			full += fmt.Sprintf("%s ", sig.Recv)
		}

		full += fmt.Sprintf("%s", sig.Name)

		if sig.In != "" {
			full += fmt.Sprintf("%s", sig.In)
		} else {
			full += "()"
		}

		if sig.Out != "" {
			full += fmt.Sprintf(" %s", sig.Out)
		}

		sig.Full = full
		return sig
	case *ast.FuncLit:
		sig := &FuncSignature{}

		if x.Type.Params != nil {
			sig.In = getParams(x.Type.Params.List)
		}
		if x.Type.Results != nil {
			sig.Out = getParams(x.Type.Results.List)
		}

		full := "func"

		if sig.In != "" {
			full += fmt.Sprintf("%s", sig.In)
		} else {
			full += "()"
		}

		if sig.Out != "" {
			full += fmt.Sprintf(" %s", sig.Out)
		}

		sig.Full = full
		return sig
	default:
		return &FuncSignature{Full: "UNKNOWN"}
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
	var files []*ast.File
	if p.file != nil {
		files = append(files, p.file)
	}

	if p.pkgs != nil {
		for _, pkg := range p.pkgs {
			for _, f := range pkg.Files {
				files = append(files, f)
			}
		}
	}

	var funcs []*Func
	inspect := func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncDecl:
			fn := &Func{
				FuncPos: ToPosition(p.fset.Position(x.Type.Func)),
				node:    x,
			}

			// can be nil for forward declarations
			if x.Body != nil {
				fn.Lbrace = ToPosition(p.fset.Position(x.Body.Lbrace))
				fn.Rbrace = ToPosition(p.fset.Position(x.Body.Rbrace))
			}

			if x.Doc != nil {
				fn.Doc = ToPosition(p.fset.Position(x.Doc.Pos()))
			}

			fn.Signature = NewFuncSignature(x)
			funcs = append(funcs, fn)
		case *ast.FuncLit:
			fn := &Func{
				Lbrace:  ToPosition(p.fset.Position(x.Body.Lbrace)),
				Rbrace:  ToPosition(p.fset.Position(x.Body.Rbrace)),
				FuncPos: ToPosition(p.fset.Position(x.Type.Func)),
				node:    x,
			}

			fn.Signature = NewFuncSignature(x)
			funcs = append(funcs, fn)
		}
		return true
	}

	for _, file := range files {
		// Inspect the AST and find all function declarements and literals
		ast.Inspect(file, inspect)
	}

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
			shift++
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

func (f Funcs) Len() int      { return len(f) }
func (f Funcs) Swap(i, j int) { f[i], f[j] = f[j], f[i] }
func (f Funcs) Less(i, j int) bool {
	return f[i].FuncPos.Offset < f[j].FuncPos.Offset
}

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
