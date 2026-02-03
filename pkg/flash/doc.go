// Package flash provides session-based flash messages for web applications.
//
// Flash messages are temporary messages stored in the session that are
// displayed once and then automatically cleared. They are commonly used
// for success/error notifications after form submissions.
//
// Features:
//   - Success, error, warning, info message types
//   - Session-based storage with SCS
//   - Automatic clearing after display
//   - Multiple messages support
//
// Basic usage:
//
//	manager := flash.NewManager(sessionManager)
//
//	// Add flash messages
//	manager.Success(ctx, "User created successfully")
//	manager.Error(ctx, "Failed to save changes")
//
//	// Retrieve and clear messages
//	messages := manager.GetAndClear(ctx)
//	for _, msg := range messages {
//		// Display message
//	}
package flash
