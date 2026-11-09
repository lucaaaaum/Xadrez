package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// MatchMove holds the schema definition for the MatchMove entity.
type MatchMove struct {
	ent.Schema
}

func (MatchMove) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
	}
}

// Fields of the MatchMove.
func (MatchMove) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("match_id", uuid.UUID{}).Immutable(),
		field.Int("ply").Positive().Immutable(),
		field.Int("from_x").Range(0, 7).Immutable(),
		field.Int("from_y").Range(0, 7).Immutable(),
		field.Int("to_x").Range(0, 7).Immutable(),
		field.Int("to_y").Range(0, 7).Immutable(),
		field.Enum("promotion").Values("queen", "rook", "bishop", "knight").Optional().Nillable().Immutable(),
	}
}

// Edges of the MatchMove.
func (MatchMove) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("match", Match.Type).
			Ref("moves").
			Field("match_id").
			Unique().
			Required().
			Immutable(),
	}
}

// Indexes of the MatchMove.
func (MatchMove) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("match_id", "ply").Unique(),
	}
}
