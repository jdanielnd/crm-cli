package cli_test

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPersonRelate(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "crm.db")
	crm(t, dbPath, "person", "add", "Jane Smith", "-f", "json")
	crm(t, dbPath, "person", "add", "Bob Jones", "-f", "json")

	_, stderr, code := crm(t, dbPath, "person", "relate", "1", "2", "--type", "colleague")
	assert.Equal(t, 0, code)
	assert.Contains(t, stderr, "Related person #1 to person #2")
}

func TestPersonRelateWithNotes(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "crm.db")
	crm(t, dbPath, "person", "add", "Jane Smith", "-f", "json")
	crm(t, dbPath, "person", "add", "Bob Jones", "-f", "json")

	_, stderr, code := crm(t, dbPath, "person", "relate", "1", "2", "--type", "friend", "--notes", "Met at conference")
	assert.Equal(t, 0, code)
	assert.Contains(t, stderr, "Related person #1 to person #2")
}

func TestPersonRelationships(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "crm.db")
	crm(t, dbPath, "person", "add", "Jane Smith", "-f", "json")
	crm(t, dbPath, "person", "add", "Bob Jones", "-f", "json")
	crm(t, dbPath, "person", "add", "Alice Lee", "-f", "json")
	crm(t, dbPath, "person", "relate", "1", "2", "--type", "colleague")
	crm(t, dbPath, "person", "relate", "3", "1", "--type", "mentor")

	stdout, _, code := crm(t, dbPath, "person", "relationships", "1", "-f", "json")
	assert.Equal(t, 0, code)

	var data []map[string]any
	require.NoError(t, json.Unmarshal([]byte(stdout), &data))
	assert.Len(t, data, 2)
}

func TestPersonRelationshipsAlias(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "crm.db")
	crm(t, dbPath, "person", "add", "Jane Smith", "-f", "json")
	crm(t, dbPath, "person", "add", "Bob Jones", "-f", "json")
	crm(t, dbPath, "person", "relate", "1", "2", "--type", "colleague")

	stdout, _, code := crm(t, dbPath, "person", "rels", "1", "-f", "json")
	assert.Equal(t, 0, code)

	var data []map[string]any
	require.NoError(t, json.Unmarshal([]byte(stdout), &data))
	assert.Len(t, data, 1)
}

func TestPersonUnrelate(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "crm.db")
	crm(t, dbPath, "person", "add", "Jane Smith", "-f", "json")
	crm(t, dbPath, "person", "add", "Bob Jones", "-f", "json")
	crm(t, dbPath, "person", "relate", "1", "2", "--type", "colleague")

	_, stderr, code := crm(t, dbPath, "person", "unrelate", "1")
	assert.Equal(t, 0, code)
	assert.Contains(t, stderr, "Removed relationship #1")

	stdout, _, _ := crm(t, dbPath, "person", "relationships", "1", "-f", "json")
	var data []map[string]any
	json.Unmarshal([]byte(stdout), &data)
	assert.Len(t, data, 0)
}

func TestPersonRelateInvalidType(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "crm.db")
	crm(t, dbPath, "person", "add", "Jane Smith", "-f", "json")
	crm(t, dbPath, "person", "add", "Bob Jones", "-f", "json")

	_, stderr, code := crm(t, dbPath, "person", "relate", "1", "2", "--type", "enemy")
	assert.Equal(t, 2, code)
	assert.Contains(t, stderr, "invalid relationship type")
}
