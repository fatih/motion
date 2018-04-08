package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

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
		flagOffset = flag.Int("offset", 0, "Byte offset of the cursor position")
		flagMode   = flag.String("mode", "",
			"Running mode. One of {enclosing, next, prev, decls, comment}")
		flagInclude = flag.String("include", "",
			"Included declarations for mode {decls}. Comma delimited. Options: {func, type}")
		flagShift         = flag.Int("shift", 0, "Shift value for the modes {next, prev}")
		flagFormat        = flag.String("format", "json", "Output format. One of {json, vim}")
		flagParseComments = flag.Bool("parse-comments", false,
			"Parse comments and add them to AST")
	)

	flag.Parse()
	if flag.NFlag() == 0 {
		flag.Usage()
		return nil
	}

	if *flagMode == "" {
		return errors.New("no mode is passed")
	}

	if *flagMode == "comment" {
		*flagParseComments = true
	}

	opts := &astcontext.ParserOptions{
		Comments: *flagParseComments,
		File:     *flagFile,
		Dir:      *flagDir,
	}

	parser, err := astcontext.NewParser(opts)
	if err != nil {
		return err
	}

	query := &astcontext.Query{
		Mode:     *flagMode,
		Offset:   *flagOffset,
		Shift:    *flagShift,
		Includes: strings.Split(*flagInclude, ","),
	}

	result, err := parser.Run(query)

	var res interface{}

	res = result
	if err != nil {
		res = struct {
			Err string `json:"err" vim:"err"`
		}{
			Err: err.Error(),
		}
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
	default:
		return fmt.Errorf("wrong -format value: %q.\n", *flagFormat)
	}

	return nil
}
