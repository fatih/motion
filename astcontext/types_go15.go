// +build go1.5

package astcontext

import (
	"bytes"
	"go/ast"
	"go/types"
)

func writeExpr(buf *bytes.Buffer, x ast.Expr) {
	types.WriteExpr(buf, x)
}
