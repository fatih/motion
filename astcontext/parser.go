package astcontext

import (
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
)

// ParserOptions defines the options that changes the Parser's behavior
type ParserOptions struct {
	// File defines the filename to be parsed
	File string

	// Dir defines the directory to be parsed
	Dir string

	// Src defines the source to be parsed
	Src []byte

	// If enabled parses the comments too
	Comments bool
}

// Parser defines the customized parser
type Parser struct {
	// fset is the default fileset that is passed to the internal parser
	fset *token.FileSet

	// file contains the parsed file
	file *ast.File

	// pkgs contains the parsed packages
	pkgs map[string]*ast.Package
}

// NewParser creates a new Parser reference from the given options
func NewParser(opts *ParserOptions) (*Parser, error) {
	var mode parser.Mode
	if opts != nil && opts.Comments {
		mode = parser.ParseComments
	}

	fset := token.NewFileSet()
	p := &Parser{fset: fset}
	var err error

	switch {
	case opts.File != "":
		p.file, err = parser.ParseFile(fset, opts.File, nil, mode)
		if err != nil {
			return nil, err
		}
	case opts.Dir != "":
		p.pkgs, err = parser.ParseDir(fset, opts.Dir, nil, mode)
		if err != nil {
			return nil, err
		}
	case opts.Src != nil:
		p.file, err = parser.ParseFile(fset, "src.go", opts.Src, mode)
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("file, src or dir is not specified")
	}

	return p, nil
}
