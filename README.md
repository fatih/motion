# Motion [![Build Status](http://img.shields.io/travis/fatih/motion.svg?style=flat-square)](https://travis-ci.org/fatih/motion)

Motion is a tool that was designed to work with editors. It is providing
contextual information for a given offset(option) from a file or directory of
files.  Editors can use these informations to implement navigation, text
editing, etc... that are specific to a Go source code.

It's optimized and created to work with
[vim-go](https://github.com/fatih/vim-go), but it's designed to work with any
editor.  It's currently work in progress and open to change.

# Install

```bash
go get github.com/fatih/motion
```

# Usage

`motion` is meant to be run via the editor. Currently it has the following
modes you can use:

* `decls`: returns a list of declarations based on the `-include` flag
* `enclosing`: returns information about the enclosing function for a given
  offset
* `next`: returns the next function information for a given offset
* `prev`: returns the previous function information for a given offset
* `comment`: returns information about the a comment block (if any).

A `function information` is currently the following type definition (defined as
`astcontext.Func`):

```
type Func struct {
	// Signature of the function
	Signature *FuncSignature `json:"sig" vim:"sig"`

	// position of the "func" keyword
	FuncPos *Position `json:"func" vim:"func"`
	Lbrace  *Position `json:"lbrace" vim:"lbrace"` // position of "{"
	Rbrace  *Position `json:"rbrace" vim:"rbrace"` // position of "}"

	// position of the doc comment, only for *ast.FuncDecl
	Doc *Position `json:"doc,omitempty" vim:"doc,omitempty"`

	node ast.Node // either *ast.FuncDecl or *ast.FuncLit
}
```

`motion` can output the information currently in formats: `json` and `vim`.

An example execution for the `enclosing` mode and output in `json` format is:

```
$ motion -file testdata/main.go -offset 180 -mode enclosing --format json
{
	"mode": "enclosing",
	"func": {
		"sig": {
			"full": "func Bar() (string, error)",
			"recv": "",
			"name": "Bar",
			"in": "",
			"out": "string, error"
		},
		"func": {
			"filename": "testdata/main.go",
			"offset": 174,
			"line": 15,
			"col": 1
		},
		"lbrace": {
			"filename": "testdata/main.go",
			"offset": 201,
			"line": 15,
			"col": 28
		},
		"rbrace": {
			"filename": "testdata/main.go",
			"offset": 225,
			"line": 17,
			"col": 1
		}
	}
}
```

To include the doc comments for function declarations include the
`--parse-comments` flag:

```
$ motion -file testdata/main.go -offset 180 -mode enclosing --format json --parse-comments
{
	"mode": "enclosing",
	"func": {
		"sig": {
			"full": "func Bar() (string, error)",
			"recv": "",
			"name": "Bar",
			"in": "",
			"out": "string, error"
		},
		"func": {
			"filename": "testdata/main.go",
			"offset": 174,
			"line": 15,
			"col": 1
		},
		"lbrace": {
			"filename": "testdata/main.go",
			"offset": 201,
			"line": 15,
			"col": 28
		},
		"rbrace": {
			"filename": "testdata/main.go",
			"offset": 225,
			"line": 17,
			"col": 1
		},
		"doc": {
			"filename": "testdata/main.go",
			"offset": 134,
			"line": 14,
			"col": 1
		}
	}
}
```

For example the same query, but with mode `-mode next` returns a different
result. Instead it returns the next function inside the source code:

```
$ motion -file testdata/main.go -offset 180 -mode next --format json
{
	"mode": "next",
	"func": {
		"sig": {
			"full": "func example() error",
			"recv": "",
			"name": "example",
			"in": "",
			"out": "error"
		},
		"func": {
			"filename": "testdata/main.go",
			"offset": 318,
			"line": 21,
			"col": 1
		},
		"lbrace": {
			"filename": "testdata/main.go",
			"offset": 339,
			"line": 21,
			"col": 22
		},
		"rbrace": {
			"filename": "testdata/main.go",
			"offset": 402,
			"line": 29,
			"col": 1
		}
	}
}
```

If there are not functions available for any mode, it returns an error in the
specified format:

```
$ motion -file testdata/main.go -offset 330 -mode next --format json
{
	"err": "no functions found"
}
```

For the mode `decls`, we pass the file (you can also pass a directory)
and instruct to only include function declarations with the `-include func`
flag:
```
$ motion -file testdata/main.go -mode decls -include func
{
	"mode": "decls",
	"decls": [
		{
			"keyword": "func",
			"ident": "main",
			"full": "func main()",
			"filename": "testdata/main.go",
			"line": 9,
			"col": 1
		},
		{
			"keyword": "func",
			"ident": "Bar",
			"full": "func Bar() (string, error)",
			"filename": "testdata/main.go",
			"line": 15,
			"col": 1
		},
		{
			"keyword": "func",
			"ident": "example",
			"full": "func example() error",
			"filename": "testdata/main.go",
			"line": 21,
			"col": 1
		}
	]
}
```

In the `comment` mode it will try to get information about the comment block for
a given offset:
```
$ motion -mode comment -file ./vim/vim.go -offset 3
{
        "mode": "comment",
        "comment": {
                "startLine": 1,
                "startCol": 1,
                "endLine": 3,
                "endCol": 50
        }
}
```
