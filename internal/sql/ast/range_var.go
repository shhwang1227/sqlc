package ast
//	范围限制变量
type RangeVar struct {
	Catalogname    *string
	Schemaname     *string
	Relname        *string
	Inh            bool
	Relpersistence byte
	Alias          *Alias
	Location       int
	TypeIn bool
}

func (n *RangeVar) Pos() int {
	return n.Location
}
