package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"
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
	Relation    *Relation
	Options     []SelectOption
	SelectedID  any
	Placeholder string
	AllowEmpty  bool
	EmptyLabel  string
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
	Name       string
	Type       RelationType
	Related    string
	ForeignKey string
	Nullable   bool
	OnDelete   string // CASCADE, SET NULL, RESTRICT
	OnUpdate   string
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

// ---------------------------------------------------------------------------
// RelationManager — UI interface for managing related records (Filament parity)
// ---------------------------------------------------------------------------

// RelationManager is the interface for managing a related resource within a parent resource.
// Equivalent to Filament's RelationManager class.
type RelationManager interface {
	// Name returns the unique identifier of this relation manager (e.g. "posts").
	Name() string
	// Label returns the display label for the tab (e.g. "Posts").
	Label() string
	// Icon returns the icon for the tab.
	Icon() string
	// RelationName returns the name of the relation on the parent model.
	RelationName() string
	// RelationType returns the type of relation (has_many, many_to_many).
	RelationType() RelationType

	// ListRelated returns the related items for a given parent ID.
	ListRelated(ctx context.Context, parentID string) ([]any, error)
	// AttachRelated attaches a related item to the parent (ManyToMany).
	AttachRelated(ctx context.Context, parentID, relatedID string) error
	// DetachRelated detaches a related item from the parent (ManyToMany).
	DetachRelated(ctx context.Context, parentID, relatedID string) error
	// CreateRelated creates a new related item linked to the parent (HasMany).
	CreateRelated(ctx context.Context, parentID string, r *http.Request) error
	// DeleteRelated deletes a related item.
	DeleteRelated(ctx context.Context, parentID, relatedID string) error

	// Columns returns the columns to display in the sub-table.
	Columns() []Column
	// CanAttach returns whether the user can attach items.
	CanAttach(ctx context.Context) bool
	// CanCreate returns whether the user can create related items.
	CanCreate(ctx context.Context) bool
	// CanDelete returns whether the user can delete related items.
	CanDelete(ctx context.Context) bool
}

// BaseRelationManager provides default no-op implementations for RelationManager.
// Embed this in your concrete relation managers and override what you need.
type BaseRelationManager struct {
	name         string
	label        string
	icon         string
	relationName string
	relationType RelationType
}

// NewBaseRelationManager creates a base relation manager.
func NewBaseRelationManager(name, label, relationName string, relType RelationType) *BaseRelationManager {
	return &BaseRelationManager{
		name:         name,
		label:        label,
		icon:         "link",
		relationName: relationName,
		relationType: relType,
	}
}

func (b *BaseRelationManager) Name() string               { return b.name }
func (b *BaseRelationManager) Label() string              { return b.label }
func (b *BaseRelationManager) Icon() string               { return b.icon }
func (b *BaseRelationManager) RelationName() string       { return b.relationName }
func (b *BaseRelationManager) RelationType() RelationType { return b.relationType }

func (b *BaseRelationManager) ListRelated(_ context.Context, _ string) ([]any, error) {
	return []any{}, nil
}
func (b *BaseRelationManager) AttachRelated(_ context.Context, _, _ string) error { return nil }
func (b *BaseRelationManager) DetachRelated(_ context.Context, _, _ string) error { return nil }
func (b *BaseRelationManager) CreateRelated(_ context.Context, _ string, _ *http.Request) error {
	return nil
}
func (b *BaseRelationManager) DeleteRelated(_ context.Context, _, _ string) error { return nil }
func (b *BaseRelationManager) Columns() []Column                                  { return []Column{} }
func (b *BaseRelationManager) CanAttach(_ context.Context) bool                   { return true }
func (b *BaseRelationManager) CanCreate(_ context.Context) bool                   { return true }
func (b *BaseRelationManager) CanDelete(_ context.Context) bool                   { return true }

// SetIcon sets the icon on the base manager.
func (b *BaseRelationManager) SetIcon(icon string) *BaseRelationManager {
	b.icon = icon
	return b
}

// RelationManagerAware is the interface for resources that expose relation managers.
type RelationManagerAware interface {
	GetRelationManagers() []RelationManager
}

// ---------------------------------------------------------------------------
// RelationManagerHandler — HTTP sub-router for relation manager endpoints
// ---------------------------------------------------------------------------

