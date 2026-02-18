package engine

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

// RelationManager is the interface for managing a related resource within a parent resource.
// Equivalent to Filament's RelationManager class.
type RelationManager interface {
	// Name returns the unique identifier of this relation manager (e.g. "posts").
	Name() string
	// Label returns the display label for the tab (e.g. "Posts").
	Label() string
	// Icon returns the icon for the tab.
	Icon() string
	// RelationName returns the name of the relation on the parent model (e.g. "Posts").
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

func (b *BaseRelationManager) Name() string             { return b.name }
func (b *BaseRelationManager) Label() string            { return b.label }
func (b *BaseRelationManager) Icon() string             { return b.icon }
func (b *BaseRelationManager) RelationName() string     { return b.relationName }
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

// RelationManagerHandler handles HTTP requests for relation manager sub-tables.
// Routes handled:
//
//	GET  /{slug}/{id}/relations/{relation}          -> list related items (JSON)
//	POST /{slug}/{id}/relations/{relation}          -> create related item
//	POST /{slug}/{id}/relations/{relation}/attach   -> attach (ManyToMany)
//	POST /{slug}/{id}/relations/{relation}/detach/{relatedID} -> detach
//	DELETE /{slug}/{id}/relations/{relation}/{relatedID}      -> delete
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
func (h *RelationManagerHandler) HasManagers() bool {
	return len(h.managers) > 0
}

// GetManagers returns all relation managers sorted by name.
func (h *RelationManagerHandler) GetManagers() []RelationManager {
	result := make([]RelationManager, 0, len(h.managers))
	for _, rm := range h.managers {
		result = append(result, rm)
	}
	return result
}

// ServeHTTP dispatches relation manager requests.
// Expected URL pattern: /{parentID}/relations/{relationName}[/action[/{relatedID}]]
func (h *RelationManagerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Parse: /relations/{name}[/attach|detach/{relatedID}]
	path := strings.TrimPrefix(r.URL.Path, "/")
	parts := strings.SplitN(path, "/", 4)
	// parts[0] = parentID, parts[1] = "relations", parts[2] = relationName, parts[3] = action/relatedID

	if len(parts) < 3 || parts[1] != "relations" {
		http.NotFound(w, r)
		return
	}

	parentID := parts[0]
	relationName := parts[2]

	rm, ok := h.managers[relationName]
	if !ok {
		http.Error(w, "relation manager not found: "+relationName, http.StatusNotFound)
		return
	}

	// Determine sub-action
	var subAction, relatedID string
	if len(parts) == 4 {
		tail := parts[3]
		if strings.HasPrefix(tail, "detach/") {
			subAction = "detach"
			relatedID = strings.TrimPrefix(tail, "detach/")
		} else if tail == "attach" {
			subAction = "attach"
		} else {
			// treat as relatedID for DELETE
			relatedID = tail
		}
	}

	ctx := r.Context()

	switch r.Method {
	case http.MethodGet:
		items, err := rm.ListRelated(ctx, parentID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"relation": relationName,
			"columns":  rm.Columns(),
			"items":    items,
			"can_create": rm.CanCreate(ctx),
			"can_attach": rm.CanAttach(ctx),
			"can_delete": rm.CanDelete(ctx),
		})

	case http.MethodPost:
		switch subAction {
		case "attach":
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

		default:
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

	case http.MethodDelete:
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

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}
