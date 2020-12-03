package schema

import (
	"github.com/facebook/ent"
	"github.com/facebook/ent/schema/edge"
	"github.com/facebook/ent/schema/field"
	"github.com/google/uuid"
)

type Choice struct {
	ent.Schema
}

func (Choice) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.New()).Immutable(),
		field.Text("title").MinLen(1).MaxLen(256),
		field.Bool("correct"),
	}
}

func (Choice) Indexes() []ent.Index {
	return []ent.Index{
	}
}

func (Choice) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("question", Question.Type).
			Ref("choices").
			Unique().
			Required(),
		edge.To("answers", Answer.Type),
	}
}

func (Choice) Config() ent.Config {
	return ent.Config{
		Table: "options",
	}
}
