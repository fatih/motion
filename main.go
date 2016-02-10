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
)

var (
	fileFlag   = flag.String("file", "", "Filename to be parsed")
	offsetFlag = flag.String("offset", "", "Byte offset of the cursor position")
	formatFlag = flag.String("format", "plain", "Output format. One of {plain, json, xml}")

	// not enabled
	modeFlag = flag.String("mode", "", "type of the query mode. One of {func}")
)

func main() {
	if err := realMain(); err != nil {
		fmt.Fprintf(os.Stderr, "motion: %s\n", err.Error())
		os.Exit(1)
	}
}

func realMain() error {
	flag.Parse()

	if *offsetFlag == "" {
		return errors.New("no offset is passed")
	}

	if *fileFlag == "" {
		return errors.New("no file is passed")
	}

	offset, err := strconv.Atoi(*offsetFlag)
	if err != nil {
		return err
	}

	fn, err := astcontext.EnclosingFuncFile(*fileFlag, offset)
	if err != nil {
		return err
	}

	switch *formatFlag {
	case "json", "plain", "xml":
	default:
		return fmt.Errorf("wrong -format value: %q.\n", *formatFlag)
	}

	// Print the result.
	switch *formatFlag {
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

	case "plain":
		fmt.Println(fn)
	}

	return nil
}
