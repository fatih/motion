package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"go/build"
	"go/token"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/fatih/motion/astcontext"
	"github.com/fatih/motion/vim"
)

func main() {
	dirsInit()
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

	if flagDir != nil {
		*flagDir = findDir(*flagDir)
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

func findDir(arg string) string {
	wd, err := os.Getwd()
	if err != nil {
		return arg
	}
	if isDotSlash(arg) {
		arg = filepath.Join(wd, arg)
	}
	if filepath.IsAbs(arg) {
		return arg
	} else {
		pkg, importErr := build.Import(arg, wd, build.FindOnly)
		if importErr == nil {
			return pkg.Dir
		}
	}
	for {
		path, ok := findNextPackage(arg)
		if !ok {
			break
		}
		if _, err = build.ImportDir(path, build.FindOnly); err == nil {
			return path
		}
	}

	return arg
}

// dotPaths lists all the dotted paths legal on Unix-like and
// Windows-like file systems. We check them all, as the chance
// of error is minute and even on Windows people will use ./
// sometimes.
var dotPaths = []string{
	`./`,
	`../`,
	`.\`,
	`..\`,
}

// isDotSlash reports whether the path begins with a reference
// to the local . or .. directory.
func isDotSlash(arg string) bool {
	if arg == "." || arg == ".." {
		return true
	}
	for _, dotPath := range dotPaths {
		if strings.HasPrefix(arg, dotPath) {
			return true
		}
	}
	return false
}

// findNextPackage returns the next full file name path that matches the
// (perhaps partial) package path pkg. The boolean reports if any match was found.
func findNextPackage(pkg string) (string, bool) {
	if filepath.IsAbs(pkg) {
		return "", false
	}
	if pkg == "" || token.IsExported(pkg) { // Upper case symbol cannot be a package name.
		return "", false
	}
	pkg = path.Clean(pkg)
	pkgSuffix := "/" + pkg
	for {
		d, ok := dirs.Next()
		if !ok {
			return "", false
		}
		if d.importPath == pkg || strings.HasSuffix(d.importPath, pkgSuffix) {
			return d.dir, true
		}
	}
}

var buildCtx = build.Default

// splitGopath splits $GOPATH into a list of roots.
func splitGopath() []string {
	return filepath.SplitList(buildCtx.GOPATH)
}
