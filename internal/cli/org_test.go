package cli_test

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrgAdd(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "crm.db")
	stdout, _, code := crm(t, dbPath, "org", "add", "Acme Corp", "--domain", "acme.com", "-f", "json")
	assert.Equal(t, 0, code)

	var data []map[string]any
	require.NoError(t, json.Unmarshal([]byte(stdout), &data))
	assert.Len(t, data, 1)
	assert.Equal(t, "Acme Corp", data[0]["name"])
	assert.Equal(t, "acme.com", data[0]["domain"])
}

func TestOrgList(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "crm.db")
	crm(t, dbPath, "org", "add", "Acme Corp", "-f", "json")
	crm(t, dbPath, "org", "add", "Globex Inc", "-f", "json")

	stdout, _, code := crm(t, dbPath, "org", "list", "-f", "json")
	assert.Equal(t, 0, code)

	var data []map[string]any
	require.NoError(t, json.Unmarshal([]byte(stdout), &data))
	assert.Len(t, data, 2)
}

func TestOrgShow(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "crm.db")
	crm(t, dbPath, "org", "add", "Acme Corp", "-f", "json")

	stdout, _, code := crm(t, dbPath, "org", "show", "1", "-f", "json")
	assert.Equal(t, 0, code)

	var data []map[string]any
	require.NoError(t, json.Unmarshal([]byte(stdout), &data))
	assert.Equal(t, "Acme Corp", data[0]["name"])
}

func TestOrgShow_NotFound(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "crm.db")
	_, stderr, code := crm(t, dbPath, "org", "show", "999")
	assert.Equal(t, 3, code)
	assert.Contains(t, stderr, "not found")
}

func TestOrgShow_WithPeople(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "crm.db")
	crm(t, dbPath, "org", "add", "Acme Corp", "-f", "json")
	// Add person linked to org (org_id = 1 via --company flag won't work, need to use edit)
	crm(t, dbPath, "person", "add", "Jane Smith", "-f", "json")
	// We can't directly link person to org via CLI yet in Phase 1,
	// but we can test that the flag works and returns empty people list
	stdout, _, code := crm(t, dbPath, "org", "show", "1", "--with-people", "-f", "json")
	assert.Equal(t, 0, code)

	var data []map[string]any
	require.NoError(t, json.Unmarshal([]byte(stdout), &data))
	assert.Equal(t, "Acme Corp", data[0]["name"])
	// people key should exist (empty array or nil is fine)
	_, hasPeople := data[0]["people"]
	assert.True(t, hasPeople)
}

func TestOrgEdit(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "crm.db")
	crm(t, dbPath, "org", "add", "Acme Corp", "-f", "json")

	stdout, _, code := crm(t, dbPath, "org", "edit", "1", "--domain", "acme.io", "-f", "json")
	assert.Equal(t, 0, code)

	var data []map[string]any
	require.NoError(t, json.Unmarshal([]byte(stdout), &data))
	assert.Equal(t, "acme.io", data[0]["domain"])
	assert.Equal(t, "Acme Corp", data[0]["name"])
}

func TestOrgDelete(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "crm.db")
	crm(t, dbPath, "org", "add", "Acme Corp", "-f", "json")

	_, stderr, code := crm(t, dbPath, "org", "delete", "1")
	assert.Equal(t, 0, code)
	assert.Contains(t, stderr, "Deleted organization #1")

	stdout, _, _ := crm(t, dbPath, "org", "list", "-f", "json")
	var data []map[string]any
	json.Unmarshal([]byte(stdout), &data)
	assert.Len(t, data, 0)
}

func TestOrgDelete_NotFound(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "crm.db")
	_, _, code := crm(t, dbPath, "org", "delete", "999")
	assert.Equal(t, 3, code)
}
