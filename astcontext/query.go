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

// Comment specified the result of the "comment" mode.
type Comment struct {
	StartLine int `json:"startLine" vim:"startLine"`
	StartCol  int `json:"startCol" vim:"startCol"`
	EndLine   int `json:"endLine" vim:"endLine"`
	EndCol    int `json:"endCol" vim:"endCol"`
}

// Result is the common result of any motion query.
// It contains a query-specific result element.
type Result struct {
	Mode string `json:"mode" vim:"mode"`

	Comment Comment `json:"comment,omitempty" vim:"comment,omitempty"`
	Decls   []Decl  `json:"decls,omitempty" vim:"decls,omitempty"`
	Func    *Func   `json:"func,omitempty" vim:"fn,omitempty"`
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
	case "comment":
		var comment *Comment
		for _, c := range p.file.Comments {
			if int(c.Pos()) <= query.Offset+1 && int(c.End()) >= query.Offset {
				start := p.fset.Position(c.Pos())
				end := p.fset.Position(c.End())
				comment = &Comment{
					StartLine: start.Line,
					StartCol:  start.Column,
					EndLine:   end.Line,
					EndCol:    end.Column,
				}
				break
			}
		}

		if comment == nil {
			return nil, errors.New("no comment block at cursor position")
		}

		return &Result{
			Comment: *comment,
			Mode:    query.Mode,
		}, nil
	default:
		return nil, fmt.Errorf("wrong mode %q passed", query.Mode)
	}
}
