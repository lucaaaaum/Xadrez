package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

type Match struct {
	ent.Schema
}

func (Match) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
	}
}

func (Match) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("room_id", uuid.UUID{}).Immutable(),
		field.UUID("white_participant_id", uuid.UUID{}).Immutable(),
		field.UUID("black_participant_id", uuid.UUID{}).Immutable(),
		field.Enum("result").Values("white", "black", "draw").Optional().Nillable(),
		field.Enum("termination").
			Values(
				"checkmate",
				"stalemate",
				"resignation",
				"disconnection",
				"fifty_moves",
				"threefold_repetition",
				"insufficient_material",
			).
			Optional().
			Nillable(),
		field.Time("finished_at").Optional().Nillable(),
	}
}

func (Match) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("room", GameRoom.Type).
			Ref("matches").
			Field("room_id").
			Unique().
			Required().
			Immutable(),
		edge.From("white_player", Participant.Type).
			Ref("white_matches").
			Field("white_participant_id").
			Unique().
			Required().
			Immutable(),
		edge.From("black_player", Participant.Type).
			Ref("black_matches").
			Field("black_participant_id").
			Unique().
			Required().
			Immutable(),
		edge.To("moves", MatchMove.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}
