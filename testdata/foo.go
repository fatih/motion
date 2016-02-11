package main

import (
	"fmt"
)

type T struct{}

type X struct {
	name string
}

func main() {
	fmt.Printf("Hello, world\n")
}

// Bar is something else
func Bar() (string, error) {
	return "vim-go", nil
}

// and this is a multi comment doc. I think docs are really different kind of
// thins
func foo() error {
	a := func() {
		fmt.Println("zeynep")
	}

	a()

	return nil
}
