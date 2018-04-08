package astcontext

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestComment(t *testing.T) {
	var src = `package main

// Hello

/* foo
bar */

// ..
// ..
`
	opts := &ParserOptions{Src: []byte(src), Comments: true}
	parser, err := NewParser(opts)
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		offset  int
		want    Comment
		wantErr string
	}{
		{4, Comment{}, "no comment block"},
		{9000, Comment{}, "no comment block"},
		{18, Comment{3, 1, 3, 9}, ""},
		{24, Comment{5, 1, 6, 7}, ""},
		{39, Comment{8, 1, 9, 6}, ""},
	}

	for _, tc := range cases {
		t.Run(fmt.Sprintf("%v", tc.offset), func(t *testing.T) {
			out, err := parser.Run(&Query{Mode: "comment", Offset: tc.offset})
			if !errorContains(err, tc.wantErr) {
				t.Fatalf("wrong error:\nwant: %v\ngot:  %v", tc.wantErr, err)
			}

			if err != nil {
				return
			}

			if !reflect.DeepEqual(out.Comment, tc.want) {
				t.Fatalf("wrong output:\nwant: %v\ngot:  %v", tc.want, out.Comment)
			}
		})
	}
}

// errorContains checks if the error message in out contains the text in
// want.
//
// This is safe when out is nil. Use an empty string for want if you want to
// test that err is nil.
func errorContains(out error, want string) bool {
	if out == nil {
		return want == ""
	}
	if want == "" {
		return false
	}
	return strings.Contains(out.Error(), want)
}
