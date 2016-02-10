package astcontext

import "testing"

func TestEnclosingFunc(t *testing.T) {
	var src = `package main

var bar = func() {}

func foo() error {
	_ = func() {
		// -------
	}
	return nil
}
`

	testPos := []struct {
		offset       int
		lbraceOffset int
	}{
		{30, 0},
		{31, 31}, // var bar = func {}
		{32, 31}, // var bar = func {}
		{52, 52}, // func foo() error {
		{53, 52}, // func foo() error {
		{67, 66}, // _ = func() {
		{70, 66}, // _ = func() {
		{85, 52}, // func foo() error {
		{96, 52}, // func foo() error {
		{97, 0},
	}

	for _, pos := range testPos {
		fn, _ := EnclosingFunc([]byte(src), pos.offset)

		if fn.Lbrace.Offset != pos.lbraceOffset {
			t.Errorf("offset %d should belong to func with offset: %d, got: %d",
				pos.offset, pos.lbraceOffset, fn.Lbrace.Offset)
		}
	}
}
