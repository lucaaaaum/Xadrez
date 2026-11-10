package schema

import (
	"regexp"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

type GameRoom struct {
	ent.Schema
}

func (GameRoom) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
	}
}

func (GameRoom) Fields() []ent.Field {
	return []ent.Field{
		varchar(field.String("name"), 255).Match(regexp.MustCompile(`^[a-z0-9][a-z0-9_-]*$`)).Unique().Immutable(),
		field.Enum("status").Values("active", "inactive").Default("active"),
		field.UUID("admin_id", uuid.UUID{}),
	}
}

func (GameRoom) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("admin", User.Type).
			Ref("administered_rooms").
			Field("admin_id").
			Unique().
			Required(),
		edge.To("participants", Participant.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("matches", Match.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}
