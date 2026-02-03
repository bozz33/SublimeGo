// Package auth provides authentication and authorization functionality.
//
// It includes user management, password hashing with bcrypt, session handling,
// and a role-based permission system. The package integrates with the SCS
// session manager for secure session storage.
//
// Features:
//   - User authentication with email/password
//   - Password hashing using bcrypt
//   - Session management with SCS
//   - Role-based access control (RBAC)
//   - Permission checking middleware
//
// Basic usage:
//
//	// Create auth manager
//	auth := auth.NewManager(sessionManager, entClient)
//
//	// Authenticate user
//	user, err := auth.Authenticate(ctx, email, password)
//
//	// Check permissions
//	if auth.Can(ctx, "users.edit") {
//		// Allow action
//	}
package auth
