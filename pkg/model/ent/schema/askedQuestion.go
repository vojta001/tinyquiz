package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

type AskedQuestion struct {
	ent.Schema
}

func (AskedQuestion) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.Nil).Immutable(),
		field.Time("asked").Immutable(),
		field.Time("ended"),
	}
}

func (AskedQuestion) Indexes() []ent.Index {
	return []ent.Index{}
}

func (AskedQuestion) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("session", Session.Type).
			Ref("askedQuestions").
			Unique().
			Required(),
		edge.From("question", Question.Type).
			Ref("asked").
			Unique().
			Required(),
	}
}
