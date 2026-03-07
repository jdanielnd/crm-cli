package cli_test

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDealAdd(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "crm.db")

	stdout, _, code := crm(t, dbPath, "deal", "add", "Website Redesign", "--value", "15000", "--stage", "proposal", "-f", "json")
	assert.Equal(t, 0, code)

	var data []map[string]any
	require.NoError(t, json.Unmarshal([]byte(stdout), &data))
	assert.Equal(t, "Website Redesign", data[0]["title"])
	assert.Equal(t, "proposal", data[0]["stage"])
	assert.Equal(t, 15000.0, data[0]["value"])
}

func TestDealList(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "crm.db")
	crm(t, dbPath, "deal", "add", "Deal 1", "-f", "json")
	crm(t, dbPath, "deal", "add", "Deal 2", "--stage", "proposal", "-f", "json")

	stdout, _, code := crm(t, dbPath, "deal", "list", "-f", "json")
	assert.Equal(t, 0, code)

	var data []map[string]any
	require.NoError(t, json.Unmarshal([]byte(stdout), &data))
	assert.Len(t, data, 2)
}

func TestDealListFilterByStage(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "crm.db")
	crm(t, dbPath, "deal", "add", "Deal 1", "-f", "json")
	crm(t, dbPath, "deal", "add", "Deal 2", "--stage", "proposal", "-f", "json")

	stdout, _, code := crm(t, dbPath, "deal", "list", "--stage", "proposal", "-f", "json")
	assert.Equal(t, 0, code)

	var data []map[string]any
	require.NoError(t, json.Unmarshal([]byte(stdout), &data))
	assert.Len(t, data, 1)
	assert.Equal(t, "Deal 2", data[0]["title"])
}

func TestDealShow(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "crm.db")
	crm(t, dbPath, "deal", "add", "Test Deal", "-f", "json")

	stdout, _, code := crm(t, dbPath, "deal", "show", "1", "-f", "json")
	assert.Equal(t, 0, code)

	var data []map[string]any
	require.NoError(t, json.Unmarshal([]byte(stdout), &data))
	assert.Equal(t, "Test Deal", data[0]["title"])
}

func TestDealEdit(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "crm.db")
	crm(t, dbPath, "deal", "add", "Test Deal", "-f", "json")

	stdout, _, code := crm(t, dbPath, "deal", "edit", "1", "--stage", "won", "--closed-at", "2026-03-07", "-f", "json")
	assert.Equal(t, 0, code)

	var data []map[string]any
	require.NoError(t, json.Unmarshal([]byte(stdout), &data))
	assert.Equal(t, "won", data[0]["stage"])
	assert.Equal(t, "2026-03-07", data[0]["closed_at"])
}

func TestDealDelete(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "crm.db")
	crm(t, dbPath, "deal", "add", "Test Deal", "-f", "json")

	_, stderr, code := crm(t, dbPath, "deal", "delete", "1")
	assert.Equal(t, 0, code)
	assert.Contains(t, stderr, "Deleted deal #1")

	_, _, code = crm(t, dbPath, "deal", "show", "1")
	assert.Equal(t, 3, code) // not found
}

func TestDealPipeline(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "crm.db")
	crm(t, dbPath, "deal", "add", "Deal 1", "--value", "10000", "-f", "json")
	crm(t, dbPath, "deal", "add", "Deal 2", "--value", "5000", "-f", "json")
	crm(t, dbPath, "deal", "add", "Deal 3", "--value", "20000", "--stage", "won", "-f", "json")

	stdout, _, code := crm(t, dbPath, "deal", "pipeline", "-f", "json")
	assert.Equal(t, 0, code)

	var data []map[string]any
	require.NoError(t, json.Unmarshal([]byte(stdout), &data))
	assert.Len(t, data, 2)
}

func TestDealAddInvalidStage(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "crm.db")
	_, _, code := crm(t, dbPath, "deal", "add", "Bad Deal", "--stage", "invalid")
	assert.Equal(t, 2, code)
}
