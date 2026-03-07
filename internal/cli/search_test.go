package cli_test

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearch(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "crm.db")
	crm(t, dbPath, "person", "add", "Jane Smith", "--email", "jane@example.com", "-f", "json")
	crm(t, dbPath, "org", "add", "Jane Corp", "-f", "json")
	crm(t, dbPath, "deal", "add", "Jane Project", "-f", "json")

	stdout, _, code := crm(t, dbPath, "search", "Jane", "-f", "json")
	assert.Equal(t, 0, code)

	var data []map[string]any
	require.NoError(t, json.Unmarshal([]byte(stdout), &data))
	assert.GreaterOrEqual(t, len(data), 3) // person + org + deal
}

func TestSearchFilterByType(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "crm.db")
	crm(t, dbPath, "person", "add", "Jane Smith", "-f", "json")
	crm(t, dbPath, "org", "add", "Jane Corp", "-f", "json")

	stdout, _, code := crm(t, dbPath, "search", "Jane", "--type", "person", "-f", "json")
	assert.Equal(t, 0, code)

	var data []map[string]any
	require.NoError(t, json.Unmarshal([]byte(stdout), &data))
	assert.Len(t, data, 1)
	assert.Equal(t, "person", data[0]["type"])
}
