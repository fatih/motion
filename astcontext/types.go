package astcontext

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/types"
)

// TypeSignature represents a type declaration signature
type TypeSignature struct {
	// Full signature representation
	Full string `json:"full" vim:"full"`

	// Name of the type declaration
	Name string `json:"name" vim:"name"`

	// Type is the representation of the type of a TypeSpec. Ie.: type MyInt
	// int. Type is here: "int".
	Type string `json:"type" vim:"type"`
}

// Type represents a type declaration
type Type struct {
	// Signature is the simplified representation of the Type declaration
	Signature *TypeSignature `json:"sig" vim:"sig"`

	// position of the TypeSpec's ident
	TypePos *Position `json:"type" vim:"type"`

	// position of the doc comment
	Doc *Position `json:"doc,omitempty" vim:"doc,omitempty"`

	node *ast.TypeSpec
}

// Types represents a list of type declarations
type Types []*Type

// NewTypeSignature returns a TypeSignature from the given typespec node
func NewTypeSignature(node *ast.TypeSpec) *TypeSignature {
	buf := new(bytes.Buffer)
	types.WriteExpr(buf, node.Type)

	sig := &TypeSignature{
		Name: node.Name.Name,
		Type: buf.String(),
	}

	sig.Full = fmt.Sprintf("type %s %s", sig.Name, sig.Type)
	return sig
}

// Types returns a list of Type's from the parsed source. Type's are sorted
// according to the order of Go type declaration in the given source.
func (p *Parser) Types() Types {
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

	var typs []*Type
	inspect := func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.TypeSpec:
			tp := &Type{
				TypePos: ToPosition(p.fset.Position(x.Name.Pos())),
				node:    x,
			}

			if x.Doc != nil {
				tp.Doc = ToPosition(p.fset.Position(x.Doc.Pos()))
			}

			tp.Signature = NewTypeSignature(x)

			typs = append(typs, tp)
		}
		return true
	}

	for _, file := range files {
		// Inspect the AST and find all type declarations
		ast.Inspect(file, inspect)
	}

	return typs
}

// TopLevel returns a copy of Types with only top level type declarations
func (t Types) TopLevel() Types {
	var typs []*Type
	for _, typ := range t {
		if typ.TypePos.Column != 6 {
			continue
		}

		typs = append(typs, typ)
	}
	return typs
}
