package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

type Game struct {
	ent.Schema
}

func (Game) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.Nil).Unique().Immutable(),
		field.Text("name").MaxLen(64),
		field.Time("created").Immutable(),
		field.Text("author").MaxLen(64),
		field.Text("code").MinLen(1).Unique(),
	}
}

func (Game) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("sessions", Session.Type).
			Annotations(entsql.Annotation{
				OnDelete: entsql.Restrict,
			}),
		edge.To("questions", Question.Type).
			Annotations(entsql.Annotation{
				OnDelete: entsql.Cascade,
			}),
	}
}
