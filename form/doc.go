// Package form provides a fluent API for building HTML forms.
//
// It supports various field types, validation, layouts, and model binding.
// Forms can be rendered as Templ components and integrate with the
// validation package for server-side validation.
//
// Features:
//   - Fluent field builder API
//   - Multiple field types (text, number, select, checkbox, etc.)
//   - Layout components (sections, grids, tabs)
//   - Model binding for edit forms
//   - Validation integration
//   - CSRF protection
//
// Basic usage:
//
//	form := form.New().SetSchema(
//		form.Text("name").Label("Name").Required(),
//		form.Email("email").Label("Email").Required(),
//		form.Select("role").Label("Role").Options(
//			form.Option("admin", "Administrator"),
//			form.Option("user", "User"),
//		),
//	)
//
//	// Bind model for editing
//	form.BindModel(user)
//
//	// Render form
//	form.Render(ctx)
package form
