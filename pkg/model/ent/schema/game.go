package schema

import (
	"github.com/facebook/ent"
	"github.com/facebook/ent/schema/edge"
	"github.com/facebook/ent/schema/field"
	"github.com/google/uuid"
)

type Game struct {
	ent.Schema
}

func (Game) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.New()).Unique().Immutable(),
		field.Text("name").MaxLen(64),
		field.Time("created").Immutable(),
		field.Text("author").MaxLen(64),
	}
}

func (Game) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("sessions", Session.Type),
		edge.To("questions", Question.Type),
	}
}

func (Game) Config() ent.Config {
	return ent.Config{
		Table: "games",
	}
}
