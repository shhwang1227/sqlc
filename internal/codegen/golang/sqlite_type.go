package golang

import (
	"log"
	"strings"

	"github.com/shhwang1227/sqlc/internal/compiler"
	"github.com/shhwang1227/sqlc/internal/config"
)

func sqliteType(r *compiler.Result, col *compiler.Column, settings config.CombinedSettings) string {
	dt := col.DataType
	notNull := col.NotNull || col.IsArray

	switch dt {

	case "integer":
		if notNull {
			return "int32"
		}
		return "sql.NullInt32"

	case "any":
		return "interface{}"

	}

	switch {

	case strings.HasPrefix(dt, "varchar"):
		if notNull {
			return "string"
		}
		return "sql.NullString"

	default:
		log.Printf("unknown SQLite type: %s\n", dt)
		return "interface{}"

	}
}
