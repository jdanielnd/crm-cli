package cli_test

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTaskAdd(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "crm.db")

	stdout, _, code := crm(t, dbPath, "task", "add", "Follow up on proposal", "--priority", "high", "--due", "2026-03-14", "-f", "json")
	assert.Equal(t, 0, code)

	var data []map[string]any
	require.NoError(t, json.Unmarshal([]byte(stdout), &data))
	assert.Equal(t, "Follow up on proposal", data[0]["title"])
	assert.Equal(t, "high", data[0]["priority"])
	assert.Equal(t, "2026-03-14", data[0]["due_at"])
}

func TestTaskList(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "crm.db")
	crm(t, dbPath, "task", "add", "Task 1", "-f", "json")
	crm(t, dbPath, "task", "add", "Task 2", "--priority", "high", "-f", "json")

	stdout, _, code := crm(t, dbPath, "task", "list", "-f", "json")
	assert.Equal(t, 0, code)

	var data []map[string]any
	require.NoError(t, json.Unmarshal([]byte(stdout), &data))
	assert.Len(t, data, 2)
}

func TestTaskShow(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "crm.db")
	crm(t, dbPath, "task", "add", "Test Task", "-f", "json")

	stdout, _, code := crm(t, dbPath, "task", "show", "1", "-f", "json")
	assert.Equal(t, 0, code)

	var data []map[string]any
	require.NoError(t, json.Unmarshal([]byte(stdout), &data))
	assert.Equal(t, "Test Task", data[0]["title"])
}

func TestTaskEdit(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "crm.db")
	crm(t, dbPath, "task", "add", "Test Task", "-f", "json")

	stdout, _, code := crm(t, dbPath, "task", "edit", "1", "--priority", "high", "--due", "2026-03-20", "-f", "json")
	assert.Equal(t, 0, code)

	var data []map[string]any
	require.NoError(t, json.Unmarshal([]byte(stdout), &data))
	assert.Equal(t, "high", data[0]["priority"])
	assert.Equal(t, "2026-03-20", data[0]["due_at"])
}

func TestTaskDone(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "crm.db")
	crm(t, dbPath, "task", "add", "Test Task", "-f", "json")

	_, stderr, code := crm(t, dbPath, "task", "done", "1")
	assert.Equal(t, 0, code)
	assert.Contains(t, stderr, "Completed task #1")

	// Should not appear in default list (excludes completed)
	stdout, _, _ := crm(t, dbPath, "task", "list", "-f", "json")
	var data []map[string]any
	json.Unmarshal([]byte(stdout), &data)
	assert.Len(t, data, 0)

	// Should appear with --all
	stdout, _, _ = crm(t, dbPath, "task", "list", "--all", "-f", "json")
	json.Unmarshal([]byte(stdout), &data)
	assert.Len(t, data, 1)
}

func TestTaskDelete(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "crm.db")
	crm(t, dbPath, "task", "add", "Test Task", "-f", "json")

	_, stderr, code := crm(t, dbPath, "task", "delete", "1")
	assert.Equal(t, 0, code)
	assert.Contains(t, stderr, "Deleted task #1")

	_, _, code = crm(t, dbPath, "task", "show", "1")
	assert.Equal(t, 3, code)
}

func TestTaskAddInvalidPriority(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "crm.db")
	_, _, code := crm(t, dbPath, "task", "add", "Bad Task", "--priority", "urgent")
	assert.Equal(t, 2, code)
}
