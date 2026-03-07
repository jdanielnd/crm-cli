package cli_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMCPServeHelp(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "crm.db")
	stdout, _, code := crm(t, dbPath, "mcp", "serve", "--help")
	assert.Equal(t, 0, code)
	assert.Contains(t, stdout, "MCP server")
}
