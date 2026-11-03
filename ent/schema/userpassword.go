package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// UserPassword holds the schema definition for the UserPassword entity.
type UserPassword struct {
	ent.Schema
}

// Fields of the UserPassword.
func (UserPassword) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("user_id", uuid.UUID{}).Immutable(),
		varchar(field.String("hash"), 255).NotEmpty().Sensitive(),
	}
}

// Edges of the UserPassword.
func (UserPassword) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("password").
			Field("user_id").
			Unique().
			Required().
			Immutable(),
	}
}
