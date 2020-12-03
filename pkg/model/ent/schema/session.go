package schema

import (
	"github.com/facebook/ent"
	"github.com/facebook/ent/schema/edge"
	"github.com/facebook/ent/schema/field"
	"github.com/google/uuid"
)

type Session struct {
	ent.Schema
}

func (Session) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.New()).Immutable(),
		field.Time("created").Immutable(),
		field.Time("started").Nillable().Optional(), // TODO remove?
		field.Time("current_question_until").Nillable().Optional(),
		field.String("code").MinLen(6).MaxLen(6).Immutable().Unique(),
	}
}

func (Session) Indexes() []ent.Index {
	return []ent.Index{
	}
}

func (Session) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("game", Game.Type).
			Ref("sessions").
			Unique().
			Required(),
		edge.To("players", Player.Type),
		edge.From("current_question", Question.Type).
			Ref("current_sessions").
			Unique(),
	}
}

func (Session) Config() ent.Config {
	return ent.Config{
		Table: "sessions",
	}
}
