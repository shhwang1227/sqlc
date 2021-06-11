package catalog

import (
	"fmt"
	"strings"

	"github.com/xiazemin/sqlc/internal/sql/ast"
	"github.com/xiazemin/sqlc/internal/sql/sqlerr"
	"github.com/xiazemin/sqlc/internal/util"
)

func (c *Catalog) schemasToSearch(ns string) []string {
	if ns == "" {
		ns = c.DefaultSchema
	}
	return append(c.SearchPath, ns)
}

func (c *Catalog) getaggColumn(tab *Table, inArgs *ast.List) *Column {
	if tab == nil || inArgs == nil {
		return nil
	}
	//一般都只有一个参数且是一个列名，ifnull 有俩，第一个是列名
	for _, arg := range inArgs.Items {
		switch n := arg.(type) {
		case *ast.FuncCall:
			return c.getaggColumn(tab, n.Args)
		case *ast.ColumnRef:
			if n.Fields == nil {
				continue
			}
			for _, i := range n.Fields.Items {
				util.Xiazeminlog("getaggColumn ", fmt.Sprintf("%T", i), false)
				switch field := i.(type) {
				case *ast.String:
					for _, c := range tab.Columns {
						util.Xiazeminlog("getaggColumn ", []string{c.Name, n.Name}, false)
						if c == nil {
							continue
						}
						if c.Name == field.Str {
							return c
						}
					}
				}
			}
		}
	}
	return nil
}

func (c *Catalog) paramMatch(arg []*Argument, inArgs *ast.List, tables []*Table) bool {
	if inArgs == nil && arg == nil {
		return true
	}
	if tables == nil {
		return true
	}
	if tables[0] == nil {
		return true
	}
	if len(inArgs.Items) != len(arg) {
		return false
	}
	col := c.getaggColumn(tables[0], inArgs)
	if col == nil {
		return true
	}
	for _, a := range arg {
		if a == nil {
			continue
		}

		if a.Type.Name == col.Type.Name {
			return true
		}
	}
	util.Xiazeminlog("paramMatch1", arg, false)
	util.Xiazeminlog("paramMatch2", inArgs, false)
	return false
}

func (c *Catalog) ListFuncsByName(rel *ast.FuncName, args *ast.List, tables []*Table) ([]Function, error) {
	var funcs []Function
	lowered := strings.ToLower(rel.Name)
	for _, ns := range c.schemasToSearch(rel.Schema) {
		s, err := c.getSchema(ns)
		if err != nil {
			return nil, err
		}
		for i := range s.Funcs {
			if strings.ToLower(s.Funcs[i].Name) == lowered && c.paramMatch(s.Funcs[i].Args, args, tables) {
				funcs = append(funcs, *s.Funcs[i])
			}
		}
	}
	return funcs, nil
}

func (c *Catalog) ResolveFuncCall(call *ast.FuncCall, tables []*Table) (*Function, bool, error) {
	// Do not validate unknown functions
	//这里需要加上入参作为判断
	funs, err := c.ListFuncsByName(call.Func, call.Args, tables)
	if err != nil || len(funs) == 0 {
		return nil, false, sqlerr.FunctionNotFound(call.Func.Name)
	}

	// https://www.postgresql.org/docs/current/sql-syntax-calling-funcs.html
	var positional []ast.Node
	var named []*ast.NamedArgExpr

	if call.Args != nil {
		for _, arg := range call.Args.Items {
			if narg, ok := arg.(*ast.NamedArgExpr); ok {
				named = append(named, narg)
			} else {
				// The mixed notation combines positional and named notation.
				// However, as already mentioned, named arguments cannot precede
				// positional arguments.
				if len(named) > 0 {
					return nil, false, &sqlerr.Error{
						Code:     "",
						Message:  "positional argument cannot follow named argument",
						Location: call.Pos(),
					}
				}
				positional = append(positional, arg)
			}
		}
	}

	var col *Column
	if len(tables) > 0 {
		col = c.getaggColumn(tables[0], call.Args)
		if strings.ToLower(call.Func.Name) == "ifnull" {
			col.IsNotNull = true
		}
		util.Xiazeminlog("getaggColumn", col, false)
	}

	for _, fun := range funs {
		args := fun.InArgs()
		var defaults int
		var variadic bool
		known := map[string]struct{}{}
		for _, arg := range args {
			if arg.HasDefault {
				defaults += 1
			}
			if arg.Mode == ast.FuncParamVariadic {
				variadic = true
				defaults += 1
			}
			if arg.Name != "" {
				known[arg.Name] = struct{}{}
			}
		}

		if variadic {
			if (len(named) + len(positional)) < (len(args) - defaults) {
				continue
			}
		} else {
			if (len(named) + len(positional)) > len(args) {
				continue
			}
			if (len(named) + len(positional)) < (len(args) - defaults) {
				continue
			}
		}

		// Validate that the provided named arguments exist in the function
		var unknownArgName bool
		for _, expr := range named {
			if expr.Name != nil {
				if _, found := known[*expr.Name]; !found {
					unknownArgName = true
				}
			}
		}
		if unknownArgName {
			continue
		}
		if col != nil {
			return &fun, col.IsNotNull, nil
		}

		return &fun, true, nil
	}

	var sig []string
	for range call.Args.Items {
		sig = append(sig, "unknown")
	}

	return nil, false, &sqlerr.Error{
		Code:     "42883",
		Message:  fmt.Sprintf("function %s(%s) does not exist", call.Func.Name, strings.Join(sig, ", ")),
		Location: call.Pos(),
		// Hint: "No function matches the given name and argument types. You might need to add explicit type casts.",
	}
}

func (c *Catalog) GetTable(rel *ast.TableName) (Table, error) {
	_, table, err := c.getTable(rel)
	if table == nil {
		return Table{}, err
	} else {
		return *table, err
	}
}
