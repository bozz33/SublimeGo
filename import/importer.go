package importer

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

// Format defines the import format.
type Format string

const (
	FormatCSV   Format = "csv"
	FormatExcel Format = "xlsx"
	FormatJSON  Format = "json"
)

// ImportResult contains the result of an import operation.
type ImportResult struct {
	TotalRows     int
	SuccessCount  int
	ErrorCount    int
	SkippedCount  int
	Errors        []ImportError
	Duration      time.Duration
}

// ImportError represents an error during import.
type ImportError struct {
	Row     int
	Column  string
	Value   string
	Message string
}

// ColumnMapping maps CSV/Excel columns to struct fields.
type ColumnMapping struct {
	SourceColumn string // Column name in the file
	TargetField  string // Field name in the struct
	Required     bool
	Default      any
	Transform    func(value string) (any, error)
}

// ImportConfig configures an import operation.
type ImportConfig struct {
	Format          Format
	Mappings        []ColumnMapping
	SkipHeader      bool
	SkipEmptyRows   bool
	StopOnError     bool
	MaxErrors       int
	BatchSize       int
	ValidateRow     func(row map[string]any) error
	BeforeImport    func(row map[string]any) (map[string]any, error)
	AfterImport     func(row map[string]any, result any) error
}

// DefaultConfig returns a default import configuration.
func DefaultConfig() *ImportConfig {
	return &ImportConfig{
		Format:        FormatCSV,
		SkipHeader:    true,
		SkipEmptyRows: true,
		StopOnError:   false,
		MaxErrors:     100,
		BatchSize:     100,
	}
}

// Importer handles data import operations.
type Importer struct {
	config *ImportConfig
}

// New creates a new importer with the given configuration.
func New(config *ImportConfig) *Importer {
	if config == nil {
		config = DefaultConfig()
	}
	return &Importer{config: config}
}

// ImportFromReader imports data from a reader.
func (i *Importer) ImportFromReader(ctx context.Context, reader io.Reader, handler func(ctx context.Context, row map[string]any) error) (*ImportResult, error) {
	start := time.Now()
	result := &ImportResult{
		Errors: make([]ImportError, 0),
	}

	var rows []map[string]any
	var err error

	switch i.config.Format {
	case FormatCSV:
		rows, err = i.parseCSV(reader)
	case FormatJSON:
		rows, err = i.parseJSON(reader)
	default:
		return nil, fmt.Errorf("unsupported format for reader: %s", i.config.Format)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to parse file: %w", err)
	}

	result.TotalRows = len(rows)

	for idx, row := range rows {
		if ctx.Err() != nil {
			return result, ctx.Err()
		}

		// Skip empty rows
		if i.config.SkipEmptyRows && isEmptyRow(row) {
			result.SkippedCount++
			continue
		}

		// Validate row
		if i.config.ValidateRow != nil {
			if err := i.config.ValidateRow(row); err != nil {
				result.ErrorCount++
				result.Errors = append(result.Errors, ImportError{
					Row:     idx + 1,
					Message: err.Error(),
				})
				if i.config.StopOnError || len(result.Errors) >= i.config.MaxErrors {
					break
				}
				continue
			}
		}

		// Before import hook
		if i.config.BeforeImport != nil {
			row, err = i.config.BeforeImport(row)
			if err != nil {
				result.ErrorCount++
				result.Errors = append(result.Errors, ImportError{
					Row:     idx + 1,
					Message: err.Error(),
				})
				continue
			}
		}

		// Process row
		if err := handler(ctx, row); err != nil {
			result.ErrorCount++
			result.Errors = append(result.Errors, ImportError{
				Row:     idx + 1,
				Message: err.Error(),
			})
			if i.config.StopOnError || len(result.Errors) >= i.config.MaxErrors {
				break
			}
			continue
		}

		result.SuccessCount++
	}

	result.Duration = time.Since(start)
	return result, nil
}

