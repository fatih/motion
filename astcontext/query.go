package astcontext

import (
	"errors"
	"fmt"
)

// Decl specifies the result of the "decls" mode
type Decl struct {
	Keyword  string `json:"keyword" vim:"keyword"`
	Ident    string `json:"ident" vim:"ident"`
	Full     string `json:"full" vim:"full"`
	Filename string `json:"filename" vim:"filename"`
	Line     int    `json:"line" vim:"line"`
	Col      int    `json:"col" vim:"col"`
}

// Result is the common result of any motion query.
// It contains a query-specific result element.
type Result struct {
	Mode string `json:"mode" vim:"mode"`

	Decls []Decl `json:"decls" vim:"decls"`
	Func  *Func  `json:"func" vim:"fn"`
}

// Query specifies a single query to the parser
type Query struct {
	Mode     string
	Offset   int
	Shift    int
	Includes []string
}

// Run runs the given query and returns the result
func (p *Parser) Run(query *Query) (*Result, error) {
	if query == nil {
		return nil, errors.New("query is nil")
	}

	switch query.Mode {
	case "enclosing", "next", "prev":
		var fn *Func
		var err error

		funcs := p.Funcs()
		switch query.Mode {
		case "enclosing":
			fn, err = funcs.EnclosingFunc(query.Offset)
		case "next":
			fn, err = funcs.Declarations().NextFuncShift(query.Offset, query.Shift)
		case "prev":
			fn, err = funcs.Declarations().PrevFuncShift(query.Offset, query.Shift)
		}

		// do no return, instead pass it to the editor so it can parse it
		if err != nil {
			return nil, err
		}

		return &Result{
			Mode: query.Mode,
			Func: fn,
		}, nil
	case "decls":
		funcs := p.Funcs().Declarations()
		types := p.Types().TopLevel()

		var decls []Decl

		for _, incl := range query.Includes {
			switch incl {
			case "type":
				for _, t := range types {
					decls = append(decls, Decl{
						Keyword:  "type",
						Ident:    t.Signature.Name,
						Full:     t.Signature.Full,
						Filename: t.TypePos.Filename,
						Line:     t.TypePos.Line,
						Col:      t.TypePos.Column,
					})
				}
			case "func":
				for _, f := range funcs {
					decls = append(decls, Decl{
						Keyword:  "func",
						Ident:    f.Signature.Name,
						Full:     f.Signature.Full,
						Filename: f.FuncPos.Filename,
						Line:     f.FuncPos.Line,
						Col:      f.FuncPos.Column,
					})
				}
			}
		}

		return &Result{
			Mode:  query.Mode,
			Decls: decls,
		}, nil
	default:
		return nil, fmt.Errorf("wrong mode %q passed", query.Mode)
	}
}
