package engine

// Identifiable interface for entities with ID.
type Identifiable interface {
	GetID() int
}

// GetEntityID extracts the ID from an entity in a robust way.
func GetEntityID(item any) int {
	if id, ok := item.(Identifiable); ok {
		return id.GetID()
	}
	return 0
}
