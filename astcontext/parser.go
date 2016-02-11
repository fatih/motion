package astcontext

import (
	"go/ast"
	"go/parser"
	"go/token"
)

// ParserOptions defines the options that changes the Parser's behavior
type ParseOptions struct {
	ParseComments bool
}

// Parser defines the customized parser
type Parser struct {
	fset *token.FileSet
	file *ast.File

	// err includes any intermediate error
	err error
}

// NewParser creates a new Parser reference
func NewParser(opts *ParseOptions) *Parser {
	return &Parser{
		fset: token.NewFileSet(),
	}
}

// ParseFile parses the given filename
func (p *Parser) ParseFile(filename string) *Parser {
	f, err := parser.ParseFile(p.fset, filename, nil, parser.ParseComments)
	if err != nil {
		p.err = err
		return p
	}
	p.file = f
	return p
}

// ParseSrc parses the given Go source code
func (p *Parser) ParseSrc(src []byte) *Parser {
	f, err := parser.ParseFile(p.fset, "src.go", src, parser.ParseComments)
	if err != nil {
		p.err = err
		return p
	}
	p.file = f
	return p
}
