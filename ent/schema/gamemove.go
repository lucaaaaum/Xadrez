package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// GameMove holds the schema definition for the GameMove entity.
type GameMove struct {
	ent.Schema
}

func (GameMove) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
	}
}

// Fields of the GameMove.
func (GameMove) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("game_id", uuid.UUID{}).Immutable(),
		field.Int("ply").Positive().Immutable(),
		field.Int("from_x").Range(0, 7).Immutable(),
		field.Int("from_y").Range(0, 7).Immutable(),
		field.Int("to_x").Range(0, 7).Immutable(),
		field.Int("to_y").Range(0, 7).Immutable(),
		field.Enum("promotion").Values("queen", "rook", "bishop", "knight").Optional().Nillable().Immutable(),
	}
}

// Edges of the GameMove.
func (GameMove) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("game", Game.Type).
			Ref("moves").
			Field("game_id").
			Unique().
			Required().
			Immutable(),
	}
}

// Indexes of the GameMove.
func (GameMove) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("game_id", "ply").Unique(),
	}
}
