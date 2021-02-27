package ast

type Statement struct {
	Raw *RawStmt
}

func (n *Statement) Pos() int {
	return 0
}

func (n *Statement)GetOpName()string{
	return n.Raw.GetOpName()
}