package compiler

import (
	"errors"
	"fmt"

	"github.com/shhwang1227/sqlc/internal/sql/ast"
	"github.com/shhwang1227/sqlc/internal/sql/astutils"
	"github.com/shhwang1227/sqlc/internal/sql/catalog"
	"github.com/shhwang1227/sqlc/internal/sql/lang"
	"github.com/shhwang1227/sqlc/internal/sql/sqlerr"
	"github.com/shhwang1227/sqlc/internal/util"
)

func hasStarRef(cf *ast.ColumnRef) bool {
	for _, item := range cf.Fields.Items {
		if _, ok := item.(*ast.A_Star); ok {
			return true
		}
	}
	return false
}

// Compute the output columns for a statement.
//
// Return an error if column references are ambiguous
// Return an error if column references don't exist
func outputColumns(qc *QueryCatalog, node ast.Node) ([]*Column, error) {
	tables, err := sourceTables(qc, node)
	if err != nil {
		return nil, err
	}

	var targets *ast.List
	switch n := node.(type) {
	case *ast.DeleteStmt:
		targets = n.ReturningList
	case *ast.InsertStmt:
		targets = n.ReturningList
	case *ast.SelectStmt:
		targets = n.TargetList
	case *ast.TruncateStmt:
		targets = &ast.List{}
	case *ast.UpdateStmt:
		targets = n.ReturningList
	default:
		return nil, fmt.Errorf("outputColumns: unsupported node type: %T", n)
	}

	var cols []*Column

	for _, target := range targets.Items {
		res, ok := target.(*ast.ResTarget)
		if !ok {
			continue
		}
		switch n := res.Val.(type) {

		case *ast.A_Expr:
			name := ""
			if res.Name != nil {
				name = *res.Name
			}
			switch {
			case lang.IsComparisonOperator(astutils.Join(n.Name, "")):
				// TODO: Generate a name for these operations
				cols = append(cols, &Column{Name: name, DataType: "bool", NotNull: true})
			case lang.IsMathematicalOperator(astutils.Join(n.Name, "")):
				cols = append(cols, &Column{Name: name, DataType: "int", NotNull: true})
			default:
				cols = append(cols, &Column{Name: name, DataType: "any", NotNull: false})
			}

		case *ast.CaseExpr:
			name := ""
			if res.Name != nil {
				name = *res.Name
			}
			// TODO: The TypeCase code has been copied from below. Instead, we need a recurse function to get the type of a node.
			if tc, ok := n.Defresult.(*ast.TypeCast); ok {
				if tc.TypeName == nil {
					return nil, errors.New("no type name type cast")
				}
				name := ""
				if ref, ok := tc.Arg.(*ast.ColumnRef); ok {
					name = astutils.Join(ref.Fields, "_")
				}
				if res.Name != nil {
					name = *res.Name
				}
				// TODO Validate column names
				col := toColumn(tc.TypeName)
				col.Name = name
				cols = append(cols, col)
			} else {
				cols = append(cols, &Column{Name: name, DataType: "any", NotNull: false})
			}

		case *ast.CoalesceExpr:
			var found bool
			for _, arg := range n.Args.Items {
				if found {
					continue
				}
				if ref, ok := arg.(*ast.ColumnRef); ok {
					columns, err := outputColumnRefs(res, tables, ref)
					if err != nil {
						return nil, err
					}
					for _, c := range columns {
						found = true
						c.NotNull = true
						cols = append(cols, c)
					}
				}
			}
			if !found {
				cols = append(cols, &Column{Name: "coalesce", DataType: "any", NotNull: false})
			}

		case *ast.ColumnRef:
			if hasStarRef(n) {
				// TODO: This code is copied in func expand()
				for _, t := range tables {
					scope := astutils.Join(n.Fields, ".")
					if scope != "" && scope != t.Rel.Name {
						continue
					}
					for _, c := range t.Columns {
						cname := c.Name
						if res.Name != nil {
							cname = *res.Name
						}
						cols = append(cols, &Column{
							Name:     cname,
							Type:     c.Type,
							Scope:    scope,
							Table:    c.Table,
							DataType: c.DataType,
							NotNull:  c.NotNull,
							IsArray:  c.IsArray,
						})
					}
				}
				continue
			}

			columns, err := outputColumnRefs(res, tables, n)
			if err != nil {
				return nil, err
			}
			cols = append(cols, columns...)

		case *ast.FuncCall:
			//这里解析函数调用
			util.Xiazeminlog("ast.FuncCall", n, false)
			util.Xiazeminlog("ast.FuncCall tables ", tables, false)
			/*
				{
					"Func": {
						"Catalog": "",
						"Schema": "",
						"Name": "ifnull"
					},
					"Funcname": {
						"Items": [
							{
								"Str": "ifnull"
							}
						]
					},
					"Args": {
						"Items": [
							{
								"Func": {
									"Catalog": "",
									"Schema": "",
									"Name": "sum"
								},
								"Funcname": {
									"Items": [
										{
											"Str": "sum"
										}
									]
								},
								"Args": {
									"Items": [
										{
											"Name": "",
											"Fields": {
												"Items": [
													{
														"Str": "size"
													}
												]
											},
											"Location": 0,
											"TableName": ""
										}
									]
								},
								"AggOrder": {
									"Items": null
								},
								"AggFilter": null,
								"AggWithinGroup": false,
								"AggStar": false,
								"AggDistinct": false,
								"FuncVariadic": false,
								"Over": null,
								"Location": 0
							},
							{
								"Val": {
									"Str": ""
								},
								"Location": 0
							}
						]
					},
					"AggOrder": null,
					"AggFilter": null,
					"AggWithinGroup": false,
					"AggStar": false,
					"AggDistinct": false,
					"FuncVariadic": false,
					"Over": null,
					"Location": 981
				}
			*/
			rel := n.Func
			name := rel.Name //这里直接用方法的名字也行，和sql保持一致
			if res.Name != nil {
				name = *res.Name
			}
			var tablse1 []*catalog.Table
			for _, tab := range tables {
				var col []*catalog.Column
				for _, c := range tab.Columns {
					col = append(col, &catalog.Column{
						Name: c.Name,
						Type: ast.TypeName{
							Name: c.DataType,
						},
						IsNotNull: c.NotNull,
						IsArray:   c.IsArray,
						Comment:   c.Comment,
						Length:    c.Length,
					})
				}

				tablse1 = append(tablse1, &catalog.Table{
					Rel:     tab.Rel,
					Columns: col,
				})
			}

			fun, notNull, err := qc.catalog.ResolveFuncCall(n, tablse1)
			util.Xiazeminlog("ResolveFuncCall", fun, false)
			util.Xiazeminlog("ResolveFuncCall", notNull, false)
			/*
				{
					"Name": "IFNULL",
					"Args": [
						{
							"Name": "",
							"Type": {
								"Catalog": "",
								"Schema": "",
								"Name": "any",
								"Names": null,
								"TypeOid": 0,
								"Setof": false,
								"PctType": false,
								"Typmods": null,
								"Typemod": 0,
								"ArrayBounds": null,
								"Location": 0
							},
							"HasDefault": false,
							"Mode": 0
						},
						{
							"Name": "",
							"Type": {
								"Catalog": "",
								"Schema": "",
								"Name": "bigint",
								"Names": null,
								"TypeOid": 0,
								"Setof": false,
								"PctType": false,
								"Typmods": null,
								"Typemod": 0,
								"ArrayBounds": null,
								"Location": 0
							},
							"HasDefault": false,
							"Mode": 0
						}
					],
					"ReturnType": {
						"Catalog": "",
						"Schema": "",
						"Name": "bigint",
						"Names": null,
						"TypeOid": 0,
						"Setof": false,
						"PctType": false,
						"Typmods": null,
						"Typemod": 0,
						"ArrayBounds": null,
						"Location": 0
					},
					"Comment": "",
					"Desc": ""
				}
			*/
			util.Xiazeminlog("\n \n ResolveFuncCall name ", []string{name, rel.Name}, false)
			if err == nil {
				//这里的NotNull 不能直接为true，否则会报0 sql: Scan error on column index 0, name "sum(size)": converting NULL to int64 is unsupported
				cols = append(cols, &Column{Name: name, DataType: dataType(fun.ReturnType), NotNull: notNull})
			} else {
				cols = append(cols, &Column{Name: name, DataType: "any"})
			}

		case *ast.SubLink:
			name := "exists"
			if res.Name != nil {
				name = *res.Name
			}
			switch n.SubLinkType {
			case ast.EXISTS_SUBLINK:
				cols = append(cols, &Column{Name: name, DataType: "bool", NotNull: true})
			default:
				cols = append(cols, &Column{Name: name, DataType: "any", NotNull: false})
			}

		case *ast.TypeCast:
			if n.TypeName == nil {
				return nil, errors.New("no type name type cast")
			}
			name := ""
			if ref, ok := n.Arg.(*ast.ColumnRef); ok {
				name = astutils.Join(ref.Fields, "_")
			}
			if res.Name != nil {
				name = *res.Name
			}
			// TODO Validate column names
			col := toColumn(n.TypeName)
			col.Name = name
			cols = append(cols, col)

		default:
			name := ""
			if res.Name != nil {
				name = *res.Name
			}
			cols = append(cols, &Column{Name: name, DataType: "any", NotNull: false})

		}
	}
	return cols, nil
}

