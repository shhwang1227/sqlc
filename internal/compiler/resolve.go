package compiler

import (
	"fmt"
	"strconv"

	"github.com/xiazemin/sqlc/internal/sql/ast"
	"github.com/xiazemin/sqlc/internal/sql/astutils"
	"github.com/xiazemin/sqlc/internal/sql/catalog"
	"github.com/xiazemin/sqlc/internal/sql/sqlerr"
	"github.com/xiazemin/sqlc/internal/util"
)

func dataType(n *ast.TypeName) string {
	if n.Schema != "" {
		return n.Schema + "." + n.Name
	} else {
		return n.Name
	}
}

func resolveCatalogValuesRefs(c *catalog.Catalog, rvs []*ast.RangeVar, args []paramRef, names map[int]string) ([]Parameter, int64, error) {
	aliasMap := map[string]*ast.TableName{}
	// TODO: Deprecate defaultTable
	var defaultTable *ast.TableName
	var tables []*ast.TableName

	parameterName := func(n int, defaultName string) string {
		if n, ok := names[n]; ok {
			return n
		}
		return defaultName
	}

	for _, rv := range rvs {
		if rv.Relname == nil {
			continue
		}
		fqn, err := ParseTableName(rv)
		if err != nil {
			return nil, 0, err
		}
		tables = append(tables, fqn)
		if defaultTable == nil {
			defaultTable = fqn
		}
		if rv.Alias == nil {
			continue
		}
		aliasMap[*rv.Alias.Aliasname] = fqn
	}

	typeMap := map[string]map[string]map[string]*catalog.Column{}
	for _, fqn := range tables {
		table, err := c.GetTable(fqn)
		if err != nil {
			continue
		}
		if _, exists := typeMap[fqn.Schema]; !exists {
			typeMap[fqn.Schema] = map[string]map[string]*catalog.Column{}
		}
		typeMap[fqn.Schema][fqn.Name] = map[string]*catalog.Column{}
		for _, c := range table.Columns {
			cc := c
			typeMap[fqn.Schema][fqn.Name][c.Name] = cc
		}
	}

	var a []Parameter
	util.Xiazeminlog("range args", args, false)
	for _, ref := range args {
		util.Xiazeminlog("args ---", ref, false)
		switch n := ref.parent.(type) {
		case *limitOffset:
			a = append(a, Parameter{
				Number: ref.ref.Number,
				Column: &Column{
					Name:     parameterName(ref.ref.Number, "offset"),
					DataType: "integer",
					NotNull:  true,
				},
			})

		case *limitCount:
			a = append(a, Parameter{
				Number: ref.ref.Number,
				Column: &Column{
					Name:     parameterName(ref.ref.Number, "limit"),
					DataType: "integer",
					NotNull:  true,
				},
			})

		case *ast.A_Expr:
			// TODO: While this works for a wide range of simple expressions,
			// more complicated expressions will cause this logic to fail.
			list := astutils.Search(n.Lexpr, func(node ast.Node) bool {
				_, ok := node.(*ast.ColumnRef)
				return ok
			})

			if len(list.Items) == 0 {
				// TODO: Move this to database-specific engine package
				dataType := "any"
				if astutils.Join(n.Name, ".") == "||" {
					dataType = "string"
				}
				a = append(a, Parameter{
					Number: ref.ref.Number,
					Column: &Column{
						Name:     parameterName(ref.ref.Number, ""),
						DataType: dataType,
					},
				})
				continue
			}

			switch left := list.Items[0].(type) {
			case *ast.ColumnRef:
				items := stringSlice(left.Fields)
				var key, alias string
				switch len(items) {
				case 1:
					key = items[0]
				case 2:
					alias = items[0]
					key = items[1]
				default:
					panic("too many field items: " + strconv.Itoa(len(items)))
				}

				search := tables
				if alias != "" {
					if original, ok := aliasMap[alias]; ok {
						search = []*ast.TableName{original}
					} else {
						for _, fqn := range tables {
							if fqn.Name == alias {
								search = []*ast.TableName{fqn}
							}
						}
					}
				}

				var found int
				for _, table := range search {
					if c, ok := typeMap[table.Schema][table.Name][key]; ok && (left.TableName == "" || left.TableName == table.Name) {
						found += 1
						if ref.name != "" {
							key = ref.name
						}
						a = append(a, Parameter{
							Number: ref.ref.Number,
							Column: &Column{
								Name:     parameterName(ref.ref.Number, key),
								DataType: dataType(&c.Type),
								NotNull:  c.IsNotNull,
								IsArray:  c.IsArray,
								Table:    table,
							},
						})
					}
				}

				if found == 0 {
					return nil, 0, &sqlerr.Error{
						Code:     "42703",
						Message:  fmt.Sprintf("168:column \"%s\" does not exist", key),
						Location: left.Location,
					}
				}
				if found > 1 {
					return nil, 0, &sqlerr.Error{
						Code:     "42703",
						Message:  fmt.Sprintf("RESOLVE MESSAGE column reference \"%s\" is ambiguous", key),
						Location: left.Location,
					}
				}
			}

		case *ast.FuncCall:
			fun, err := c.ResolveFuncCall(n)
			if err != nil {
				// Synthesize a function on the fly to avoid returning with an error
				// for an unknown Postgres function (e.g. defined in an extension)
				var args []*catalog.Argument
				for range n.Args.Items {
					args = append(args, &catalog.Argument{
						Type: &ast.TypeName{Name: "any"},
					})
				}
				fun = &catalog.Function{
					Name:       n.Func.Name,
					Args:       args,
					ReturnType: &ast.TypeName{Name: "any"},
				}
			}
			for i, item := range n.Args.Items {
				funcName := fun.Name
				var argName string
				switch inode := item.(type) {
				case *ast.ParamRef:
					if inode.Number != ref.ref.Number {
						continue
					}
				case *ast.TypeCast:
					pr, ok := inode.Arg.(*ast.ParamRef)
					if !ok {
						continue
					}
					if pr.Number != ref.ref.Number {
						continue
					}
				case *ast.NamedArgExpr:
					pr, ok := inode.Arg.(*ast.ParamRef)
					if !ok {
						continue
					}
					if pr.Number != ref.ref.Number {
						continue
					}
					if inode.Name != nil {
						argName = *inode.Name
					}
				default:
					continue
				}

				if fun.Args == nil {
					defaultName := funcName
					if argName != "" {
						defaultName = argName
					}
					a = append(a, Parameter{
						Number: ref.ref.Number,
						Column: &Column{
							Name:     parameterName(ref.ref.Number, defaultName),
							DataType: "any",
						},
					})
					continue
				}

				var paramName string
				var paramType *ast.TypeName
				if argName == "" {
					paramName = fun.Args[i].Name
					paramType = fun.Args[i].Type
				} else {
					paramName = argName
					for _, arg := range fun.Args {
						if arg.Name == argName {
							paramType = arg.Type
						}
					}
					if paramType == nil {
						panic(fmt.Sprintf("named argument %s has no type", paramName))
					}
				}
				if paramName == "" {
					paramName = funcName
				}

				a = append(a, Parameter{
					Number: ref.ref.Number,
					Column: &Column{
						Name:     parameterName(ref.ref.Number, paramName),
						DataType: dataType(paramType),
						NotNull:  true,
					},
				})
			}

		case *ast.ResTarget:
			if n.Name == nil {
				return nil, 0, fmt.Errorf("*ast.ResTarget has nil name")
			}
			key := *n.Name

			// TODO: Deprecate defaultTable
			schema := defaultTable.Schema
			rel := defaultTable.Name
			if ref.rv != nil {
				fqn, err := ParseTableName(ref.rv)
				if err != nil {
					return nil, 0, err
				}
				schema = fqn.Schema
				rel = fqn.Name
			}
			if c, ok := typeMap[schema][rel][key]; ok {
				a = append(a, Parameter{
					Number: ref.ref.Number,
					Column: &Column{
						Name:     parameterName(ref.ref.Number, key),
						DataType: dataType(&c.Type),
						NotNull:  c.IsNotNull,
						IsArray:  c.IsArray,
						Table:    &ast.TableName{Schema: schema, Name: rel},
					},
				})
				util.Xiazeminlog("ast.ResTarget here", a, false)
			} else {
				return nil, 0, &sqlerr.Error{
					Code:     "42703",
					Message:  fmt.Sprintf("306: column \"%s\" does not exist", key),
					Location: n.Location,
				}
			}

		case *ast.TypeCast:
			if n.TypeName == nil {
				return nil, 0, fmt.Errorf("*ast.TypeCast has nil type name")
			}
			col := toColumn(n.TypeName)
			col.Name = parameterName(ref.ref.Number, col.Name)
			a = append(a, Parameter{
				Number: ref.ref.Number,
				Column: col,
			})

		case *ast.ParamRef:
			a = append(a, Parameter{Number: ref.ref.Number})
		case *ast.In:
			if n == nil || n.List == nil {
				fmt.Println("ast.In is nil")
				continue
			}

			number := 0
			if pr, ok := n.List[0].(*ast.ParamRef); ok {
				number = pr.Number
			}

			left, ok := n.Expr.(*ast.ColumnRef)
			if !ok {
				continue
			}
			util.Xiazeminlog("left", left, false)
			util.Xiazeminlog("n", n, false)
			items := stringSlice(left.Fields)
			var key, alias string
			switch len(items) {
			case 1:
				key = items[0]
			case 2:
				alias = items[0]
				key = items[1]
			default:
				panic("too many field items: " + strconv.Itoa(len(items)))
			}

			var found int
			util.Xiazeminlog("range sel-", n, false)
			if n.Sel == nil {
				search := tables
				if alias != "" {
					if original, ok := aliasMap[alias]; ok {
						search = []*ast.TableName{original}
					} else {
						for _, fqn := range tables {
							if fqn.Name == alias {
								search = []*ast.TableName{fqn}
							}
						}
					}
				}

				util.Xiazeminlog("range table", search, false)
				for _, table := range search {
					if c, ok := typeMap[table.Schema][table.Name][key]; ok && n.TableName == table.Name {
						found += 1
						if ref.name != "" {
							key = ref.name
						}
						a = append(a, Parameter{
							Number: number,
							Column: &Column{
								IsSlice:  true,
								Name:     parameterName(ref.ref.Number, key),
								DataType: dataType(&c.Type),
								NotNull:  c.IsNotNull,
								IsArray:  c.IsArray,
								Table:    table,
							},
						})
					}
				}
			} else {
				fmt.Println("------------------------")
			}

			if found == 0 {
				return nil, 0, &sqlerr.Error{
					Code:     "42703",
					Message:  fmt.Sprintf("396: column \"%s\" does not exist", key),
					Location: left.Location,
				}
			}
			if found > 1 {
				util.Xiazeminlog("typeMap", typeMap, false)
				return nil, 0, &sqlerr.Error{
					Code:     "42703",
					Message:  fmt.Sprintf("in same name column reference \"%s\" is ambiguous", key),
					Location: left.Location,
				}
			}
		default:
			fmt.Printf("unsupported reference type: %T", n)
		}
	}
	return a, 0, nil
}

