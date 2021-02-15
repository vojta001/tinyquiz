package schema

import (
	"github.com/facebook/ent"
	"github.com/facebook/ent/schema/edge"
	"github.com/facebook/ent/schema/field"
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

func (Answer) Config() ent.Config {
	return ent.Config{
		Table: "answers",
	}
}
