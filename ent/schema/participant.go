package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type Participant struct {
	ent.Schema
}

func (Participant) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
	}
}

func (Participant) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("room_id", uuid.UUID{}).Immutable(),
		field.UUID("user_id", uuid.UUID{}).Immutable(),
	}
}

func (Participant) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("room", GameRoom.Type).
			Ref("participants").
			Field("room_id").
			Unique().
			Required().
			Immutable(),
		edge.From("user", User.Type).
			Ref("participations").
			Field("user_id").
			Unique().
			Required().
			Immutable(),
		edge.To("white_matches", Match.Type),
		edge.To("black_matches", Match.Type),
	}
}

func (Participant) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("room_id", "user_id").Unique(),
	}
}
