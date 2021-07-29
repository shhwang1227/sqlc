// Code generated by sqlc-pg-gen. DO NOT EDIT.

package contrib

import (
	"github.com/shhwang1227/sqlc/internal/sql/ast"
	"github.com/shhwang1227/sqlc/internal/sql/catalog"
)

func Tablefunc() *catalog.Schema {
	s := &catalog.Schema{Name: "pg_catalog"}
	s.Funcs = []*catalog.Function{
		{
			Name: "connectby",
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
				{
					Type: &ast.TypeName{Name: "text"},
				},
				{
					Type: &ast.TypeName{Name: "integer"},
				},
			},
			ReturnType: &ast.TypeName{Name: "record"},
		},
		{
			Name: "connectby",
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
				{
					Type: &ast.TypeName{Name: "text"},
				},
				{
					Type: &ast.TypeName{Name: "integer"},
				},
				{
					Type: &ast.TypeName{Name: "text"},
				},
			},
			ReturnType: &ast.TypeName{Name: "record"},
		},
		{
			Name: "connectby",
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
				{
					Type: &ast.TypeName{Name: "text"},
				},
				{
					Type: &ast.TypeName{Name: "text"},
				},
				{
					Type: &ast.TypeName{Name: "integer"},
				},
			},
			ReturnType: &ast.TypeName{Name: "record"},
		},
		{
			Name: "connectby",
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
				{
					Type: &ast.TypeName{Name: "text"},
				},
				{
					Type: &ast.TypeName{Name: "text"},
				},
				{
					Type: &ast.TypeName{Name: "integer"},
				},
				{
					Type: &ast.TypeName{Name: "text"},
				},
			},
			ReturnType: &ast.TypeName{Name: "record"},
		},
		{
			Name: "crosstab",
			Args: []*catalog.Argument{
				{
					Type: &ast.TypeName{Name: "text"},
				},
				{
					Type: &ast.TypeName{Name: "integer"},
				},
			},
			ReturnType: &ast.TypeName{Name: "record"},
		},
		{
			Name: "crosstab",
			Args: []*catalog.Argument{
				{
					Type: &ast.TypeName{Name: "text"},
				},
			},
			ReturnType: &ast.TypeName{Name: "record"},
		},
		{
			Name: "crosstab",
			Args: []*catalog.Argument{
				{
					Type: &ast.TypeName{Name: "text"},
				},
				{
					Type: &ast.TypeName{Name: "text"},
				},
			},
			ReturnType: &ast.TypeName{Name: "record"},
		},
		{
			Name: "crosstab2",
			Args: []*catalog.Argument{
				{
					Type: &ast.TypeName{Name: "text"},
				},
			},
			ReturnType: &ast.TypeName{Name: "tablefunc_crosstab_2"},
		},
		{
			Name: "crosstab3",
			Args: []*catalog.Argument{
				{
					Type: &ast.TypeName{Name: "text"},
				},
			},
			ReturnType: &ast.TypeName{Name: "tablefunc_crosstab_3"},
		},
		{
			Name: "crosstab4",
			Args: []*catalog.Argument{
				{
					Type: &ast.TypeName{Name: "text"},
				},
			},
			ReturnType: &ast.TypeName{Name: "tablefunc_crosstab_4"},
		},
		{
			Name: "normal_rand",
			Args: []*catalog.Argument{
				{
					Type: &ast.TypeName{Name: "integer"},
				},
				{
					Type: &ast.TypeName{Name: "double precision"},
				},
				{
					Type: &ast.TypeName{Name: "double precision"},
				},
			},
			ReturnType: &ast.TypeName{Name: "double precision"},
		},
	}
	return s
}
