// Package actions provides a fluent API for defining CRUD actions on resources.
//
// Actions are used in tables and forms to provide interactive functionality
// like editing, deleting, and viewing records. Each action can be customized
// with icons, colors, confirmation dialogs, and custom URL resolvers.
//
// Basic usage:
//
//	editAction := actions.EditAction("/users")
//	deleteAction := actions.DeleteAction("/users").RequireConfirmation()
//
// Custom actions:
//
//	customAction := actions.New("archive").
//		SetLabel("Archive").
//		SetIcon("archive").
//		SetColor("warning").
//		SetUrl(func(item any) string {
//			return fmt.Sprintf("/users/%s/archive", actions.GetItemID(item))
//		})
package actions
