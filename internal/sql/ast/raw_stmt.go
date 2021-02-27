package ast

type RawStmt struct {
	Stmt         Node
	StmtLocation int
	StmtLen      int
	TypeName string
}

func (n *RawStmt) Pos() int {
	return n.StmtLocation
}

func (n *RawStmt)GetOpName()string{
	return  n.TypeName
}
