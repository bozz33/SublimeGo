package auth

import (
	"time"

	"github.com/samber/lo"
)

// User represents an authenticated user.
type User struct {
	ID          int
	Email       string
	Name        string
	Permissions []string
	Roles       []string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Metadata    map[string]any
}

// NewUser creates a new user.
func NewUser(id int, email, name string) *User {
	return &User{
		ID:          id,
		Email:       email,
		Name:        name,
		Permissions: []string{},
		Roles:       []string{},
		Metadata:    make(map[string]any),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// HasPermission checks if the user has a specific permission.
func (u *User) HasPermission(perm string) bool {
	if u.IsSuperAdmin() {
		return true
	}

	return lo.Contains(u.Permissions, perm)
}

// HasAnyPermission checks if the user has at least one of the permissions.
func (u *User) HasAnyPermission(perms ...string) bool {
	if u.IsSuperAdmin() {
		return true
	}

	return lo.SomeBy(perms, func(perm string) bool {
		return u.HasPermission(perm)
	})
}

// HasAllPermissions checks if the user has all the permissions.
func (u *User) HasAllPermissions(perms ...string) bool {
	if u.IsSuperAdmin() {
		return true
	}

	return lo.EveryBy(perms, func(perm string) bool {
		return u.HasPermission(perm)
	})
}

// AddPermission adds a permission to the user.
func (u *User) AddPermission(perm string) {
	if !u.HasPermission(perm) {
		u.Permissions = append(u.Permissions, perm)
	}
}

// RemovePermission removes a permission from the user.
func (u *User) RemovePermission(perm string) {
	u.Permissions = lo.Filter(u.Permissions, func(p string, _ int) bool {
		return p != perm
	})
}

// HasRole checks if the user has a specific role.
func (u *User) HasRole(role string) bool {
	return lo.Contains(u.Roles, role)
}

// HasAnyRole checks if the user has at least one of the roles.
func (u *User) HasAnyRole(roles ...string) bool {
	return lo.SomeBy(roles, func(role string) bool {
		return u.HasRole(role)
	})
}

// AddRole adds a role to the user.
func (u *User) AddRole(role string) {
	if !u.HasRole(role) {
		u.Roles = append(u.Roles, role)
	}
}

// RemoveRole removes a role from the user.
func (u *User) RemoveRole(role string) {
	u.Roles = lo.Filter(u.Roles, func(r string, _ int) bool {
		return r != role
	})
}

// IsAdmin checks if the user is an admin.
func (u *User) IsAdmin() bool {
	return u.HasRole("admin")
}

// IsSuperAdmin checks if the user is a super admin.
func (u *User) IsSuperAdmin() bool {
	return u.HasRole("super_admin")
}

// Can is an alias for HasPermission.
func (u *User) Can(perm string) bool {
	return u.HasPermission(perm)
}

// Cannot checks if the user does NOT have a permission.
func (u *User) Cannot(perm string) bool {
	return !u.HasPermission(perm)
}

// GetMetadata retrieves a metadata value.
func (u *User) GetMetadata(key string) any {
	return u.Metadata[key]
}

// SetMetadata sets a metadata value.
func (u *User) SetMetadata(key string, value any) {
	u.Metadata[key] = value
}

// Clone creates a copy of the user.
func (u *User) Clone() *User {
	return &User{
		ID:          u.ID,
		Email:       u.Email,
		Name:        u.Name,
		Permissions: append([]string{}, u.Permissions...),
		Roles:       append([]string{}, u.Roles...),
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
		Metadata:    lo.MapEntries(u.Metadata, func(k string, v any) (string, any) { return k, v }),
	}
}

// ToMap converts the user to a map for session storage.
func (u *User) ToMap() map[string]any {
	return map[string]any{
		"id":          u.ID,
		"email":       u.Email,
		"name":        u.Name,
		"permissions": u.Permissions,
		"roles":       u.Roles,
		"created_at":  u.CreatedAt,
		"updated_at":  u.UpdatedAt,
		"metadata":    u.Metadata,
	}
}

// FromMap creates a user from a map.
func FromMap(data map[string]any) *User {
	user := &User{
		Permissions: []string{},
		Roles:       []string{},
		Metadata:    make(map[string]any),
	}

	if id, ok := data["id"].(int); ok {
		user.ID = id
	}

	if email, ok := data["email"].(string); ok {
		user.Email = email
	}

	if name, ok := data["name"].(string); ok {
		user.Name = name
	}

	if perms, ok := data["permissions"].([]string); ok {
		user.Permissions = perms
	}

	if roles, ok := data["roles"].([]string); ok {
		user.Roles = roles
	}

	if createdAt, ok := data["created_at"].(time.Time); ok {
		user.CreatedAt = createdAt
	}

	if updatedAt, ok := data["updated_at"].(time.Time); ok {
		user.UpdatedAt = updatedAt
	}

	if metadata, ok := data["metadata"].(map[string]any); ok {
		user.Metadata = metadata
	}

	return user
}

// Guest returns a guest user (not authenticated)
func Guest() *User {
	return &User{
		ID:          0,
		Email:       "",
		Name:        "Guest",
		Permissions: []string{},
		Roles:       []string{"guest"},
		Metadata:    make(map[string]any),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// IsGuest checks if the user is a guest
func (u *User) IsGuest() bool {
	return u.ID == 0
}

// IsAuthenticated checks if the user is authenticated
func (u *User) IsAuthenticated() bool {
	return u.ID > 0
}

// Common permissions (constants)
const (
	PermissionUsersView   = "users.view"
	PermissionUsersCreate = "users.create"
	PermissionUsersUpdate = "users.update"
	PermissionUsersDelete = "users.delete"

	PermissionPostsView   = "posts.view"
	PermissionPostsCreate = "posts.create"
	PermissionPostsUpdate = "posts.update"
	PermissionPostsDelete = "posts.delete"

	PermissionAdminAccess = "admin.access"
)

// Common roles (constants)
const (
	RoleGuest      = "guest"
	RoleUser       = "user"
	RoleModerator  = "moderator"
	RoleAdmin      = "admin"
	RoleSuperAdmin = "super_admin"
)
