package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewUser(t *testing.T) {
	user := NewUser(1, "test@example.com", "Test User")

	assert.Equal(t, 1, user.ID)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, "Test User", user.Name)
	assert.NotNil(t, user.Permissions)
	assert.NotNil(t, user.Roles)
	assert.NotNil(t, user.Metadata)
}

func TestUserHasPermission(t *testing.T) {
	user := NewUser(1, "test@example.com", "Test")

	// No permission
	assert.False(t, user.HasPermission("users.create"))

	// Add permission
	user.AddPermission("users.create")
	assert.True(t, user.HasPermission("users.create"))

	// Super admin has all permissions
	user.AddRole(RoleSuperAdmin)
	assert.True(t, user.HasPermission("anything"))
}

func TestUserHasAnyPermission(t *testing.T) {
	user := NewUser(1, "test@example.com", "Test")
	user.AddPermission("users.view")

	assert.True(t, user.HasAnyPermission("users.view", "users.create"))
	assert.False(t, user.HasAnyPermission("users.create", "users.delete"))
}

func TestUserHasAllPermissions(t *testing.T) {
	user := NewUser(1, "test@example.com", "Test")
	user.AddPermission("users.view")
	user.AddPermission("users.create")

	assert.True(t, user.HasAllPermissions("users.view", "users.create"))
	assert.False(t, user.HasAllPermissions("users.view", "users.delete"))
}

func TestUserAddPermission(t *testing.T) {
	user := NewUser(1, "test@example.com", "Test")

	user.AddPermission("users.create")
	assert.Len(t, user.Permissions, 1)

	// Should not add duplicate
	user.AddPermission("users.create")
	assert.Len(t, user.Permissions, 1)
}

func TestUserRemovePermission(t *testing.T) {
	user := NewUser(1, "test@example.com", "Test")
	user.AddPermission("users.create")
	user.AddPermission("users.delete")

	assert.Len(t, user.Permissions, 2)

	user.RemovePermission("users.create")
	assert.Len(t, user.Permissions, 1)
	assert.False(t, user.HasPermission("users.create"))
	assert.True(t, user.HasPermission("users.delete"))
}

func TestUserHasRole(t *testing.T) {
	user := NewUser(1, "test@example.com", "Test")

	assert.False(t, user.HasRole(RoleAdmin))

	user.AddRole(RoleAdmin)
	assert.True(t, user.HasRole(RoleAdmin))
}

func TestUserHasAnyRole(t *testing.T) {
	user := NewUser(1, "test@example.com", "Test")
	user.AddRole(RoleUser)

	assert.True(t, user.HasAnyRole(RoleUser, RoleAdmin))
	assert.False(t, user.HasAnyRole(RoleAdmin, RoleModerator))
}

func TestUserAddRole(t *testing.T) {
	user := NewUser(1, "test@example.com", "Test")

	user.AddRole(RoleAdmin)
	assert.Len(t, user.Roles, 1)

	// Should not add duplicate
	user.AddRole(RoleAdmin)
	assert.Len(t, user.Roles, 1)
}

func TestUserRemoveRole(t *testing.T) {
	user := NewUser(1, "test@example.com", "Test")
	user.AddRole(RoleAdmin)
	user.AddRole(RoleUser)

	assert.Len(t, user.Roles, 2)

	user.RemoveRole(RoleAdmin)
	assert.Len(t, user.Roles, 1)
	assert.False(t, user.HasRole(RoleAdmin))
	assert.True(t, user.HasRole(RoleUser))
}

func TestUserIsAdmin(t *testing.T) {
	user := NewUser(1, "test@example.com", "Test")

	assert.False(t, user.IsAdmin())

	user.AddRole(RoleAdmin)
	assert.True(t, user.IsAdmin())
}

func TestUserIsSuperAdmin(t *testing.T) {
	user := NewUser(1, "test@example.com", "Test")

	assert.False(t, user.IsSuperAdmin())

	user.AddRole(RoleSuperAdmin)
	assert.True(t, user.IsSuperAdmin())
}

