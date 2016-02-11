package astcontext

import (
	"go/ast"
	"go/parser"
	"go/token"
)

// ParserOptions defines the options that changes the Parser's behavior
type ParserOptions struct {
	ParseComments bool
}

// Parser defines the customized parser
type Parser struct {
	// fset is the default fileset that is passed to the internal parser
	fset *token.FileSet

	// file contains the current parsed file. In the future we might have
	// multiple files.
	file *ast.File

	// opts contains the parser options
	opts *ParserOptions

	// err includes any intermediate error
	err error
}

// NewParser creates a new Parser reference
func NewParser() *Parser {
	return &Parser{
		fset: token.NewFileSet(),
	}
}

// SetOptions set the parser options. If nil the default parser options are
// used.
func (p *Parser) SetOptions(opts *ParserOptions) *Parser {
	if opts == nil {
		return p
	}
	p.opts = opts
	return p
}

// ParseFile parses the given filename
func (p *Parser) ParseFile(filename string) *Parser {
	var mode parser.Mode
	if p.opts != nil && p.opts.ParseComments {
		mode = parser.ParseComments
	}

	f, err := parser.ParseFile(p.fset, filename, nil, mode)
	if err != nil {
		p.err = err
		return p
	}
	p.file = f
	return p
}

// ParseSrc parses the given Go source code
func (p *Parser) ParseSrc(src []byte) *Parser {
	var mode parser.Mode
	if p.opts != nil && p.opts.ParseComments {
		mode = parser.ParseComments
	}

	f, err := parser.ParseFile(p.fset, "src.go", src, mode)
	if err != nil {
		p.err = err
		return p
	}
	p.file = f
	return p
}
