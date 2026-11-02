package schema

import (
	"fmt"

	"entgo.io/ent/dialect"
)

func varchar[B interface {
	MaxLen(int) B
	SchemaType(map[string]string) B
}](builder B, size int) B {
	return builder.MaxLen(size).SchemaType(map[string]string{dialect.Postgres: fmt.Sprintf("varchar(%d)", size)})
}
