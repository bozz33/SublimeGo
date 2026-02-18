package export

import (
	"encoding/csv"
	"fmt"
	"io"
	"reflect"
	"time"

	"github.com/xuri/excelize/v2"
)

// Format defines the export format.
type Format string

const (
	FormatCSV   Format = "csv"
	FormatExcel Format = "xlsx"
)

// Exporter manages data export.
type Exporter struct {
	format  Format
	headers []string
	data    [][]string
}

// New creates a new exporter.
func New(format Format) *Exporter {
	return &Exporter{
		format: format,
	}
}

// SetHeaders sets the column headers.
func (e *Exporter) SetHeaders(headers []string) *Exporter {
	e.headers = headers
	return e
}

// AddRow adds a data row.
func (e *Exporter) AddRow(row []string) *Exporter {
	e.data = append(e.data, row)
	return e
}

// AddRows adds multiple rows.
func (e *Exporter) AddRows(rows [][]string) *Exporter {
	e.data = append(e.data, rows...)
	return e
}

// FromStructs converts a slice of structs into exportable data.
func (e *Exporter) FromStructs(items interface{}) *Exporter {
	v := reflect.ValueOf(items)
	if v.Kind() != reflect.Slice {
		return e
	}

	if v.Len() == 0 {
		return e
	}

	first := v.Index(0)
	if first.Kind() == reflect.Ptr {
		first = first.Elem()
	}

	if first.Kind() != reflect.Struct {
		return e
	}

	headers := make([]string, 0)
	for i := 0; i < first.NumField(); i++ {
		field := first.Type().Field(i)

		if !field.IsExported() {
			continue
		}

		name := field.Tag.Get("export")
		if name == "" {
			name = field.Name
		}
		if name != "-" {
			headers = append(headers, name)
		}
	}
	e.headers = headers

	for i := 0; i < v.Len(); i++ {
		item := v.Index(i)
		if item.Kind() == reflect.Ptr {
			item = item.Elem()
		}

		row := make([]string, 0)
		for j := 0; j < item.NumField(); j++ {
			field := item.Type().Field(j)

			if !field.IsExported() {
				continue
			}

			name := field.Tag.Get("export")
			if name == "-" {
				continue
			}

			value := item.Field(j)
			row = append(row, formatValue(value))
		}
		e.data = append(e.data, row)
	}

	return e
}

// Write writes the data to a writer.
func (e *Exporter) Write(w io.Writer) error {
	switch e.format {
	case FormatCSV:
		return e.writeCSV(w)
	case FormatExcel:
		return e.writeExcel(w)
	default:
		return fmt.Errorf("unsupported format: %s", e.format)
	}
}

// writeCSV writes in CSV format.
func (e *Exporter) writeCSV(w io.Writer) error {
	writer := csv.NewWriter(w)
	defer writer.Flush()

	if len(e.headers) > 0 {
		if err := writer.Write(e.headers); err != nil {
			return fmt.Errorf("error writing headers: %w", err)
		}
	}
	for _, row := range e.data {
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("error writing row: %w", err)
		}
	}

	return nil
}

// writeExcel writes in Excel format.
func (e *Exporter) writeExcel(w io.Writer) error {
	f := excelize.NewFile()
	defer func() { _ = f.Close() }()

	sheetName := "Sheet1"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return fmt.Errorf("error creating sheet: %w", err)
	}

	f.SetActiveSheet(index)

	headerStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#E0E0E0"},
			Pattern: 1,
		},
	})
	if err != nil {
		return fmt.Errorf("error creating style: %w", err)
	}

	if len(e.headers) > 0 {
		for i, header := range e.headers {
			cell := fmt.Sprintf("%s1", columnName(i))
			_ = f.SetCellValue(sheetName, cell, header)
			_ = f.SetCellStyle(sheetName, cell, cell, headerStyle)
		}
	}

	for rowIdx, row := range e.data {
		for colIdx, value := range row {
			cell := fmt.Sprintf("%s%d", columnName(colIdx), rowIdx+2)
			_ = f.SetCellValue(sheetName, cell, value)
		}
	}

	for i := range e.headers {
		col := columnName(i)
		_ = f.SetColWidth(sheetName, col, col, 15)
	}

	return f.Write(w)
}

// columnName converts an index to an Excel column name (A, B, C, ..., AA, AB, ...).
func columnName(index int) string {
	name := ""
	index++
	for index > 0 {
		index--
		name = string(rune('A'+(index%26))) + name
		index /= 26
	}
	return name
}

// formatValue converts a reflect value to string.
func formatValue(v reflect.Value) string {
	switch v.Kind() {
	case reflect.String:
		return v.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("%d", v.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fmt.Sprintf("%d", v.Uint())
	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("%.2f", v.Float())
	case reflect.Bool:
		if v.Bool() {
			return "Yes"
		}
		return "No"
	case reflect.Struct:
		if v.Type().String() == "time.Time" {
			t := v.Interface().(time.Time)
			return t.Format("2006-01-02 15:04:05")
		}
		return fmt.Sprintf("%v", v.Interface())
	case reflect.Ptr:
		if v.IsNil() {
			return ""
		}
		return formatValue(v.Elem())
	default:
		return fmt.Sprintf("%v", v.Interface())
	}
}

// QuickExportCSV is a helper to quickly export to CSV.
func QuickExportCSV(w io.Writer, headers []string, data [][]string) error {
	return New(FormatCSV).SetHeaders(headers).AddRows(data).Write(w)
}

// QuickExportExcel is a helper to quickly export to Excel.
func QuickExportExcel(w io.Writer, headers []string, data [][]string) error {
	return New(FormatExcel).SetHeaders(headers).AddRows(data).Write(w)
}

// ExportStructsCSV exports a slice of structs to CSV.
func ExportStructsCSV(w io.Writer, items interface{}) error {
	return New(FormatCSV).FromStructs(items).Write(w)
}

// ExportStructsExcel exports a slice of structs to Excel.
func ExportStructsExcel(w io.Writer, items interface{}) error {
	return New(FormatExcel).FromStructs(items).Write(w)
}

// GetContentType returns the content-type for the format.
func GetContentType(format Format) string {
	switch format {
	case FormatCSV:
		return "text/csv"
	case FormatExcel:
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	default:
		return "application/octet-stream"
	}
}

// GetFileExtension returns the file extension for the format.
func GetFileExtension(format Format) string {
	switch format {
	case FormatCSV:
		return ".csv"
	case FormatExcel:
		return ".xlsx"
	default:
		return ".bin"
	}
}

// GenerateFilename generates a filename with timestamp.
func GenerateFilename(prefix string, format Format) string {
	timestamp := time.Now().Format("20060102_150405")
	return fmt.Sprintf("%s_%s%s", prefix, timestamp, GetFileExtension(format))
}
