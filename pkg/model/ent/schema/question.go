package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type Question struct {
	ent.Schema
}

func (Question) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.Nil).Immutable(),
		field.Text("title").MaxLen(256).MinLen(1),
		field.Int("order"),
		field.Uint64("defaultLength"), // in milliseconds
	}
}

func (Question) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("order").Edges("game").Unique(),
	}
}

func (Question) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("game", Game.Type).
			Ref("questions").
			Unique().
			Required(),
		edge.To("choices", Choice.Type),
		edge.To("asked", AskedQuestion.Type),
	}
}

func (Question) Config() ent.Config {
	return ent.Config{
		Table: "questions",
	}
}
