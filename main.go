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
			"Running mode. One of {enclosing, next, prev, decls}")
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

	switch *flagMode {
	case "enclosing", "next", "prev":
		if *flagOffset == "" {
			return errors.New("no offset is passed")
		}

		offset, err := strconv.Atoi(*flagOffset)
		if err != nil {
			return err
		}

		var fn *astcontext.Func

		funcs := parser.Funcs()
		switch *flagMode {
		case "enclosing":
			fn, err = funcs.EnclosingFunc(offset)
		case "next":
			fn, err = funcs.Declarations().NextFuncShift(offset, *flagShift)
		case "prev":
			fn, err = funcs.Declarations().PrevFuncShift(offset, *flagShift)
		}

		// do no return, instead pass it to the editor so it can parse it
		if err != nil {
			return printErr(*flagFormat, err)
		}

		if fn != nil {
			funcs = append(funcs, fn)
		}

		return printResult(*flagFormat, funcs)
	case "decls":
		funcs := parser.Funcs().Declarations()
		types := parser.Types().TopLevel()

		type decl struct {
			Keyword  string `json:"keyword" vim:"keyword"`
			Ident    string `json:"ident" vim:"ident"`
			Full     string `json:"full" vim:"full"`
			Filename string `json:"filename" vim:"filename"`
			Line     int    `json:"line" vim:"line"`
			Col      int    `json:"col" vim:"col"`
		}

		var decls []decl

		for _, t := range types {
			decls = append(decls, decl{
				Keyword:  "type",
				Ident:    t.Signature.Name,
				Full:     t.Signature.Full,
				Filename: t.TypePos.Filename,
				Line:     t.TypePos.Line,
				Col:      t.TypePos.Column,
			})
		}

		for _, f := range funcs {
			decls = append(decls, decl{
				Keyword:  "func",
				Ident:    f.Signature.Name,
				Full:     f.Signature.Full,
				Filename: f.FuncPos.Filename,
				Line:     f.FuncPos.Line,
				Col:      f.FuncPos.Column,
			})
		}

		return printResult(*flagFormat, decls)
	default:
		return fmt.Errorf("wrong mode %q passed", *flagMode)
	}

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