// Compute the output columns for a statement.
//
// Return an error if column references are ambiguous
// Return an error if column references don't exist
// Return an error if a table is referenced twice
// Return an error if an unknown column is referenced
func sourceTables(qc *QueryCatalog, node ast.Node) ([]*Table, error) {
	var list *ast.List
	switch n := node.(type) {
	case *ast.DeleteStmt:
		list = &ast.List{
			Items: []ast.Node{n.Relation},
		}
	case *ast.InsertStmt:
		list = &ast.List{
			Items: []ast.Node{n.Relation},
		}
	case *ast.SelectStmt:
		list = astutils.Search(n.FromClause, func(node ast.Node) bool {
			switch node.(type) {
			case *ast.RangeVar, *ast.RangeSubselect:
				return true
			default:
				return false
			}
		})
	case *ast.TruncateStmt:
		list = astutils.Search(n.Relations, func(node ast.Node) bool {
			_, ok := node.(*ast.RangeVar)
			return ok
		})
	case *ast.UpdateStmt:
		list = &ast.List{
			Items: append(n.FromClause.Items, n.Relation),
		}
	default:
		return nil, fmt.Errorf("sourceTables: unsupported node type: %T", n)
	}

	var tables []*Table
	for _, item := range list.Items {
		switch n := item.(type) {
		case *ast.RangeSubselect:
			cols, err := outputColumns(qc, n.Subquery)
			if err != nil {
				return nil, err
			}
			tables = append(tables, &Table{
				Rel: &ast.TableName{
					Name: *n.Alias.Aliasname,
				},
				Columns: cols,
			})

		case *ast.RangeVar:
			fqn, err := ParseTableName(n)
			if err != nil {
				return nil, err
			}
			table, cerr := qc.GetTable(fqn)
			if cerr != nil {
				// TODO: Update error location
				// cerr.Location = n.Location
				// return nil, *cerr
				return nil, cerr
			}
			if n.Alias != nil {
				table.Rel = &ast.TableName{
					Catalog: table.Rel.Catalog,
					Schema:  table.Rel.Schema,
					Name:    *n.Alias.Aliasname,
				}
			}
			tables = append(tables, table)
		default:
			return nil, fmt.Errorf("sourceTable: unsupported list item type: %T", n)
		}
	}
	return tables, nil
}

