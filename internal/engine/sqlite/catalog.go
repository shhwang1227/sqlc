package sqlite

import "github.com/xiazemin/sqlc/internal/sql/catalog"

func NewCatalog() *catalog.Catalog {
	c := catalog.New("main")
	return c
}
