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
		flagDir    = flag.String("dir", "", "Directory to be parsed")
		flagOffset = flag.String("offset", "", "Byte offset of the cursor position")
		flagMode   = flag.String("mode", "",
			"Running mode. One of {enclosing, next, prev}")
		flagShift  = flag.Int("shift", 0, "Shift value for the modes {next, prev}")
		flagFormat = flag.String("format", "gnu",
			"Output format. One of {gnu, json, vim}")
		flagParseComments = flag.Bool("parse-comments", false,
			"Parse comments and add them to AST")
	)

	flag.Parse()

	if *flagMode == "" {
		return errors.New("no mode is passed")
	}

	opts := &astcontext.ParserOptions{
		Comments: *flagParseComments,
	}

	parser := astcontext.NewParser().SetOptions(opts)
	var err error

	switch {
	case *flagFile != "":
		parser, err = parser.ParseFile(*flagFile)
	case *flagDir != "":
		parser, err = parser.ParseDir(*flagDir)
	default:
		return errors.New("-file or -dir is missing")
	}
	if err != nil {
		return err
	}

	var funcs astcontext.Funcs
	var fn *astcontext.Func

	switch *flagMode {
	case "enclosing", "next", "prev":
		if *flagOffset == "" {
			return errors.New("no offset is passed")
		}

		offset, err := strconv.Atoi(*flagOffset)
		if err != nil {
			return err
		}

		funcs := parser.Funcs()
		switch *flagMode {
		case "enclosing":
			fn, err = funcs.EnclosingFunc(offset)
		case "next":
			fn, err = funcs.Declarations().NextFuncShift(offset, *flagShift)
		case "prev":
			fn, err = funcs.Declarations().PrevFuncShift(offset, *flagShift)
		}
	case "funcs":
		// TODO(arslan): change the scope from file to package
		funcs = parser.Funcs().Declarations()
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
