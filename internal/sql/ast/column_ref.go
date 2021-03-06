package ast

type ColumnRef struct {
	Name string

	// From pg.ColumnRef
	Fields    *List
	Location  int
	TableName string
}

func (n *ColumnRef) Pos() int {
	return n.Location
}
