package named

import (
	"github.com/shhwang1227/sqlc/internal/sql/ast"
	"github.com/shhwang1227/sqlc/internal/sql/astutils"
)

func IsParamFunc(node ast.Node) bool {
	call, ok := node.(*ast.FuncCall)
	if !ok {
		return false
	}
	if call.Func == nil {
		return false
	}
	return call.Func.Schema == "sqlc" && call.Func.Name == "arg"
}

func IsParamSign(node ast.Node) bool {
	expr, ok := node.(*ast.A_Expr)
	return ok && astutils.Join(expr.Name, ".") == "@"
}

func IsIn(node ast.Node) bool {
	if _, ok := node.(*ast.In); ok {
		return true
	}
	return false
}