// ImportFromFile imports data from a multipart file.
func (i *Importer) ImportFromFile(ctx context.Context, file multipart.File, header *multipart.FileHeader, handler func(ctx context.Context, row map[string]any) error) (*ImportResult, error) {
	// Detect format from filename
	filename := strings.ToLower(header.Filename)
	switch {
	case strings.HasSuffix(filename, ".csv"):
		i.config.Format = FormatCSV
	case strings.HasSuffix(filename, ".xlsx"), strings.HasSuffix(filename, ".xls"):
		i.config.Format = FormatExcel
	case strings.HasSuffix(filename, ".json"):
		i.config.Format = FormatJSON
	default:
		return nil, fmt.Errorf("unsupported file format: %s", filename)
	}

	if i.config.Format == FormatExcel {
		return i.importExcel(ctx, file, handler)
	}

	return i.ImportFromReader(ctx, file, handler)
}

// parseCSV parses CSV data.
func (i *Importer) parseCSV(reader io.Reader) ([]map[string]any, error) {
	csvReader := csv.NewReader(reader)
	csvReader.FieldsPerRecord = -1 // Allow variable number of fields

	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(records) == 0 {
		return []map[string]any{}, nil
	}

	// Get headers
	headers := records[0]
	startRow := 0
	if i.config.SkipHeader {
		startRow = 1
	}

	rows := make([]map[string]any, 0, len(records)-startRow)
	for idx := startRow; idx < len(records); idx++ {
		record := records[idx]
		row := make(map[string]any)

		for j, header := range headers {
			if j < len(record) {
				row[header] = i.transformValue(header, record[j])
			}
		}

		rows = append(rows, row)
	}

	return rows, nil
}

// parseJSON parses JSON data.
func (i *Importer) parseJSON(reader io.Reader) ([]map[string]any, error) {
	var rows []map[string]any
	if err := json.NewDecoder(reader).Decode(&rows); err != nil {
		// Try single object
		var single map[string]any
		if err := json.NewDecoder(reader).Decode(&single); err != nil {
			return nil, err
		}
		rows = []map[string]any{single}
	}
	return rows, nil
}

// importExcel imports from an Excel file.
func (i *Importer) importExcel(ctx context.Context, file io.Reader, handler func(ctx context.Context, row map[string]any) error) (*ImportResult, error) {
	start := time.Now()
	result := &ImportResult{
		Errors: make([]ImportError, 0),
	}

	f, err := excelize.OpenReader(file)
	if err != nil {
		return nil, fmt.Errorf("failed to open Excel file: %w", err)
	}
	defer f.Close()

	// Get the first sheet
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("no sheets found in Excel file")
	}

	rows, err := f.GetRows(sheets[0])
	if err != nil {
		return nil, fmt.Errorf("failed to read rows: %w", err)
	}

	if len(rows) == 0 {
		return result, nil
	}

	// Get headers
	headers := rows[0]
	startRow := 0
	if i.config.SkipHeader {
		startRow = 1
	}

	result.TotalRows = len(rows) - startRow

	for idx := startRow; idx < len(rows); idx++ {
		if ctx.Err() != nil {
			return result, ctx.Err()
		}

		record := rows[idx]
		row := make(map[string]any)

		for j, header := range headers {
			if j < len(record) {
				row[header] = i.transformValue(header, record[j])
			}
		}

		// Skip empty rows
		if i.config.SkipEmptyRows && isEmptyRow(row) {
			result.SkippedCount++
			continue
		}

		// Validate row
		if i.config.ValidateRow != nil {
			if err := i.config.ValidateRow(row); err != nil {
				result.ErrorCount++
				result.Errors = append(result.Errors, ImportError{
					Row:     idx + 1,
					Message: err.Error(),
				})
				if i.config.StopOnError || len(result.Errors) >= i.config.MaxErrors {
					break
				}
				continue
			}
		}

		// Before import hook
		if i.config.BeforeImport != nil {
			row, err = i.config.BeforeImport(row)
			if err != nil {
				result.ErrorCount++
				result.Errors = append(result.Errors, ImportError{
					Row:     idx + 1,
					Message: err.Error(),
				})
				continue
			}
		}

		// Process row
		if err := handler(ctx, row); err != nil {
			result.ErrorCount++
			result.Errors = append(result.Errors, ImportError{
				Row:     idx + 1,
				Message: err.Error(),
			})
			if i.config.StopOnError || len(result.Errors) >= i.config.MaxErrors {
				break
			}
			continue
		}

		result.SuccessCount++
	}

	result.Duration = time.Since(start)
	return result, nil
}

