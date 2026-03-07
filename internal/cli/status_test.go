package cli_test

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatus(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "crm.db")
	crm(t, dbPath, "person", "add", "Jane Smith", "-f", "json")
	crm(t, dbPath, "person", "add", "Bob Jones", "-f", "json")
	crm(t, dbPath, "org", "add", "Acme Corp", "-f", "json")
	crm(t, dbPath, "deal", "add", "Big Deal", "--value", "10000", "-f", "json")
	crm(t, dbPath, "task", "add", "Follow up", "-f", "json")

	stdout, _, code := crm(t, dbPath, "status", "-f", "json")
	assert.Equal(t, 0, code)

	var data []map[string]any
	require.NoError(t, json.Unmarshal([]byte(stdout), &data))
	assert.Len(t, data, 1)
	assert.Equal(t, float64(2), data[0]["contacts"])
	assert.Equal(t, float64(1), data[0]["organizations"])
	assert.Equal(t, float64(1), data[0]["open_deals"])
	assert.Equal(t, float64(1), data[0]["open_tasks"])
}

func TestStatusEmpty(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "crm.db")

	stdout, _, code := crm(t, dbPath, "status", "-f", "json")
	assert.Equal(t, 0, code)

	var data []map[string]any
	require.NoError(t, json.Unmarshal([]byte(stdout), &data))
	assert.Equal(t, float64(0), data[0]["contacts"])
}
