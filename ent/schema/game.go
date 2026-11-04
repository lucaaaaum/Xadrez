package schema

import (
	"regexp"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// Game holds the schema definition for the Game entity.
type Game struct {
	ent.Schema
}

func (Game) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
	}
}

// Fields of the Game.
func (Game) Fields() []ent.Field {
	return []ent.Field{
		varchar(field.String("name"), 255).Match(regexp.MustCompile(`^[a-z0-9][a-z0-9_-]*$`)).Unique().Immutable(),
		field.Enum("status").Values("active", "inactive").Default("active"),
		field.UUID("white_player_id", uuid.UUID{}).Immutable(),
		field.UUID("black_player_id", uuid.UUID{}).Immutable(),
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

// Edges of the Game.
func (Game) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("white_player", User.Type).
			Ref("white_games").
			Field("white_player_id").
			Unique().
			Required().
			Immutable(),
		edge.From("black_player", User.Type).
			Ref("black_games").
			Field("black_player_id").
			Unique().
			Required().
			Immutable(),
		edge.To("moves", GameMove.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}
