package astcontext

import (
	"go/ast"
	"go/parser"
	"go/token"
)

// ParserOptions defines the options that changes the Parser's behavior
type ParserOptions struct {
	// If enabled parses the comments too
	Comments bool
}

// Parser defines the customized parser
type Parser struct {
	// fset is the default fileset that is passed to the internal parser
	fset *token.FileSet

	// pkgs contains the parsed packages
	pkgs map[string]*ast.Package

	file *ast.File
	// opts contains the parser options
	opts *ParserOptions
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

// ParseDir parses the given directory path
func (p *Parser) ParseDir(path string) (*Parser, error) {
	var mode parser.Mode
	if p.opts != nil && p.opts.Comments {
		mode = parser.ParseComments
	}

	var err error
	p.pkgs, err = parser.ParseDir(p.fset, path, nil, mode)
	if err != nil {
		return nil, err
	}

	return p, nil
}

// ParseFile parses the given filename
func (p *Parser) ParseFile(filename string) (*Parser, error) {
	var mode parser.Mode
	if p.opts != nil && p.opts.Comments {
		mode = parser.ParseComments
	}

	f, err := parser.ParseFile(p.fset, filename, nil, mode)
	if err != nil {
		return nil, err
	}

	p.file = f
	return p, nil
}

// ParseSrc parses the given Go source code
func (p *Parser) ParseSrc(src []byte) (*Parser, error) {
	var mode parser.Mode
	if p.opts != nil && p.opts.Comments {
		mode = parser.ParseComments
	}

	f, err := parser.ParseFile(p.fset, "src.go", src, mode)
	if err != nil {
		return nil, err
	}
	p.file = f
	return p, nil
}