func TestUserCan(t *testing.T) {
	user := NewUser(1, "test@example.com", "Test")
	user.AddPermission("users.create")

	assert.True(t, user.Can("users.create"))
	assert.False(t, user.Cannot("users.create"))

	assert.False(t, user.Can("users.delete"))
	assert.True(t, user.Cannot("users.delete"))
}

func TestUserMetadata(t *testing.T) {
	user := NewUser(1, "test@example.com", "Test")

	assert.Nil(t, user.GetMetadata("key"))

	user.SetMetadata("key", "value")
	assert.Equal(t, "value", user.GetMetadata("key"))

	user.SetMetadata("number", 123)
	assert.Equal(t, 123, user.GetMetadata("number"))
}

func TestUserClone(t *testing.T) {
	user := NewUser(1, "test@example.com", "Test")
	user.AddPermission("users.create")
	user.AddRole(RoleAdmin)
	user.SetMetadata("key", "value")

	clone := user.Clone()

	assert.Equal(t, user.ID, clone.ID)
	assert.Equal(t, user.Email, clone.Email)
	assert.Equal(t, user.Permissions, clone.Permissions)
	assert.Equal(t, user.Roles, clone.Roles)

	// Modifying the clone should not affect the original
	clone.AddPermission("users.delete")
	assert.True(t, clone.HasPermission("users.delete"))
	assert.False(t, user.HasPermission("users.delete"))
}

func TestUserToMap(t *testing.T) {
	user := NewUser(1, "test@example.com", "Test")
	user.AddPermission("users.create")
	user.AddRole(RoleAdmin)
	user.SetMetadata("key", "value")

	data := user.ToMap()

	assert.Equal(t, 1, data["id"])
	assert.Equal(t, "test@example.com", data["email"])
	assert.Equal(t, "Test", data["name"])
	assert.NotNil(t, data["permissions"])
	assert.NotNil(t, data["roles"])
	assert.NotNil(t, data["metadata"])
}

func TestFromMap(t *testing.T) {
	data := map[string]any{
		"id":          1,
		"email":       "test@example.com",
		"name":        "Test",
		"permissions": []string{"users.create"},
		"roles":       []string{RoleAdmin},
		"created_at":  time.Now(),
		"updated_at":  time.Now(),
		"metadata":    map[string]any{"key": "value"},
	}

	user := FromMap(data)

	assert.Equal(t, 1, user.ID)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, "Test", user.Name)
	assert.Len(t, user.Permissions, 1)
	assert.Len(t, user.Roles, 1)
}

func TestGuest(t *testing.T) {
	user := Guest()

	assert.Equal(t, 0, user.ID)
	assert.Equal(t, "Guest", user.Name)
	assert.True(t, user.IsGuest())
	assert.False(t, user.IsAuthenticated())
	assert.True(t, user.HasRole(RoleGuest))
}

func TestUserIsGuest(t *testing.T) {
	guest := Guest()
	assert.True(t, guest.IsGuest())

	user := NewUser(1, "test@example.com", "Test")
	assert.False(t, user.IsGuest())
}

func TestUserIsAuthenticated(t *testing.T) {
	guest := Guest()
	assert.False(t, guest.IsAuthenticated())

	user := NewUser(1, "test@example.com", "Test")
	assert.True(t, user.IsAuthenticated())
}

func BenchmarkUserHasPermission(b *testing.B) {
	user := NewUser(1, "test@example.com", "Test")
	user.AddPermission("users.create")
	user.AddPermission("users.view")
	user.AddPermission("users.update")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		user.HasPermission("users.view")
	}
}

func BenchmarkUserHasAllPermissions(b *testing.B) {
	user := NewUser(1, "test@example.com", "Test")
	user.AddPermission("users.create")
	user.AddPermission("users.view")
	user.AddPermission("users.update")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		user.HasAllPermissions("users.create", "users.view")
	}
}

func BenchmarkUserToMap(b *testing.B) {
	user := NewUser(1, "test@example.com", "Test")
	user.AddPermission("users.create")
	user.AddRole(RoleAdmin)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = user.ToMap()
	}
}
