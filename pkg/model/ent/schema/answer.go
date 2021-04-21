package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

type Answer struct {
	ent.Schema
}

func (Answer) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.Nil).Immutable().Unique(),
		field.Time("answered").Immutable(),
	}
}

func (Answer) Indexes() []ent.Index {
	return []ent.Index{}
}

func (Answer) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("choice", Choice.Type).
			Ref("answers").
			Unique().
			Required(),
		edge.From("answerer", Player.Type).
			Ref("answers").
			Unique().
			Required(),
	}
}
