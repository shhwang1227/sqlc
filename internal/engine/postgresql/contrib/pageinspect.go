// Code generated by sqlc-pg-gen. DO NOT EDIT.

package contrib

import (
	"github.com/xiazemin/sqlc/internal/sql/ast"
	"github.com/xiazemin/sqlc/internal/sql/catalog"
)

func Pageinspect() *catalog.Schema {
	s := &catalog.Schema{Name: "pg_catalog"}
	s.Funcs = []*catalog.Function{
		{
			Name: "brin_page_type",
			Args: []*catalog.Argument{
				{
					Name: "page",
					Type: &ast.TypeName{Name: "bytea"},
				},
			},
			ReturnType: &ast.TypeName{Name: "text"},
		},
		{
			Name: "fsm_page_contents",
			Args: []*catalog.Argument{
				{
					Name: "page",
					Type: &ast.TypeName{Name: "bytea"},
				},
			},
			ReturnType: &ast.TypeName{Name: "text"},
		},
		{
			Name: "get_raw_page",
			Args: []*catalog.Argument{
				{
					Type: &ast.TypeName{Name: "text"},
				},
				{
					Type: &ast.TypeName{Name: "integer"},
				},
			},
			ReturnType: &ast.TypeName{Name: "bytea"},
		},
		{
			Name: "get_raw_page",
			Args: []*catalog.Argument{
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
			ReturnType: &ast.TypeName{Name: "bytea"},
		},
		{
			Name: "hash_page_type",
			Args: []*catalog.Argument{
				{
					Name: "page",
					Type: &ast.TypeName{Name: "bytea"},
				},
			},
			ReturnType: &ast.TypeName{Name: "text"},
		},
		{
			Name: "page_checksum",
			Args: []*catalog.Argument{
				{
					Name: "page",
					Type: &ast.TypeName{Name: "bytea"},
				},
				{
					Name: "blkno",
					Type: &ast.TypeName{Name: "integer"},
				},
			},
			ReturnType: &ast.TypeName{Name: "smallint"},
		},
		{
			Name: "tuple_data_split",
			Args: []*catalog.Argument{
				{
					Name: "rel_oid",
					Type: &ast.TypeName{Name: "oid"},
				},
				{
					Name: "t_data",
					Type: &ast.TypeName{Name: "bytea"},
				},
				{
					Name: "t_infomask",
					Type: &ast.TypeName{Name: "integer"},
				},
				{
					Name: "t_infomask2",
					Type: &ast.TypeName{Name: "integer"},
				},
				{
					Name: "t_bits",
					Type: &ast.TypeName{Name: "text"},
				},
				{
					Name: "do_detoast",
					Type: &ast.TypeName{Name: "boolean"},
				},
			},
			ReturnType: &ast.TypeName{Name: "bytea[]"},
		},
		{
			Name: "tuple_data_split",
			Args: []*catalog.Argument{
				{
					Name: "rel_oid",
					Type: &ast.TypeName{Name: "oid"},
				},
				{
					Name: "t_data",
					Type: &ast.TypeName{Name: "bytea"},
				},
				{
					Name: "t_infomask",
					Type: &ast.TypeName{Name: "integer"},
				},
				{
					Name: "t_infomask2",
					Type: &ast.TypeName{Name: "integer"},
				},
				{
					Name: "t_bits",
					Type: &ast.TypeName{Name: "text"},
				},
			},
			ReturnType: &ast.TypeName{Name: "bytea[]"},
		},
	}
	return s
}
