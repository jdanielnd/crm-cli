package cli_test

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContextByID(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "crm.db")
	crm(t, dbPath, "person", "add", "Jane Smith", "-f", "json")
	crm(t, dbPath, "tag", "apply", "person", "1", "vip")
	crm(t, dbPath, "log", "call", "1", "--subject", "Catch up", "-f", "json")
	crm(t, dbPath, "task", "add", "Follow up", "--person", "1", "-f", "json")

	stdout, _, code := crm(t, dbPath, "context", "1", "-f", "json")
	assert.Equal(t, 0, code)

	var data []map[string]any
	require.NoError(t, json.Unmarshal([]byte(stdout), &data))
	assert.Len(t, data, 1)

	result := data[0]
	assert.NotNil(t, result["person"])
	assert.NotNil(t, result["tags"])
	assert.NotNil(t, result["recent_interactions"])
	assert.NotNil(t, result["tasks"])
}

func TestContextByName(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "crm.db")
	crm(t, dbPath, "person", "add", "Jane Smith", "-f", "json")

	stdout, _, code := crm(t, dbPath, "context", "Jane", "-f", "json")
	assert.Equal(t, 0, code)

	var data []map[string]any
	require.NoError(t, json.Unmarshal([]byte(stdout), &data))
	assert.Len(t, data, 1)

	person := data[0]["person"].(map[string]any)
	assert.Equal(t, "Jane", person["first_name"])
}

func TestContextNotFound(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "crm.db")
	_, _, code := crm(t, dbPath, "context", "Nonexistent")
	assert.Equal(t, 3, code)
}
