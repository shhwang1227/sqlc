package ast

type Values struct {
	Location int
}

func (n *Values) Pos() int {
	return n.Location
}
