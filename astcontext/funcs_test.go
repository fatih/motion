package astcontext

import (
	"log"
	"testing"
)

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
		offset     int
		funcOffset int
	}{
		{23, 0}, // var bar = func {}
		{25, 24},
		{32, 24}, // var bar = func {}
		{35, 35}, // func foo() error {
		{53, 35}, // func foo() error {
		{67, 59}, // _ = func() {
		{70, 59}, // _ = func() {
		{85, 35}, // func foo() error {
		{96, 35}, // func foo() error {
		{97, 0},
	}

	for _, pos := range testPos {
		fn, _ := EnclosingFunc([]byte(src), pos.offset)
		log.Println("fn", fn.FuncPos.Offset)

		if fn.FuncPos.Offset != pos.funcOffset {
			t.Errorf("offset %d should belong to func with offset: %d, got: %d",
				pos.offset, pos.funcOffset, fn.FuncPos.Offset)
		}
	}
}