func outputColumnRefs(res *ast.ResTarget, tables []*Table, node *ast.ColumnRef) ([]*Column, error) {
	parts := stringSlice(node.Fields)
	var name, alias string
	switch {
	case len(parts) == 1:
		name = parts[0]
	case len(parts) == 2:
		alias = parts[0]
		name = parts[1]
	default:
		return nil, fmt.Errorf("unknown number of fields: %d", len(parts))
	}
	var cols []*Column
	var found int
	for _, t := range tables {
		if alias != "" && t.Rel.Name != alias {
			continue
		}
		for _, c := range t.Columns {
			if c.Name == name {
				found += 1
				cname := c.Name
				if res.Name != nil {
					cname = *res.Name
				}
				cols = append(cols, &Column{
					Name:     cname,
					Type:     c.Type,
					Table:    c.Table,
					DataType: c.DataType,
					NotNull:  c.NotNull,
					IsArray:  c.IsArray,
				})
			}
		}
	}
	if found == 0 {
		return nil, &sqlerr.Error{
			Code:     "42703",
			Message:  fmt.Sprintf("column \"%s\" does not exist", name),
			Location: res.Location,
		}
	}
	if found > 1 {
		return nil, &sqlerr.Error{
			Code:     "42703",
			Message:  fmt.Sprintf("OUT column reference \"%s\" is ambiguous", name),
			Location: res.Location,
		}
	}
	return cols, nil
}
