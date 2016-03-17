package astcontext

import "testing"

func TestEnclosingBlock(t *testing.T) {
	var src = `package main

import "log"

func foo() error{
	for i:=0; i<10; i++ {
		log.Println(i)
	}
	return nil
}`

	opts := &ParserOptions{Src: []byte(src)}
	parser, err := NewParser(opts)
	if err != nil {
		t.Fatal(err)
	}
	blocks := parser.Blocks()

	pos := 80
	wanted := 67
	bl, err := blocks.EnclosingBlock(pos)
	if err != nil {
		t.Logf("blocks len = %d\n", len(blocks))

		borders := blocks.borders()
		t.Logf("borders len = %d\n", len(borders))

		t.Logf("borders[0] offset = %d\n", borders[0].Pos.Offset)
		t.Logf("borders[0] line = %d\n", borders[0].Pos.Line)
		t.Logf("borders[0] column = %d\n", borders[0].Pos.Column)

		t.Logf("borders[1] offset = %d\n", borders[1].Pos.Offset)
		t.Logf("borders[1] line = %d\n", borders[1].Pos.Line)
		t.Logf("borders[1] column = %d\n", borders[1].Pos.Column)

		t.Fatal(err)
	}

	if bl.Lbrace.Offset != wanted {
		t.Fatalf("Wanted %d, got = %d", wanted, bl.Lbrace.Offset)

		//for _, b := range blocks {
		//t.Logf("", b.BlockPos.Offset)
		//}
	}
}
