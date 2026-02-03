package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// User schema
type User struct {
	ent.Schema
}

// Fields du User
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("name"),
		field.String("email").Unique(),
		field.String("password").Sensitive(),
		field.String("role").Default("user"),

		// Protection systeme
		field.Bool("is_system").Default(false).Comment("True si cet user ne peut pas etre supprime"),

		field.Time("created_at").Default(time.Now).Immutable(),
	}
}

// Edges du User
func (User) Edges() []ent.Edge {
	return nil
}