// transformValue transforms a value based on column mappings.
func (i *Importer) transformValue(column, value string) any {
	for _, mapping := range i.config.Mappings {
		if mapping.SourceColumn == column && mapping.Transform != nil {
			transformed, err := mapping.Transform(value)
			if err == nil {
				return transformed
			}
		}
	}
	return value
}

// isEmptyRow checks if a row is empty.
func isEmptyRow(row map[string]any) bool {
	for _, v := range row {
		if v != nil && v != "" {
			return false
		}
	}
	return true
}

// MapToStruct maps a row to a struct using reflection.
func MapToStruct(row map[string]any, dest any) error {
	val := reflect.ValueOf(dest)
	if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("dest must be a pointer to a struct")
	}

	val = val.Elem()
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		if !field.CanSet() {
			continue
		}

		// Get field name from json tag or field name
		jsonTag := fieldType.Tag.Get("json")
		fieldName := strings.SplitN(jsonTag, ",", 2)[0]
		if fieldName == "" || fieldName == "-" {
			fieldName = fieldType.Name
		}

		// Also try lowercase
		value, ok := row[fieldName]
		if !ok {
			value, ok = row[strings.ToLower(fieldName)]
		}
		if !ok {
			continue
		}

		if err := setFieldValue(field, value); err != nil {
			return fmt.Errorf("failed to set field %s: %w", fieldName, err)
		}
	}

	return nil
}

// setFieldValue sets a field value with type conversion.
func setFieldValue(field reflect.Value, value any) error {
	if value == nil {
		return nil
	}

	strValue := fmt.Sprintf("%v", value)

	switch field.Kind() {
	case reflect.String:
		field.SetString(strValue)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intVal, err := strconv.ParseInt(strValue, 10, 64)
		if err != nil {
			return err
		}
		field.SetInt(intVal)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintVal, err := strconv.ParseUint(strValue, 10, 64)
		if err != nil {
			return err
		}
		field.SetUint(uintVal)
	case reflect.Float32, reflect.Float64:
		floatVal, err := strconv.ParseFloat(strValue, 64)
		if err != nil {
			return err
		}
		field.SetFloat(floatVal)
	case reflect.Bool:
		boolVal := strings.ToLower(strValue) == "true" || strValue == "1" || strValue == "yes"
		field.SetBool(boolVal)
	case reflect.Struct:
		if field.Type().String() == "time.Time" {
			// Try common date formats
			formats := []string{
				"2006-01-02",
				"2006-01-02 15:04:05",
				"02/01/2006",
				"01/02/2006",
				time.RFC3339,
			}
			for _, format := range formats {
				if t, err := time.Parse(format, strValue); err == nil {
					field.Set(reflect.ValueOf(t))
					return nil
				}
			}
			return fmt.Errorf("cannot parse date: %s", strValue)
		}
	}

	return nil
}

// Importable is the interface for resources that support import.
type Importable interface {
	// GetImportConfig returns the import configuration.
	GetImportConfig() *ImportConfig
	// GetImportFields returns the fields available for import.
	GetImportFields() []ImportField
	// Import imports a single row.
	Import(ctx context.Context, row map[string]any) error
}

// ImportField describes a field available for import.
type ImportField struct {
	Name        string
	Label       string
	Required    bool
	Type        string // string, int, float, bool, date
	Example     string
	Description string
}

// GetSampleCSV generates a sample CSV for import.
func GetSampleCSV(fields []ImportField) string {
	var headers []string
	var examples []string

	for _, f := range fields {
		headers = append(headers, f.Name)
		examples = append(examples, f.Example)
	}

	return strings.Join(headers, ",") + "\n" + strings.Join(examples, ",")
}
