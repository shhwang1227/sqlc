// Code generated by sqlc-pg-gen. DO NOT EDIT.

package contrib

import (
	"github.com/xiazemin/sqlc/internal/sql/ast"
	"github.com/xiazemin/sqlc/internal/sql/catalog"
)

func PostgresFdw() *catalog.Schema {
	s := &catalog.Schema{Name: "pg_catalog"}
	s.Funcs = []*catalog.Function{
		{
			Name:       "postgres_fdw_handler",
			Args:       []*catalog.Argument{},
			ReturnType: &ast.TypeName{Name: "fdw_handler"},
		},
		{
			Name: "postgres_fdw_validator",
			Args: []*catalog.Argument{
				{
					Type: &ast.TypeName{Name: "text[]"},
				},
				{
					Type: &ast.TypeName{Name: "oid"},
				},
			},
			ReturnType: &ast.TypeName{Name: "void"},
		},
	}
	return s
}
