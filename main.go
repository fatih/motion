package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/fatih/motion/astcontext"
	"github.com/fatih/motion/vim"
)

func main() {
	if err := realMain(); err != nil {
		fmt.Fprint(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func realMain() error {
	var (
		flagFile   = flag.String("file", "", "Filename to be parsed")
		flagOffset = flag.String("offset", "", "Byte offset of the cursor position")
		flagMode   = flag.String("mode", "", "Running mode. One of {enclosing, next, prev}")
		flagShift  = flag.Int("shift", 0, "Shift value for the modes {next, prev}")
		flagFormat = flag.String("format", "gnu",
			"Output format. One of {gnu, json, vim}")
		flagParseComments = flag.Bool("parse-comments", false,
			"Parse comments and add them to AST")
	)

	flag.Parse()

	if *flagOffset == "" {
		return errors.New("no offset is passed")
	}
	if *flagMode == "" {
		return errors.New("no mode is passed")
	}
	if *flagFile == "" {
		return errors.New("no file is passed")
	}

	offset, err := strconv.Atoi(*flagOffset)
	if err != nil {
		return err
	}

	opts := &astcontext.ParserOptions{
		ParseComments: *flagParseComments,
	}

	parser, err := astcontext.NewParser().SetOptions(opts).ParseFile(*flagFile)
	if err != nil {
		return err
	}

	var funcs astcontext.Funcs
	var fn *astcontext.Func

	switch *flagMode {
	case "enclosing":
		fn, err = parser.Funcs().EnclosingFunc(offset)
	case "next":
		fn, err = parser.Funcs().Declarations().NextFuncShift(offset, *flagShift)
	case "prev":
		fn, err = parser.Funcs().Declarations().PrevFuncShift(offset, *flagShift)
	case "funcs":
		// TODO(arslan): change the scope from file to package
		funcs = parser.Funcs()
	default:
		return fmt.Errorf("wrong mode %q passed", *flagMode)
	}

	// do no return, instead pass it to the editor so it can parse it
	if err != nil {
		return printErr(*flagFormat, err)
	}

	if fn != nil {
		funcs = append(funcs, fn)
	}

	return printResult(*flagFormat, funcs)
}

func printErr(mode string, err error) error {
	var res interface{}
	switch mode {
	case "json", "vim":
		res = struct {
			Err string `json:"err" vim:"err"`
		}{
			Err: err.Error(),
		}
	case "gnu":
		res = err
	}

	return printResult(mode, res)
}

func printResult(mode string, res interface{}) error {
	// if it contains only one function, return it as a single function
	if funcs, ok := res.(astcontext.Funcs); ok && len(funcs) == 1 {
		res = funcs[0]
	}

	switch mode {
	case "json":
		b, err := json.MarshalIndent(&res, "", "\t")
		if err != nil {
			return fmt.Errorf("JSON error: %s\n", err)
		}
		os.Stdout.Write(b)
	case "vim":
		b, err := vim.Marshal(&res)
		if err != nil {
			return fmt.Errorf("VIM error: %s\n", err)
		}
		os.Stdout.Write(b)
	case "gnu":
		switch x := res.(type) {
		case *astcontext.Func:
			fmt.Println(x)
		case astcontext.Funcs:
			for _, fn := range x {
				fmt.Println(fn)
			}
		case error:
			fmt.Println(x.Error())
		default:
			fmt.Println(res)
		}
	default:
		return fmt.Errorf("wrong -format value: %q.\n", mode)
	}

	return nil
}
