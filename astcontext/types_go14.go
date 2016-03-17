// +build go1.4,!go1.5

package astcontext

import (
	"bytes"
	"go/ast"

	"golang.org/x/tools/go/types"
)

func writeExpr(buf *bytes.Buffer, x ast.Expr) {
	types.WriteExpr(buf, x)
}
