package engine

import (
	"context"
	"fmt"
	"reflect"
)

// RelationType defines the type of relationship.
type RelationType string

const (
	RelationBelongsTo  RelationType = "belongs_to"
	RelationHasOne     RelationType = "has_one"
	RelationHasMany    RelationType = "has_many"
	RelationManyToMany RelationType = "many_to_many"
)

// Relation defines a relationship between resources.
type Relation struct {
	Name         string       // Name of the relation (e.g., "author", "posts")
	Type         RelationType // Type of relation
	RelatedSlug  string       // Slug of the related resource
	ForeignKey   string       // Foreign key field name
	OwnerKey     string       // Owner key field name (usually "id")
	PivotTable   string       // Pivot table for many-to-many
	DisplayField string       // Field to display in select/list
	Eager        bool         // Whether to eager load by default
}

// RelationBuilder provides a fluent API for defining relations.
type RelationBuilder struct {
	relation *Relation
}

// BelongsTo creates a belongs-to relation.
func BelongsTo(name, relatedSlug string) *RelationBuilder {
	return &RelationBuilder{
		relation: &Relation{
			Name:         name,
			Type:         RelationBelongsTo,
			RelatedSlug:  relatedSlug,
			ForeignKey:   name + "_id",
			OwnerKey:     "id",
			DisplayField: "name",
		},
	}
}

// HasOne creates a has-one relation.
func HasOne(name, relatedSlug string) *RelationBuilder {
	return &RelationBuilder{
		relation: &Relation{
			Name:         name,
			Type:         RelationHasOne,
			RelatedSlug:  relatedSlug,
			OwnerKey:     "id",
			DisplayField: "name",
		},
	}
}

// HasMany creates a has-many relation.
func HasMany(name, relatedSlug string) *RelationBuilder {
	return &RelationBuilder{
		relation: &Relation{
			Name:         name,
			Type:         RelationHasMany,
			RelatedSlug:  relatedSlug,
			OwnerKey:     "id",
			DisplayField: "name",
		},
	}
}

// ManyToMany creates a many-to-many relation.
func ManyToMany(name, relatedSlug string) *RelationBuilder {
	return &RelationBuilder{
		relation: &Relation{
			Name:         name,
			Type:         RelationManyToMany,
			RelatedSlug:  relatedSlug,
			OwnerKey:     "id",
			DisplayField: "name",
		},
	}
}

// ForeignKey sets the foreign key field.
func (rb *RelationBuilder) ForeignKey(key string) *RelationBuilder {
	rb.relation.ForeignKey = key
	return rb
}

// OwnerKey sets the owner key field.
func (rb *RelationBuilder) OwnerKey(key string) *RelationBuilder {
	rb.relation.OwnerKey = key
	return rb
}

// PivotTable sets the pivot table for many-to-many.
func (rb *RelationBuilder) PivotTable(table string) *RelationBuilder {
	rb.relation.PivotTable = table
	return rb
}

// DisplayField sets the field to display.
func (rb *RelationBuilder) DisplayField(field string) *RelationBuilder {
	rb.relation.DisplayField = field
	return rb
}

// Eager enables eager loading.
func (rb *RelationBuilder) Eager() *RelationBuilder {
	rb.relation.Eager = true
	return rb
}

// Build returns the built relation.
func (rb *RelationBuilder) Build() *Relation {
	return rb.relation
}

// RelationAware is the interface for resources that have relations.
type RelationAware interface {
	// GetRelations returns the relations defined for this resource.
	GetRelations() []*Relation
}

// RelationLoader is the interface for loading related data.
type RelationLoader interface {
	// LoadRelation loads related data for an item.
	LoadRelation(ctx context.Context, item any, relation *Relation) (any, error)
	// LoadRelations loads multiple relations for an item.
	LoadRelations(ctx context.Context, item any, relations []*Relation) (map[string]any, error)
}

// RelationOptions provides options for select fields based on relations.
type RelationOptions struct {
	Relation     *Relation
	Options      []SelectOption
	SelectedID   any
	Placeholder  string
	AllowEmpty   bool
	EmptyLabel   string
}

// SelectOption represents an option in a select field.
type SelectOption struct {
	Value    string
	Label    string
	Selected bool
}

// GetRelationOptions fetches options for a relation from the registry.
func GetRelationOptions(ctx context.Context, relation *Relation, selectedID any) (*RelationOptions, error) {
	opts := &RelationOptions{
		Relation:    relation,
		Options:     make([]SelectOption, 0),
		SelectedID:  selectedID,
		Placeholder: fmt.Sprintf("Select %s", relation.Name),
		AllowEmpty:  true,
		EmptyLabel:  "-- None --",
	}

	// This would be implemented to fetch from the related resource
	// For now, return empty options - the actual implementation would
	// use the registry to find the related resource and fetch its data

	return opts, nil
}

// ExtractRelatedID extracts the related ID from an item using reflection.
func ExtractRelatedID(item any, foreignKey string) any {
	val := reflect.ValueOf(item)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return nil
	}

	field := val.FieldByName(foreignKey)
	if !field.IsValid() {
		// Try with different casing
		for i := 0; i < val.NumField(); i++ {
			if val.Type().Field(i).Tag.Get("json") == foreignKey {
				field = val.Field(i)
				break
			}
		}
	}

	if field.IsValid() && field.CanInterface() {
		return field.Interface()
	}
	return nil
}

// SetRelatedID sets the related ID on an item using reflection.
func SetRelatedID(item any, foreignKey string, value any) error {
	val := reflect.ValueOf(item)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return fmt.Errorf("item must be a struct")
	}

	field := val.FieldByName(foreignKey)
	if !field.IsValid() || !field.CanSet() {
		return fmt.Errorf("cannot set field %s", foreignKey)
	}

	field.Set(reflect.ValueOf(value))
	return nil
}

// RelationSchema provides schema information for a relation.
type RelationSchema struct {
	Name        string
	Type        RelationType
	Related     string
	ForeignKey  string
	Nullable    bool
	OnDelete    string // CASCADE, SET NULL, RESTRICT
	OnUpdate    string
}

// GetRelationSchema returns schema information for a relation.
func GetRelationSchema(relation *Relation) *RelationSchema {
	return &RelationSchema{
		Name:       relation.Name,
		Type:       relation.Type,
		Related:    relation.RelatedSlug,
		ForeignKey: relation.ForeignKey,
		Nullable:   true,
		OnDelete:   "SET NULL",
		OnUpdate:   "CASCADE",
	}
}
