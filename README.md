# motion

Motion is a tool that was designed to work with editors. It is providing
contextual information for a given offset. Editors can use these informations
to implement navigation, text editing, etc... that are specific to a Go source
code. 

It's optimized and created to work with
[vim-go](https://github.com/fatih/vim-go), but it's designed to work with any
editor.  It's currently work in progress and open to change.

# Install

```bash
go get github.com/fatih/motion
```

# Usage

`motion` is meant to be run via the editor. Currently it has three modes you can use:

* `enclosing`: returns information about the enclosing function for a given offset
* `next`: returns the next function information for a given offset
* `prev`: returns the previous function information for a given offset

A `function information` is currently the following type definition (defined as
`astcontext.Func`):

```
type Func struct {
	FuncPos *Position `json:"func" vim:"func"`     // position of the "func" keyword
	Lbrace  *Position `json:"lbrace" vim:"lbrace"` // position of "{"
	Rbrace  *Position `json:"rbrace" vim:"rbrace"` // position of "}"

	// position of the doc comment, only for *ast.FuncDecl
	Doc *Position `json:"doc,omitempty" vim:"doc,omitempty"`

	node ast.Node // either *ast.FuncDecl or *ast.FuncLit
}
```

`motion` can output the information currently in three formats: `plain`, `json`
and `vim`. 


An example exeuction for the `enclosing` mode and output in `json` format is:

```
$ motion -file testdata/foo.go -offset 160 -mode enclosing --format json
{
	"func": {
		"offset": 151,
		"line": 15,
		"col": 1
	},
	"lbrace": {
		"offset": 178,
		"line": 15,
		"col": 28
	},
	"rbrace": {
		"offset": 202,
		"line": 17,
		"col": 1
	}
}
```

To inluce the doc comments for function declarations include the `--parse-comments` flag:

$ motion -file testdata/foo.go -offset 160 -mode enclosing --format json --parse-comments
{
	"func": {
		"offset": 151,
		"line": 15,
		"col": 1
	},
	"lbrace": {
		"offset": 178,
		"line": 15,
		"col": 28
	},
	"rbrace": {
		"offset": 202,
		"line": 17,
		"col": 1
	},
	"doc": {
		"offset": 126,
		"line": 14,
		"col": 1
	}
}
```

For example the same query, but with mode `-mode next` returns a different
result. Instead it returns the next function inside the source code:

```
$ motion -file testdata/foo.go -offset 160 -mode next --format json
{
	"func": {
		"offset": 302,
		"line": 21,
		"col": 1
	},
	"lbrace": {
		"offset": 319,
		"line": 21,
		"col": 18
	},
	"rbrace": {
		"offset": 382,
		"line": 29,
		"col": 1
	}
}
```

If there are not functions available for any mode, it returns an error in the
specified format:

```
motion -file testdata/foo.go -offset 330 -mode next --format json
{
	"err": "no functions found"
}
```
