package cli_test

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var binaryPath string

func TestMain(m *testing.M) {
	// Build binary once for all integration tests
	tmp, err := os.MkdirTemp("", "crm-test-bin-")
	if err != nil {
		panic(err)
	}
	binaryPath = filepath.Join(tmp, "crm")

	// Find project root (where go.mod lives)
	projectRoot, err := findProjectRoot()
	if err != nil {
		panic("cannot find project root: " + err.Error())
	}
	cmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/crm")
	cmd.Dir = projectRoot
	if out, err := cmd.CombinedOutput(); err != nil {
		panic("build failed: " + string(out))
	}

	code := m.Run()
	os.RemoveAll(tmp)
	os.Exit(code)
}

func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", os.ErrNotExist
		}
		dir = parent
	}
}

func crm(t *testing.T, dbPath string, args ...string) (string, string, int) {
	t.Helper()
	fullArgs := append([]string{"--db", dbPath}, args...)
	cmd := exec.Command(binaryPath, fullArgs...)
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	return stdout.String(), stderr.String(), exitCode
}

func TestVersion(t *testing.T) {
	cmd := exec.Command(binaryPath, "--version")
	out, err := cmd.Output()
	require.NoError(t, err)
	assert.Contains(t, string(out), "crm version")
}

func TestHelp(t *testing.T) {
	stdout, _, code := crm(t, ":memory:", "--help")
	assert.Equal(t, 0, code)
	assert.Contains(t, stdout, "local-first personal CRM")
	assert.Contains(t, stdout, "person")
}

func TestPersonAdd(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "crm.db")
	stdout, _, code := crm(t, dbPath, "person", "add", "Jane Smith", "--email", "jane@example.com", "-f", "json")
	assert.Equal(t, 0, code)

	var data []map[string]any
	require.NoError(t, json.Unmarshal([]byte(stdout), &data))
	assert.Len(t, data, 1)
	assert.Equal(t, "Jane", data[0]["first_name"])
	assert.Equal(t, "Smith", data[0]["last_name"])
	assert.Equal(t, "jane@example.com", data[0]["email"])
}

func TestPersonList(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "crm.db")
	crm(t, dbPath, "person", "add", "Jane Smith", "-f", "json")
	crm(t, dbPath, "person", "add", "Bob Jones", "-f", "json")

	stdout, _, code := crm(t, dbPath, "person", "list", "-f", "json")
	assert.Equal(t, 0, code)

	var data []map[string]any
	require.NoError(t, json.Unmarshal([]byte(stdout), &data))
	assert.Len(t, data, 2)
}

func TestPersonShow(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "crm.db")
	crm(t, dbPath, "person", "add", "Jane Smith", "-f", "json")

	stdout, _, code := crm(t, dbPath, "person", "show", "1", "-f", "json")
	assert.Equal(t, 0, code)

	var data []map[string]any
	require.NoError(t, json.Unmarshal([]byte(stdout), &data))
	assert.Equal(t, "Jane", data[0]["first_name"])
}

func TestPersonShow_NotFound(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "crm.db")
	crm(t, dbPath, "person", "add", "Jane Smith", "-f", "json")

	_, stderr, code := crm(t, dbPath, "person", "show", "999")
	assert.Equal(t, 3, code)
	assert.Contains(t, stderr, "not found")
}

func TestPersonEdit(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "crm.db")
	crm(t, dbPath, "person", "add", "Jane Smith", "-f", "json")

	stdout, _, code := crm(t, dbPath, "person", "edit", "1", "--email", "new@example.com", "-f", "json")
	assert.Equal(t, 0, code)

	var data []map[string]any
	require.NoError(t, json.Unmarshal([]byte(stdout), &data))
	assert.Equal(t, "new@example.com", data[0]["email"])
}

func TestPersonDelete(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "crm.db")
	crm(t, dbPath, "person", "add", "Jane Smith", "-f", "json")

	_, stderr, code := crm(t, dbPath, "person", "delete", "1")
	assert.Equal(t, 0, code)
	assert.Contains(t, stderr, "Deleted person #1")

	stdout, _, _ := crm(t, dbPath, "person", "list", "-f", "json")
	var data []map[string]any
	require.NoError(t, json.Unmarshal([]byte(stdout), &data))
	assert.Len(t, data, 0)
}

func TestPersonDelete_NotFound(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "crm.db")
	_, _, code := crm(t, dbPath, "person", "delete", "999")
	assert.Equal(t, 3, code)
}

func TestPersonQuietMode(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "crm.db")
	crm(t, dbPath, "person", "add", "Jane Smith", "-f", "json")
	crm(t, dbPath, "person", "add", "Bob Jones", "-f", "json")

	stdout, _, code := crm(t, dbPath, "person", "list", "-q")
	assert.Equal(t, 0, code)
	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	assert.Len(t, lines, 2)
}

func TestPersonCSVFormat(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "crm.db")
	crm(t, dbPath, "person", "add", "Jane Smith", "--email", "jane@example.com", "-f", "json")

	stdout, _, code := crm(t, dbPath, "person", "list", "-f", "csv")
	assert.Equal(t, 0, code)
	assert.Contains(t, stdout, "jane@example.com")
	assert.Contains(t, stdout, "id,name,email")
}

func TestPersonInvalidID(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "crm.db")
	_, stderr, code := crm(t, dbPath, "person", "show", "abc")
	assert.Equal(t, 2, code)
	assert.Contains(t, stderr, "invalid person ID")
}
