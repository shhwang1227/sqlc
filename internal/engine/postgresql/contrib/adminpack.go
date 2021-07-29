// Code generated by sqlc-pg-gen. DO NOT EDIT.

package contrib

import (
	"github.com/shhwang1227/sqlc/internal/sql/ast"
	"github.com/shhwang1227/sqlc/internal/sql/catalog"
)

func Adminpack() *catalog.Schema {
	s := &catalog.Schema{Name: "pg_catalog"}
	s.Funcs = []*catalog.Function{
		{
			Name: "pg_file_rename",
			Args: []*catalog.Argument{
				{
					Type: &ast.TypeName{Name: "text"},
				},
				{
					Type: &ast.TypeName{Name: "text"},
				},
				{
					Type: &ast.TypeName{Name: "text"},
				},
			},
			ReturnType: &ast.TypeName{Name: "boolean"},
		},
		{
			Name: "pg_file_rename",
			Args: []*catalog.Argument{
				{
					Type: &ast.TypeName{Name: "text"},
				},
				{
					Type: &ast.TypeName{Name: "text"},
				},
			},
			ReturnType: &ast.TypeName{Name: "boolean"},
		},
		{
			Name: "pg_file_unlink",
			Args: []*catalog.Argument{
				{
					Type: &ast.TypeName{Name: "text"},
				},
			},
			ReturnType: &ast.TypeName{Name: "boolean"},
		},
		{
			Name: "pg_file_write",
			Args: []*catalog.Argument{
				{
					Type: &ast.TypeName{Name: "text"},
				},
				{
					Type: &ast.TypeName{Name: "text"},
				},
				{
					Type: &ast.TypeName{Name: "boolean"},
				},
			},
			ReturnType: &ast.TypeName{Name: "bigint"},
		},
		{
			Name:       "pg_logdir_ls",
			Args:       []*catalog.Argument{},
			ReturnType: &ast.TypeName{Name: "record"},
		},
	}
	return s
}
