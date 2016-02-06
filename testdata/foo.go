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

func bar() (string, error) {
	return "vim-go", nil
}

func foo() error {
	a := func() {
		fmt.Println("zeynep")
	}

	a()

	return nil
}
