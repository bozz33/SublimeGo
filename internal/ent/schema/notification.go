package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Notification schema
type Notification struct {
	ent.Schema
}

func (Notification) Fields() []ent.Field {
	return []ent.Field{
		field.String("user_id"),
		field.String("title"),
		field.String("body").Optional(),
		field.String("level").Default("info"),
		field.String("icon").Optional(),
		field.String("action_url").Optional(),
		field.String("action_label").Optional(),
		field.Bool("read").Default(false),
		field.Time("created_at").Default(time.Now).Immutable(),
	}
}

func (Notification) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "read"),
	}
}

func (Notification) Edges() []ent.Edge {
	return nil
}
