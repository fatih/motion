package astcontext

import "go/token"

// Position describes a function position
type Position struct {
	Filename string `json:"filename" vim:"filename"`
	Offset   int    `json:"offset" vim:"offset"` // offset, starting at 0
	Line     int    `json:"line" vim:"line"`     // line number, starting at 1
	Column   int    `json:"col" vim:"col"`       // column number, starting at 1 (byte count)
}

// ToPosition returns a Position from the given token.Position
func ToPosition(pos token.Position) *Position {
	return &Position{
		Filename: pos.Filename,
		Offset:   pos.Offset,
		Line:     pos.Line,
		Column:   pos.Column,
	}
}

// IsValid returns true if position is valid
func (pos Position) IsValid() bool { return pos.Line > 0 }
