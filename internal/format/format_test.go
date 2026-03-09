package format_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/jdanielnd/crm-cli/internal/format"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testColumns = []format.ColumnDef{
	{Header: "ID", Field: "id"},
	{Header: "Name", Field: "name"},
	{Header: "Email", Field: "email"},
}

var testData = []map[string]any{
	{"id": 1, "name": "Jane Smith", "email": "jane@example.com"},
	{"id": 2, "name": "Bob Jones", "email": nil},
}

func TestResolve(t *testing.T) {
	tests := []struct {
		input string
		want  format.Format
	}{
		{"json", format.FormatJSON},
		{"JSON", format.FormatJSON},
		{"csv", format.FormatCSV},
		{"tsv", format.FormatTSV},
		{"table", format.FormatTable},
		{"unknown", format.FormatTable},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.want, format.Resolve(tt.input))
		})
	}
}

func TestOutputJSON(t *testing.T) {
	var buf bytes.Buffer
	err := format.Output(&buf, format.FormatJSON, testData, testColumns, false)
	require.NoError(t, err)

	var result []map[string]any
	err = json.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "Jane Smith", result[0]["name"])
}

func TestOutputCSV(t *testing.T) {
	var buf bytes.Buffer
	err := format.Output(&buf, format.FormatCSV, testData, testColumns, false)
	require.NoError(t, err)

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	assert.Len(t, lines, 3) // header + 2 rows
	assert.Equal(t, "ID,Name,Email", lines[0])
	assert.Contains(t, lines[1], "Jane Smith")
	assert.Contains(t, lines[1], "jane@example.com")
}

func TestOutputTSV(t *testing.T) {
	var buf bytes.Buffer
	err := format.Output(&buf, format.FormatTSV, testData, testColumns, false)
	require.NoError(t, err)

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	assert.Contains(t, lines[0], "\t")
}

func TestOutputTable(t *testing.T) {
	var buf bytes.Buffer
	err := format.Output(&buf, format.FormatTable, testData, testColumns, false)
	require.NoError(t, err)

	out := buf.String()
	assert.Contains(t, out, "ID")
	assert.Contains(t, out, "Jane Smith")
	assert.Contains(t, out, "Bob Jones")
}

func TestOutputTable_Empty(t *testing.T) {
	var buf bytes.Buffer
	err := format.Output(&buf, format.FormatTable, nil, testColumns, false)
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "No results.")
}

func TestOutputQuiet(t *testing.T) {
	var buf bytes.Buffer
	err := format.Output(&buf, format.FormatTable, testData, testColumns, true)
	require.NoError(t, err)

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	assert.Equal(t, []string{"1", "2"}, lines)
}
