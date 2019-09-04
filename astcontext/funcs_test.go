package astcontext

import (
	"fmt"
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
		{25, 24},
		{32, 24}, // var bar = func {}
		{35, 35}, // func foo() error {
		{53, 35}, // func foo() error {
		{67, 59}, // _ = func() {
		{70, 59}, // _ = func() {
		{85, 35}, // func foo() error {
		{96, 35}, // func foo() error {
	}

	opts := &ParserOptions{Src: []byte(src)}
	parser, err := NewParser(opts)
	if err != nil {
		t.Fatal(err)
	}
	funcs := parser.Funcs()

	for _, pos := range testPos {
		fn, err := funcs.EnclosingFunc(pos.offset)
		if err != nil {
			fmt.Printf("err = %+v\n", err)
			continue
		}

		if fn.FuncPos.Offset != pos.funcOffset {
			t.Errorf("offset %d should belong to func with offset: %d, got: %d",
				pos.offset, pos.funcOffset, fn.FuncPos.Offset)
		}
	}
}

func TestNextFuncComment(t *testing.T) {
	var src = `package main

// Comment foo
// Comment bar
func foo() error {
	_ = func() {
		// -------
	}
	return nil
}

func bar() error {
	return nil
}`

	testPos := []struct {
		start int
		want  int
	}{
		{start: 14, want: 108},
		{start: 29, want: 108},
	}

	opts := &ParserOptions{
		Src:      []byte(src),
		Comments: true,
	}
	parser, err := NewParser(opts)
	if err != nil {
		t.Fatal(err)
	}
	funcs := parser.Funcs().Declarations()

	for _, pos := range testPos {
		fn, _ := funcs.NextFunc(pos.start)

		if fn.FuncPos.Offset != pos.want {
			t.Errorf("offset %d should pick func with offset: %d, got: %d",
				pos.start, pos.want, fn.FuncPos.Offset)
		}
	}
}

func TestFunc_Signature(t *testing.T) {
	var src = `package main

var a = func() { fmt.Println("tokyo") }

func foo(
	a int,
	b string,
	c bool,
) (
	bool,
	error,
) {
	return false, nil
}

func foo(a, b int, foo string) (string, error) {
	_ = func() {
		// -------
	}
	return nil
}

func (q *qaz) example(x,y,z int) error {
	_ = func(foo int) error {
		return nil
	}
	_ = func() (err error) {
		return nil
	}
}

func example() {}

func variadic(x ...string) {}

func bar(x int) error {
	return nil
}

func namedSingleOut() (err error) {
	return nil
}

func namedMultipleOut() (err error, res string) {
	return nil
}`

	testFuncs := []struct {
		want string
	}{
		{want: "func()"},
		{want: "func foo(a int, b string, c bool) (bool, error)"},
		{want: "func foo(a, b int, foo string) (string, error)"},
		{want: "func()"},
		{want: "func (q *qaz) example(x, y, z int) error"},
		{want: "func(foo int) error"},
		{want: "func() (err error)"},
		{want: "func example()"},
		{want: "func variadic(x ...string)"},
		{want: "func bar(x int) error"},
		{want: "func namedSingleOut() (err error)"},
		{want: "func namedMultipleOut() (err error, res string)"},
	}

	opts := &ParserOptions{
		Src: []byte(src),
	}
	parser, err := NewParser(opts)
	if err != nil {
		t.Fatal(err)
	}

	funcs := parser.Funcs()

	for i, fn := range funcs {
		fmt.Printf("[%d] %s\n", i, fn.Signature.Full)
		if fn.Signature.Full != testFuncs[i].want {
			t.Errorf("function signatures\n\twant: %s\n\tgot : %s",
				testFuncs[i].want, fn.Signature)
		}
	}
}

func TestFuncs_NoFuncs(t *testing.T) {
	var src = `package foo`

	opts := &ParserOptions{
		Src: []byte(src),
	}
	parser, err := NewParser(opts)
	if err != nil {
		t.Fatal(err)
	}

	funcs := parser.Funcs()
	if len(funcs) != 0 {
		t.Errorf("There should be no functions, but got %d", len(funcs))

	}
}
