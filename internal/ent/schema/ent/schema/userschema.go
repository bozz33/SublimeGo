package schema

import "entgo.io/ent"

// UserSchema holds the schema definition for the UserSchema entity.
type UserSchema struct {
	ent.Schema
}

// Fields of the UserSchema.
func (UserSchema) Fields() []ent.Field {
	return nil
}

// Edges of the UserSchema.
func (UserSchema) Edges() []ent.Edge {
	return nil
}
