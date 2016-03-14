package main

import (
	"fmt"
)

var a = func() { fmt.Println("Ankara") }

func main() {
	fmt.Println("vim-go: text-objects")
	a()
}

// Bar is a simple function declaration
func Bar() (string, error) {
	return "vim-go", nil
}

// and this is a multi line doc comment. example is
// just a function that prints zeynep
func example() error {
	a := func() {
		fmt.Println("zeynep")
	}

	a()

	return nil
}
