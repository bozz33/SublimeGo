// Package export provides data export functionality to CSV and Excel formats.
//
// It supports exporting slices of structs with customizable headers using
// struct tags. The package handles type conversion and formatting automatically.
//
// Features:
//   - CSV export
//   - Excel export (XLSX)
//   - Struct tag-based header mapping
//   - Custom formatters
//   - Streaming for large datasets
//
// Basic usage:
//
//	type User struct {
//		ID    int    `export:"ID"`
//		Name  string `export:"Name"`
//		Email string `export:"Email"`
//	}
//
//	users := []User{{1, "John", "john@example.com"}}
//
//	// Export to CSV
//	exporter := export.New(export.FormatCSV).FromStructs(users)
//	exporter.Write(writer)
//
//	// Quick export
//	export.QuickCSV(headers, data, writer)
package export