func resolveCatalogRefs(c *catalog.Catalog, rvs []*ast.RangeVar, args []paramRef, names map[int]string) ([]Parameter, error) {
	aliasMap := map[string]*ast.TableName{}
	// TODO: Deprecate defaultTable
	var defaultTable *ast.TableName
	var tables []*ast.TableName

	parameterName := func(n int, defaultName string) string {
		if n, ok := names[n]; ok {
			return n
		}
		return defaultName
	}

	for _, rv := range rvs {
		if rv.Relname == nil {
			continue
		}
		fqn, err := ParseTableName(rv)
		if err != nil {
			return nil, err
		}
		tables = append(tables, fqn)
		if defaultTable == nil {
			defaultTable = fqn
		}
		if rv.Alias == nil {
			continue
		}
		aliasMap[*rv.Alias.Aliasname] = fqn
	}

	typeMap := map[string]map[string]map[string]*catalog.Column{}
	for _, fqn := range tables {
		table, err := c.GetTable(fqn)
		if err != nil {
			continue
		}
		if _, exists := typeMap[fqn.Schema]; !exists {
			typeMap[fqn.Schema] = map[string]map[string]*catalog.Column{}
		}
		typeMap[fqn.Schema][fqn.Name] = map[string]*catalog.Column{}
		for _, c := range table.Columns {
			cc := c
			typeMap[fqn.Schema][fqn.Name][c.Name] = cc
		}
	}

	var a []Parameter
	util.Xiazeminlog("range args", args, false)
	for _, ref := range args {
		util.Xiazeminlog("args ---", ref, false)
		switch n := ref.parent.(type) {
		case *limitOffset:
			a = append(a, Parameter{
				Number: ref.ref.Number,
				Column: &Column{
					Name:     parameterName(ref.ref.Number, "offset"),
					DataType: "integer",
					NotNull:  true,
				},
			})

		case *limitCount:
			a = append(a, Parameter{
				Number: ref.ref.Number,
				Column: &Column{
					Name:     parameterName(ref.ref.Number, "limit"),
					DataType: "integer",
					NotNull:  true,
				},
			})

		case *ast.A_Expr:
			// TODO: While this works for a wide range of simple expressions,
			// more complicated expressions will cause this logic to fail.
			list := astutils.Search(n.Lexpr, func(node ast.Node) bool {
				_, ok := node.(*ast.ColumnRef)
				return ok
			})

			if len(list.Items) == 0 {
				// TODO: Move this to database-specific engine package
				dataType := "any"
				if astutils.Join(n.Name, ".") == "||" {
					dataType = "string"
				}
				a = append(a, Parameter{
					Number: ref.ref.Number,
					Column: &Column{
						Name:     parameterName(ref.ref.Number, ""),
						DataType: dataType,
					},
				})
				continue
			}

			switch left := list.Items[0].(type) {
			case *ast.ColumnRef:
				items := stringSlice(left.Fields)
				var key, alias string
				switch len(items) {
				case 1:
					key = items[0]
				case 2:
					alias = items[0]
					key = items[1]
				default:
					panic("too many field items: " + strconv.Itoa(len(items)))
				}

				search := tables
				if alias != "" {
					if original, ok := aliasMap[alias]; ok {
						search = []*ast.TableName{original}
					} else {
						for _, fqn := range tables {
							if fqn.Name == alias {
								search = []*ast.TableName{fqn}
							}
						}
					}
				}

				var found int
				for _, table := range search {
					if c, ok := typeMap[table.Schema][table.Name][key]; ok && (left.TableName == "" || left.TableName == table.Name) {
						found += 1
						if ref.name != "" {
							key = ref.name
						}
						a = append(a, Parameter{
							Number: ref.ref.Number,
							Column: &Column{
								Name:     parameterName(ref.ref.Number, key),
								DataType: dataType(&c.Type),
								NotNull:  c.IsNotNull,
								IsArray:  c.IsArray,
								Table:    table,
							},
						})
					}
				}

				if found == 0 {
					return nil, &sqlerr.Error{
						Code:     "42703",
						Message:  fmt.Sprintf("168:column \"%s\" does not exist", key),
						Location: left.Location,
					}
				}
				if found > 1 {
					return nil, &sqlerr.Error{
						Code:     "42703",
						Message:  fmt.Sprintf("RESOLVE MESSAGE column reference \"%s\" is ambiguous", key),
						Location: left.Location,
					}
				}
			}

		case *ast.FuncCall:
			fun, err := c.ResolveFuncCall(n)
			if err != nil {
				// Synthesize a function on the fly to avoid returning with an error
				// for an unknown Postgres function (e.g. defined in an extension)
				var args []*catalog.Argument
				for range n.Args.Items {
					args = append(args, &catalog.Argument{
						Type: &ast.TypeName{Name: "any"},
					})
				}
				fun = &catalog.Function{
					Name:       n.Func.Name,
					Args:       args,
					ReturnType: &ast.TypeName{Name: "any"},
				}
			}
			for i, item := range n.Args.Items {
				funcName := fun.Name
				var argName string
				switch inode := item.(type) {
				case *ast.ParamRef:
					if inode.Number != ref.ref.Number {
						continue
					}
				case *ast.TypeCast:
					pr, ok := inode.Arg.(*ast.ParamRef)
					if !ok {
						continue
					}
					if pr.Number != ref.ref.Number {
						continue
					}
				case *ast.NamedArgExpr:
					pr, ok := inode.Arg.(*ast.ParamRef)
					if !ok {
						continue
					}
					if pr.Number != ref.ref.Number {
						continue
					}
					if inode.Name != nil {
						argName = *inode.Name
					}
				default:
					continue
				}

				if fun.Args == nil {
					defaultName := funcName
					if argName != "" {
						defaultName = argName
					}
					a = append(a, Parameter{
						Number: ref.ref.Number,
						Column: &Column{
							Name:     parameterName(ref.ref.Number, defaultName),
							DataType: "any",
						},
					})
					continue
				}

				var paramName string
				var paramType *ast.TypeName
				if argName == "" {
					paramName = fun.Args[i].Name
					paramType = fun.Args[i].Type
				} else {
					paramName = argName
					for _, arg := range fun.Args {
						if arg.Name == argName {
							paramType = arg.Type
						}
					}
					if paramType == nil {
						panic(fmt.Sprintf("named argument %s has no type", paramName))
					}
				}
				if paramName == "" {
					paramName = funcName
				}

				a = append(a, Parameter{
					Number: ref.ref.Number,
					Column: &Column{
						Name:     parameterName(ref.ref.Number, paramName),
						DataType: dataType(paramType),
						NotNull:  true,
					},
				})
			}

		case *ast.ResTarget:
			if n.Name == nil {
				return nil, fmt.Errorf("*ast.ResTarget has nil name")
			}
			key := *n.Name

			// TODO: Deprecate defaultTable
			schema := defaultTable.Schema
			rel := defaultTable.Name
			if ref.rv != nil {
				fqn, err := ParseTableName(ref.rv)
				if err != nil {
					return nil, err
				}
				schema = fqn.Schema
				rel = fqn.Name
			}
			if c, ok := typeMap[schema][rel][key]; ok {
				a = append(a, Parameter{
					Number: ref.ref.Number,
					Column: &Column{
						Name:     parameterName(ref.ref.Number, key),
						DataType: dataType(&c.Type),
						NotNull:  c.IsNotNull,
						IsArray:  c.IsArray,
						Table:    &ast.TableName{Schema: schema, Name: rel},
					},
				})
			} else {
				return nil, &sqlerr.Error{
					Code:     "42703",
					Message:  fmt.Sprintf("306: column \"%s\" does not exist", key),
					Location: n.Location,
				}
			}

		case *ast.TypeCast:
			if n.TypeName == nil {
				return nil, fmt.Errorf("*ast.TypeCast has nil type name")
			}
			col := toColumn(n.TypeName)
			col.Name = parameterName(ref.ref.Number, col.Name)
			a = append(a, Parameter{
				Number: ref.ref.Number,
				Column: col,
			})

		case *ast.ParamRef:
			a = append(a, Parameter{Number: ref.ref.Number})
		case *ast.In:
			if n == nil || n.List == nil {
				fmt.Println("ast.In is nil")
				continue
			}

			number := 0
			if pr, ok := n.List[0].(*ast.ParamRef); ok {
				number = pr.Number
			}

			left, ok := n.Expr.(*ast.ColumnRef)
			if !ok {
				continue
			}
			util.Xiazeminlog("left", left, false)
			util.Xiazeminlog("n", n, false)
			items := stringSlice(left.Fields)
			var key, alias string
			switch len(items) {
			case 1:
				key = items[0]
			case 2:
				alias = items[0]
				key = items[1]
			default:
				panic("too many field items: " + strconv.Itoa(len(items)))
			}

			var found int
			util.Xiazeminlog("range sel-", n, false)
			if n.Sel == nil {
				search := tables
				if alias != "" {
					if original, ok := aliasMap[alias]; ok {
						search = []*ast.TableName{original}
					} else {
						for _, fqn := range tables {
							if fqn.Name == alias {
								search = []*ast.TableName{fqn}
							}
						}
					}
				}

				util.Xiazeminlog("range table", search, false)
				for _, table := range search {
					if c, ok := typeMap[table.Schema][table.Name][key]; ok && n.TableName == table.Name {
						found += 1
						if ref.name != "" {
							key = ref.name
						}
						a = append(a, Parameter{
							Number: number,
							Column: &Column{
								IsSlice:  true,
								Name:     parameterName(ref.ref.Number, key),
								DataType: dataType(&c.Type),
								NotNull:  c.IsNotNull,
								IsArray:  c.IsArray,
								Table:    table,
							},
						})
					}
				}
			} else {
				fmt.Println("------------------------")
			}

			if found == 0 {
				return nil, &sqlerr.Error{
					Code:     "42703",
					Message:  fmt.Sprintf("396: column \"%s\" does not exist", key),
					Location: left.Location,
				}
			}
			if found > 1 {
				util.Xiazeminlog("typeMap", typeMap, false)
				return nil, &sqlerr.Error{
					Code:     "42703",
					Message:  fmt.Sprintf("in same name column reference \"%s\" is ambiguous", key),
					Location: left.Location,
				}
			}
		default:
			fmt.Printf("unsupported reference type: %T", n)
		}
	}
	return a, nil
}
