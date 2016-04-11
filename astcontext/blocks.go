package astcontext

import (
	"errors"
	"go/ast"
	"sort"
)

// Block represents anything, that contain block: Func, If, For, Range, Switch, TypeSwitch, Select.
type Block struct {
	BlockPos *Position `json:"block" vim:"block"`
	Lbrace   *Position `json:"lbrace" vim:"lbrace"` // position of "{"
	Rbrace   *Position `json:"rbrace" vim:"rbrace"` // position of "}"

	node ast.Node
}

// Blocks represents a list of Block's.
type Blocks []*Block

// Blocks returns a list of Block's from the parsed source.
func (p *Parser) Blocks() Blocks {
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

	var blocks []*Block
	inspect := func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.ForStmt:
			bl := &Block{
				BlockPos: ToPosition(p.fset.Position(x.For)),
				node:     x,
			}

			if x.Body != nil {
				bl.Lbrace = ToPosition(p.fset.Position(x.Body.Lbrace))
				bl.Rbrace = ToPosition(p.fset.Position(x.Body.Rbrace))
			}
			blocks = append(blocks, bl)

		case *ast.RangeStmt:
			bl := &Block{
				BlockPos: ToPosition(p.fset.Position(x.For)),
				node:     x,
			}

			if x.Body != nil {
				bl.Lbrace = ToPosition(p.fset.Position(x.Body.Lbrace))
				bl.Rbrace = ToPosition(p.fset.Position(x.Body.Rbrace))
			}
			blocks = append(blocks, bl)

		case *ast.IfStmt:
			bl := &Block{
				BlockPos: ToPosition(p.fset.Position(x.If)),
				node:     x,
			}

			if x.Body != nil {
				bl.Lbrace = ToPosition(p.fset.Position(x.Body.Lbrace))
				bl.Rbrace = ToPosition(p.fset.Position(x.Body.Rbrace))
			}
			blocks = append(blocks, bl)

		case *ast.SelectStmt:
			bl := &Block{
				BlockPos: ToPosition(p.fset.Position(x.Select)),
				node:     x,
			}

			if x.Body != nil {
				bl.Lbrace = ToPosition(p.fset.Position(x.Body.Lbrace))
				bl.Rbrace = ToPosition(p.fset.Position(x.Body.Rbrace))
			}
			blocks = append(blocks, bl)

		case *ast.SwitchStmt:
			bl := &Block{
				BlockPos: ToPosition(p.fset.Position(x.Switch)),
				node:     x,
			}

			if x.Body != nil {
				bl.Lbrace = ToPosition(p.fset.Position(x.Body.Lbrace))
				bl.Rbrace = ToPosition(p.fset.Position(x.Body.Rbrace))
			}
			blocks = append(blocks, bl)

		case *ast.FuncDecl:
			bl := &Block{
				BlockPos: ToPosition(p.fset.Position(x.Type.Func)),
				node:     x,
			}

			if x.Body != nil {
				bl.Lbrace = ToPosition(p.fset.Position(x.Body.Lbrace))
				bl.Rbrace = ToPosition(p.fset.Position(x.Body.Rbrace))
			}
			blocks = append(blocks, bl)

		case *ast.FuncLit:
			bl := &Block{
				BlockPos: ToPosition(p.fset.Position(x.Type.Func)),
				node:     x,
			}

			if x.Body != nil {
				bl.Lbrace = ToPosition(p.fset.Position(x.Body.Lbrace))
				bl.Rbrace = ToPosition(p.fset.Position(x.Body.Rbrace))
			}
			blocks = append(blocks, bl)

		case *ast.TypeSwitchStmt:
			bl := &Block{
				BlockPos: ToPosition(p.fset.Position(x.Switch)),
				node:     x,
			}

			if x.Body != nil {
				bl.Lbrace = ToPosition(p.fset.Position(x.Body.Lbrace))
				bl.Rbrace = ToPosition(p.fset.Position(x.Body.Rbrace))
			}
			blocks = append(blocks, bl)

		}
		return true
	}

	for _, file := range files {
		// Inspect the AST and find all function declarements and literals
		ast.Inspect(file, inspect)
	}

	return blocks
}

// BorderType represents Block's border type: left or right.
type borderType string

const (
	leftBorderType  borderType = "left"
	rightBorderType borderType = "right"
)

// Border is helper struct, that represents Block's border
type border struct {
	Block *Block
	Pos   *Position
	Type  borderType
}

// Borders represent a list of Border's
type borders []*border

// Borders returns Borders for Blocks
func (b Blocks) borders() borders {
	var borders []*border
	for _, block := range b {
		bdl := &border{
			Block: block,
			Pos:   block.Lbrace,
			Type:  leftBorderType,
		}
		bdr := &border{
			Block: block,
			Pos:   block.Rbrace,
			Type:  rightBorderType,
		}
		borders = append(borders, bdl, bdr)
	}

	return borders
}

// Sort implementation for Borders.
func (b borders) Len() int      { return len(b) }
func (b borders) Swap(i, j int) { b[i], b[j] = b[j], b[i] }
func (b borders) Less(i, j int) bool {
	return b[i].Pos.Offset < b[j].Pos.Offset
}

// FindBlock returns innermost Block within Borders for offset,
// or error, if there is no such Block.
func (b borders) findBlock(offset int) (*Block, error) {
	sort.Sort(sort.Reverse(b))
	rightTypes := 0

	for _, border := range b {
		// skip borders, that greater than offset
		if border.Pos.Offset >= offset {
			continue
		}

		if border.Type == leftBorderType {
			if rightTypes == 0 {
				return border.Block, nil
			}
			rightTypes--
		} else {
			rightTypes++
		}
	}

	return nil, errors.New("no block found")
}

// EnclosingBlock returns innermost Block within Blocks for given offset
// or error, if there is no such Block.
func (b Blocks) EnclosingBlock(offset int) (*Block, error) {
	return b.borders().findBlock(offset)
}
