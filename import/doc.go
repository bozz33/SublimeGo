// Package importer provides data import functionality for SublimeGo.
//
// The importer supports CSV, Excel, and JSON formats with automatic
// column mapping, validation, and transformation.
//
// Example usage:
//
//	// Create an importer with configuration
//	config := importer.DefaultConfig()
//	config.Format = importer.FormatCSV
//	config.Mappings = []importer.ColumnMapping{
//		{SourceColumn: "name", TargetField: "Name", Required: true},
//		{SourceColumn: "email", TargetField: "Email", Required: true},
//		{SourceColumn: "age", TargetField: "Age", Transform: func(v string) (any, error) {
//			return strconv.Atoi(v)
//		}},
//	}
//	config.ValidateRow = func(row map[string]any) error {
//		if row["email"] == "" {
//			return errors.New("email is required")
//		}
//		return nil
//	}
//
//	imp := importer.New(config)
//
//	// Import from a file
//	result, err := imp.ImportFromFile(ctx, file, header, func(ctx context.Context, row map[string]any) error {
//		user := &User{}
//		if err := importer.MapToStruct(row, user); err != nil {
//			return err
//		}
//		return db.User.Create().
//			SetName(user.Name).
//			SetEmail(user.Email).
//			Save(ctx)
//	})
//
//	fmt.Printf("Imported %d rows, %d errors\n", result.SuccessCount, result.ErrorCount)
package importer
