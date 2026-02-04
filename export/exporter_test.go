package export

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestUser struct {
	ID        int       `export:"ID"`
	Name      string    `export:"Name"`
	Email     string    `export:"Email"`
	Active    bool      `export:"Active"`
	CreatedAt time.Time `export:"Created At"`
	Internal  string    `export:"-"` // Ignored
}

func TestNew(t *testing.T) {
	exp := New(FormatCSV)
	require.NotNil(t, exp)
	assert.Equal(t, FormatCSV, exp.format)
}

func TestSetHeaders(t *testing.T) {
	exp := New(FormatCSV)
	headers := []string{"ID", "Name", "Email"}
	exp.SetHeaders(headers)
	assert.Equal(t, headers, exp.headers)
}

func TestAddRow(t *testing.T) {
	exp := New(FormatCSV)
	row := []string{"1", "John", "john@example.com"}
	exp.AddRow(row)
	assert.Len(t, exp.data, 1)
	assert.Equal(t, row, exp.data[0])
}

func TestAddRows(t *testing.T) {
	exp := New(FormatCSV)
	rows := [][]string{
		{"1", "John", "john@example.com"},
		{"2", "Jane", "jane@example.com"},
	}
	exp.AddRows(rows)
	assert.Len(t, exp.data, 2)
}

func TestFromStructs(t *testing.T) {
	users := []TestUser{
		{
			ID:        1,
			Name:      "John Doe",
			Email:     "john@example.com",
			Active:    true,
			CreatedAt: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			Internal:  "secret",
		},
		{
			ID:        2,
			Name:      "Jane Smith",
			Email:     "jane@example.com",
			Active:    false,
			CreatedAt: time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC),
			Internal:  "secret2",
		},
	}

	exp := New(FormatCSV).FromStructs(users)

	assert.Len(t, exp.headers, 5) // ID, Name, Email, Active, CreatedAt (Internal ignored)
	assert.Equal(t, "ID", exp.headers[0])
	assert.Equal(t, "Name", exp.headers[1])
	assert.Equal(t, "Email", exp.headers[2])
	assert.Equal(t, "Active", exp.headers[3])
	assert.Equal(t, "Created At", exp.headers[4])

	assert.Len(t, exp.data, 2)
	assert.Equal(t, "1", exp.data[0][0])
	assert.Equal(t, "John Doe", exp.data[0][1])
	assert.Equal(t, "Yes", exp.data[0][3]) // Active = true
	assert.Equal(t, "2", exp.data[1][0])
	assert.Equal(t, "No", exp.data[1][3]) // Active = false
}

func TestWriteCSV(t *testing.T) {
	exp := New(FormatCSV)
	exp.SetHeaders([]string{"ID", "Name", "Email"})
	exp.AddRows([][]string{
		{"1", "John", "john@example.com"},
		{"2", "Jane", "jane@example.com"},
	})

	var buf bytes.Buffer
	err := exp.Write(&buf)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "ID,Name,Email")
	assert.Contains(t, output, "1,John,john@example.com")
	assert.Contains(t, output, "2,Jane,jane@example.com")
}

func TestWriteExcel(t *testing.T) {
	exp := New(FormatExcel)
	exp.SetHeaders([]string{"ID", "Name", "Email"})
	exp.AddRows([][]string{
		{"1", "John", "john@example.com"},
		{"2", "Jane", "jane@example.com"},
	})

	var buf bytes.Buffer
	err := exp.Write(&buf)
	require.NoError(t, err)

	// Verify data was written
	assert.Greater(t, buf.Len(), 0)
}

func TestQuickExportCSV(t *testing.T) {
	headers := []string{"ID", "Name"}
	data := [][]string{
		{"1", "John"},
		{"2", "Jane"},
	}

	var buf bytes.Buffer
	err := QuickExportCSV(&buf, headers, data)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "ID,Name")
	assert.Contains(t, output, "1,John")
}

func TestQuickExportExcel(t *testing.T) {
	headers := []string{"ID", "Name"}
	data := [][]string{
		{"1", "John"},
		{"2", "Jane"},
	}

	var buf bytes.Buffer
	err := QuickExportExcel(&buf, headers, data)
	require.NoError(t, err)
	assert.Greater(t, buf.Len(), 0)
}

func TestExportStructsCSV(t *testing.T) {
	users := []TestUser{
		{ID: 1, Name: "John", Email: "john@example.com", Active: true},
		{ID: 2, Name: "Jane", Email: "jane@example.com", Active: false},
	}

	var buf bytes.Buffer
	err := ExportStructsCSV(&buf, users)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "ID,Name,Email")
	assert.Contains(t, output, "1,John,john@example.com")
}

func TestExportStructsExcel(t *testing.T) {
	users := []TestUser{
		{ID: 1, Name: "John", Email: "john@example.com"},
		{ID: 2, Name: "Jane", Email: "jane@example.com"},
	}

	var buf bytes.Buffer
	err := ExportStructsExcel(&buf, users)
	require.NoError(t, err)
	assert.Greater(t, buf.Len(), 0)
}

func TestColumnName(t *testing.T) {
	tests := []struct {
		index    int
		expected string
	}{
		{0, "A"},
		{1, "B"},
		{25, "Z"},
		{26, "AA"},
		{27, "AB"},
		{51, "AZ"},
		{52, "BA"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := columnName(tt.index)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetContentType(t *testing.T) {
	assert.Equal(t, "text/csv", GetContentType(FormatCSV))
	assert.Equal(t, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", GetContentType(FormatExcel))
}

func TestGetFileExtension(t *testing.T) {
	assert.Equal(t, ".csv", GetFileExtension(FormatCSV))
	assert.Equal(t, ".xlsx", GetFileExtension(FormatExcel))
}

func TestGenerateFilename(t *testing.T) {
	filename := GenerateFilename("users", FormatCSV)
	assert.True(t, strings.HasPrefix(filename, "users_"))
	assert.True(t, strings.HasSuffix(filename, ".csv"))

	filename = GenerateFilename("products", FormatExcel)
	assert.True(t, strings.HasPrefix(filename, "products_"))
	assert.True(t, strings.HasSuffix(filename, ".xlsx"))
}

func BenchmarkExportCSV(b *testing.B) {
	users := make([]TestUser, 1000)
	for i := 0; i < 1000; i++ {
		users[i] = TestUser{
			ID:        i,
			Name:      "User " + string(rune(i)),
			Email:     "user" + string(rune(i)) + "@example.com",
			Active:    i%2 == 0,
			CreatedAt: time.Now(),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		ExportStructsCSV(&buf, users)
	}
}

func BenchmarkExportExcel(b *testing.B) {
	users := make([]TestUser, 1000)
	for i := 0; i < 1000; i++ {
		users[i] = TestUser{
			ID:        i,
			Name:      "User " + string(rune(i)),
			Email:     "user" + string(rune(i)) + "@example.com",
			Active:    i%2 == 0,
			CreatedAt: time.Now(),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		ExportStructsExcel(&buf, users)
	}
}
