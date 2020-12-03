package schema

import (
	"github.com/facebook/ent"
	"github.com/facebook/ent/schema/edge"
	"github.com/facebook/ent/schema/field"
	"github.com/facebook/ent/schema/index"
	"github.com/google/uuid"
)

type Question struct {
	ent.Schema
}

func (Question) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.New()).Immutable(),
		field.Text("title").MaxLen(256).MinLen(1),
		field.Int("order"),
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
		edge.To("current_sessions", Session.Type),
	}
}

func (Question) Config() ent.Config {
	return ent.Config{
		Table: "questions",
	}
}
