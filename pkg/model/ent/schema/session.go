package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

type Session struct {
	ent.Schema
}

func (Session) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.Nil).Immutable(),
		field.Time("created").Immutable(),
		field.Time("started").Nillable().Optional(),
		field.String("code").MinLen(1).Immutable().Unique(),
	}
}

func (Session) Indexes() []ent.Index {
	return []ent.Index{}
}

func (Session) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("game", Game.Type).
			Ref("sessions").
			Unique().
			Required(),
		edge.To("players", Player.Type),
		edge.To("askedQuestions", AskedQuestion.Type),
	}
}
