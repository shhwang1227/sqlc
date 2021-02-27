package ast

type In struct {
	// Expr is the value expression to be compared.
	Expr Node
	// List is the list expression in compare list.
	List []Node
	// Not is true, the expression is "not in".
	Not bool
	// Sel is the subquery, may be rewritten to other type of expression.
	Sel      Node
	TypeName string
}

func (n *In) Pos() int {
	return 0
}
