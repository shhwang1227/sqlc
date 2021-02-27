package golang

import (
	"strings"

	"github.com/xiazemin/sqlc/internal/metadata"
)

type QueryValue struct {
	Emit    bool
	Name    string
	Struct  *Struct
	Typ     string
	IsSlice bool
	Slice   []*QueryValue
}

func (v QueryValue) EmitStruct() bool {
	return v.Emit
}

func (v QueryValue) IsStruct() bool {
	return v.Struct != nil
}

func (v QueryValue) isEmpty() bool {
	return v.Typ == "" && v.Name == "" && v.Struct == nil
}

func (v QueryValue) IsSliceType() bool {
	return v.IsSlice
}

func (v QueryValue) ContainSlice() bool {
	if v.Struct != nil {
		for _, f := range v.Struct.Fields {
			if f.IsSlice {
				return true
			}
		}
	}
	if v.IsSlice {
		return true
	}
	return false
}

func (v QueryValue) Pair() string {
	if v.isEmpty() {
		return ""
	}
	if v.IsSlice {
		return v.Name + " []" + v.Type()
	}
	return v.Name + " " + v.Type()
}

func (v QueryValue) Type() string {
	if v.Typ != "" {
		return v.Typ
	}
	if v.Struct != nil {
		return v.Struct.Name
	}
	panic("no type for QueryValue: " + v.Name)
}

func (v QueryValue) Params() string {
	if v.isEmpty() {
		return ""
	}
	var out []string
	if v.Struct == nil {
		if strings.HasPrefix(v.Typ, "[]") && v.Typ != "[]byte" {
			out = append(out, "pq.Array("+v.Name+")")
		} else {
			out = append(out, v.Name)
		}
	} else {
		for _, f := range v.Struct.Fields {
			if strings.HasPrefix(f.Type, "[]") && f.Type != "[]byte" {
				out = append(out, "pq.Array("+v.Name+"."+f.Name+")")
			} else if f.IsSlice {
				out = append(out, v.Name+"."+f.Name)
			} else {
				out = append(out, v.Name+"."+f.Name)
			}
		}
	}
	//fmt.Println(out)
	if len(out) <= 3 {
		return strings.Join(out, ",")
	}
	out = append(out, "")
	return "\n" + strings.Join(out, ",\n")
}

func (v QueryValue) Scan() string {
	var out []string
	if v.Struct == nil {
		if strings.HasPrefix(v.Typ, "[]") && v.Typ != "[]byte" {
			out = append(out, "pq.Array(&"+v.Name+")")
		} else {
			out = append(out, "&"+v.Name)
		}
	} else {
		for _, f := range v.Struct.Fields {
			if strings.HasPrefix(f.Type, "[]") && f.Type != "[]byte" {
				out = append(out, "pq.Array(&"+v.Name+"."+f.Name+")")
			} else {
				out = append(out, "&"+v.Name+"."+f.Name)
			}
		}
	}
	//fmt.Println(out)
	if len(out) <= 3 {
		return strings.Join(out, ",")
	}
	out = append(out, "")
	return "\n" + strings.Join(out, ",\n")
}

// A struct used to generate methods and fields on the Queries struct
type Query struct {
	Cmd          string
	Comments     []string
	MethodName   string
	FieldName    string
	ConstantName string
	SQL          string
	SourceName   string
	Ret          QueryValue
	Arg          QueryValue
}

func (q Query) hasRetType() bool {
	scanned := q.Cmd == metadata.CmdOne || q.Cmd == metadata.CmdMany
	return scanned && !q.Ret.isEmpty()
}
