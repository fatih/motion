package main

import (
	"fmt"
)

var a = func() { fmt.Println("tokyo") }

func main() {
	fmt.Printf("Hello, world\n")
	a()
}

// Bar is something else
func Bar() (string, error) {
	return "vim-go", nil
}

// and this is a multi comment doc. I think
// docs are really different kind of things,
// bla bla
func example() error {
	a := func() {
		fmt.Println("zeynep")
	}

	a()

	return nil
}