// RelationManagerHandler handles HTTP requests for relation manager sub-tables.
// Routes handled:
//
//	GET    /{parentID}/relations/{name}              -> list related items (JSON)
//	POST   /{parentID}/relations/{name}              -> create related item
//	POST   /{parentID}/relations/{name}/attach       -> attach (ManyToMany)
//	POST   /{parentID}/relations/{name}/detach/{id}  -> detach (ManyToMany)
//	DELETE /{parentID}/relations/{name}/{id}         -> delete related item
type RelationManagerHandler struct {
	resource Resource
	managers map[string]RelationManager
}

// NewRelationManagerHandler creates a handler for a resource's relation managers.
func NewRelationManagerHandler(resource Resource) *RelationManagerHandler {
	h := &RelationManagerHandler{
		resource: resource,
		managers: make(map[string]RelationManager),
	}
	if rma, ok := resource.(RelationManagerAware); ok {
		for _, rm := range rma.GetRelationManagers() {
			h.managers[rm.Name()] = rm
		}
	}
	return h
}

// HasManagers returns true if the resource has any relation managers.
func (h *RelationManagerHandler) HasManagers() bool { return len(h.managers) > 0 }

// GetManagers returns all registered relation managers.
func (h *RelationManagerHandler) GetManagers() []RelationManager {
	result := make([]RelationManager, 0, len(h.managers))
	for _, rm := range h.managers {
		result = append(result, rm)
	}
	return result
}

// ServeHTTP dispatches relation manager requests.
func (h *RelationManagerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	parentID, relationName, subAction, relatedID, ok := h.parseRelationPath(r)
	if !ok {
		http.NotFound(w, r)
		return
	}
	rm, exists := h.managers[relationName]
	if !exists {
		http.Error(w, "relation manager not found: "+relationName, http.StatusNotFound)
		return
	}
	ctx := r.Context()
	switch r.Method {
	case http.MethodGet:
		h.handleRelationGET(w, rm, parentID, relationName, ctx)
	case http.MethodPost:
		h.handleRelationPOST(w, r, rm, parentID, subAction, ctx)
	case http.MethodDelete:
		h.handleRelationDELETE(w, rm, parentID, relatedID, subAction, ctx)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// parseRelationPath extracts parentID, relationName, subAction, relatedID from the URL.
func (h *RelationManagerHandler) parseRelationPath(r *http.Request) (parentID, relationName, subAction, relatedID string, ok bool) {
	path := strings.TrimPrefix(r.URL.Path, "/")
	parts := strings.SplitN(path, "/", 4)
	if len(parts) < 3 || parts[1] != "relations" {
		return
	}
	parentID, relationName = parts[0], parts[2]
	if len(parts) == 4 {
		tail := parts[3]
		switch {
		case strings.HasPrefix(tail, "detach/"):
			subAction, relatedID = "detach", strings.TrimPrefix(tail, "detach/")
		case tail == "attach":
			subAction = "attach"
		default:
			relatedID = tail
		}
	}
	return parentID, relationName, subAction, relatedID, true
}

func (h *RelationManagerHandler) handleRelationGET(w http.ResponseWriter, rm RelationManager, parentID, relationName string, ctx context.Context) {
	items, err := rm.ListRelated(ctx, parentID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"relation":   relationName,
		"columns":    rm.Columns(),
		"items":      items,
		"can_create": rm.CanCreate(ctx),
		"can_attach": rm.CanAttach(ctx),
		"can_delete": rm.CanDelete(ctx),
	})
}

func (h *RelationManagerHandler) handleRelationPOST(w http.ResponseWriter, r *http.Request, rm RelationManager, parentID, subAction string, ctx context.Context) {
	if subAction == "attach" {
		if !rm.CanAttach(ctx) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		relID := r.FormValue("related_id")
		if relID == "" {
			http.Error(w, "related_id required", http.StatusBadRequest)
			return
		}
		if err := rm.AttachRelated(ctx, parentID, relID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if !rm.CanCreate(ctx) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	if err := rm.CreateRelated(ctx, parentID, r); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *RelationManagerHandler) handleRelationDELETE(w http.ResponseWriter, rm RelationManager, parentID, relatedID, subAction string, ctx context.Context) {
	if subAction == "detach" {
		if !rm.CanAttach(ctx) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		if err := rm.DetachRelated(ctx, parentID, relatedID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		if !rm.CanDelete(ctx) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		if err := rm.DeleteRelated(ctx, parentID, relatedID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	w.WriteHeader(http.StatusNoContent)
}
