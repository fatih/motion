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

func main() {
	if err := realMain(); err != nil {
		fmt.Fprint(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func realMain() error {
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

	var fn *astcontext.Func
	switch *flagMode {
	case "enclosing":
		fn, err = parser.Funcs().EnclosingFunc(offset)
	case "next":
		fn, err = parser.Funcs().Declarations().NextFuncShift(offset, *flagShift)
	case "prev":
		fn, err = parser.Funcs().Declarations().PrevFuncShift(offset, *flagShift)
	default:
		return fmt.Errorf("wrong mode %q passed", *flagMode)
	}

	var res interface{} = fn

	// do no return, instead pass it to the editor so it can parse it
	if err != nil {
		res = struct {
			Err string `json:"err" vim:"err"`
		}{
			Err: err.Error(),
		}
	}

	switch *flagFormat {
	case "json", "gnu", "vim":
	default:
		return fmt.Errorf("wrong -format value: %q.\n", *flagFormat)
	}

	switch *flagFormat {
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
		default:
			fmt.Println(err.Error())
		}
	}

	return nil
}
