package main

import (
	"encoding/json"
	"encoding/xml"
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
	flagFormat = flag.String("format", "plain",
		"Output format. One of {plain, json, xml, vim}")
	flagParseComments = flag.Bool("parse-comments", false,
		"Parse comments and add them to AST")
	flagMode = flag.String("mode", "", "Running mode. One of {enclosing, next, prev}")
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
		fn, err = parser.Funcs().Declarations().NextFunc(offset)
	case "prev":
		fn, err = parser.Funcs().Declarations().PrevFunc(offset)
	default:
		return fmt.Errorf("wrong mode %q passed", *flagMode)
	}

	if err != nil {
		return err
	}

	switch *flagFormat {
	case "json", "plain", "xml", "vim":
	default:
		return fmt.Errorf("wrong -format value: %q.\n", *flagFormat)
	}

	// Print the result.
	switch *flagFormat {
	case "json":
		b, err := json.MarshalIndent(&fn, "", "\t")
		if err != nil {
			return fmt.Errorf("JSON error: %s\n", err)
		}
		os.Stdout.Write(b)
	case "xml":
		b, err := xml.MarshalIndent(&fn, "", "\t")
		if err != nil {
			return fmt.Errorf("XML error: %s\n", err)
		}
		os.Stdout.Write(b)
	case "vim":
		b, err := vim.Marshal(&fn)
		if err != nil {
			return fmt.Errorf("XML error: %s\n", err)
		}
		os.Stdout.Write(b)
	case "plain":
		fmt.Print(fn)
	}

	return nil
}
