package cli_test

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTagApply(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "crm.db")
	crm(t, dbPath, "person", "add", "Jane Smith", "-f", "json")

	_, stderr, code := crm(t, dbPath, "tag", "apply", "person", "1", "vip")
	assert.Equal(t, 0, code)
	assert.Contains(t, stderr, "Tagged person #1")
}

func TestTagShow(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "crm.db")
	crm(t, dbPath, "person", "add", "Jane Smith", "-f", "json")
	crm(t, dbPath, "tag", "apply", "person", "1", "vip")
	crm(t, dbPath, "tag", "apply", "person", "1", "client")

	stdout, _, code := crm(t, dbPath, "tag", "show", "person", "1", "-f", "json")
	assert.Equal(t, 0, code)

	var data []map[string]any
	require.NoError(t, json.Unmarshal([]byte(stdout), &data))
	assert.Len(t, data, 2)
}

func TestTagList(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "crm.db")
	crm(t, dbPath, "person", "add", "Jane Smith", "-f", "json")
	crm(t, dbPath, "tag", "apply", "person", "1", "vip")
	crm(t, dbPath, "tag", "apply", "person", "1", "client")

	stdout, _, code := crm(t, dbPath, "tag", "list", "-f", "json")
	assert.Equal(t, 0, code)

	var data []map[string]any
	require.NoError(t, json.Unmarshal([]byte(stdout), &data))
	assert.Len(t, data, 2)
}

func TestTagRemove(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "crm.db")
	crm(t, dbPath, "person", "add", "Jane Smith", "-f", "json")
	crm(t, dbPath, "tag", "apply", "person", "1", "vip")

	_, stderr, code := crm(t, dbPath, "tag", "remove", "person", "1", "vip")
	assert.Equal(t, 0, code)
	assert.Contains(t, stderr, "Removed tag")

	stdout, _, _ := crm(t, dbPath, "tag", "show", "person", "1", "-f", "json")
	var data []map[string]any
	require.NoError(t, json.Unmarshal([]byte(stdout), &data))
	assert.Len(t, data, 0)
}

func TestTagDelete(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "crm.db")
	crm(t, dbPath, "person", "add", "Jane Smith", "-f", "json")
	crm(t, dbPath, "tag", "apply", "person", "1", "vip")

	_, stderr, code := crm(t, dbPath, "tag", "delete", "vip")
	assert.Equal(t, 0, code)
	assert.Contains(t, stderr, "Deleted tag")
}

func TestTagInvalidEntityType(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "crm.db")
	_, stderr, code := crm(t, dbPath, "tag", "apply", "invalid", "1", "vip")
	assert.Equal(t, 2, code)
	assert.Contains(t, stderr, "invalid entity type")
}
