package schema

import (
	"github.com/facebook/ent"
	"github.com/facebook/ent/schema/edge"
	"github.com/facebook/ent/schema/field"
	"github.com/facebook/ent/schema/index"
	"github.com/google/uuid"
	"regexp"
)

type Player struct {
	ent.Schema
}

func (Player) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.New()).Immutable(),
		field.Text("name").MaxLen(64).MinLen(1).Match(regexp.MustCompile("(?:[a-z]|[A-Z]|_|-|.|,|[0-9])+")),
		field.Time("joined").Immutable(),
		field.Bool("organiser").Default(false),
	}
}

func (Player) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("name").Edges("session").Unique(),
	}
}

func (Player) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("session", Session.Type).
			Ref("players").
			Unique().
			Required(),
		edge.To("answers", Answer.Type),
	}
}

func (Player) Config() ent.Config {
	return ent.Config{
		Table: "players",
	}
}
