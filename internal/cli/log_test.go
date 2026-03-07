package cli_test

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogCall(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "crm.db")
	crm(t, dbPath, "person", "add", "Jane Smith", "-f", "json")

	stdout, _, code := crm(t, dbPath, "log", "call", "1", "--subject", "Intro call", "-f", "json")
	assert.Equal(t, 0, code)

	var data []map[string]any
	require.NoError(t, json.Unmarshal([]byte(stdout), &data))
	assert.Equal(t, "call", data[0]["type"])
	assert.Equal(t, "Intro call", data[0]["subject"])
}

func TestLogEmail(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "crm.db")
	crm(t, dbPath, "person", "add", "Jane Smith", "-f", "json")

	stdout, _, code := crm(t, dbPath, "log", "email", "1", "--subject", "Follow-up", "--direction", "outbound", "-f", "json")
	assert.Equal(t, 0, code)

	var data []map[string]any
	require.NoError(t, json.Unmarshal([]byte(stdout), &data))
	assert.Equal(t, "email", data[0]["type"])
	assert.Equal(t, "outbound", data[0]["direction"])
}

func TestLogMeeting_MultiplePeople(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "crm.db")
	crm(t, dbPath, "person", "add", "Jane Smith", "-f", "json")
	crm(t, dbPath, "person", "add", "Bob Jones", "-f", "json")

	stdout, _, code := crm(t, dbPath, "log", "meeting", "1", "2", "--subject", "Product demo", "-f", "json")
	assert.Equal(t, 0, code)

	var data []map[string]any
	require.NoError(t, json.Unmarshal([]byte(stdout), &data))
	assert.Equal(t, "meeting", data[0]["type"])
	personIDs := data[0]["person_ids"].([]any)
	assert.Len(t, personIDs, 2)
}

func TestLogList(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "crm.db")
	crm(t, dbPath, "person", "add", "Jane Smith", "-f", "json")
	crm(t, dbPath, "log", "call", "1", "--subject", "Call 1", "-f", "json")
	crm(t, dbPath, "log", "email", "1", "--subject", "Email 1", "-f", "json")

	stdout, _, code := crm(t, dbPath, "log", "list", "-f", "json")
	assert.Equal(t, 0, code)

	var data []map[string]any
	require.NoError(t, json.Unmarshal([]byte(stdout), &data))
	assert.Len(t, data, 2)
}

func TestLogList_FilterByPerson(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "crm.db")
	crm(t, dbPath, "person", "add", "Jane Smith", "-f", "json")
	crm(t, dbPath, "person", "add", "Bob Jones", "-f", "json")
	crm(t, dbPath, "log", "call", "1", "--subject", "Jane call", "-f", "json")
	crm(t, dbPath, "log", "call", "2", "--subject", "Bob call", "-f", "json")

	stdout, _, code := crm(t, dbPath, "log", "list", "--person", "1", "-f", "json")
	assert.Equal(t, 0, code)

	var data []map[string]any
	require.NoError(t, json.Unmarshal([]byte(stdout), &data))
	assert.Len(t, data, 1)
}

func TestLogList_FilterByType(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "crm.db")
	crm(t, dbPath, "person", "add", "Jane Smith", "-f", "json")
	crm(t, dbPath, "log", "call", "1", "-f", "json")
	crm(t, dbPath, "log", "email", "1", "-f", "json")

	stdout, _, code := crm(t, dbPath, "log", "list", "--type", "call", "-f", "json")
	assert.Equal(t, 0, code)

	var data []map[string]any
	require.NoError(t, json.Unmarshal([]byte(stdout), &data))
	assert.Len(t, data, 1)
	assert.Equal(t, "call", data[0]["type"])
}

func TestLogInvalidPersonID(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "crm.db")
	_, stderr, code := crm(t, dbPath, "log", "call", "abc")
	assert.Equal(t, 2, code)
	assert.Contains(t, stderr, "invalid person ID")
}
