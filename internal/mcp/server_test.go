package mcp_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/jdanielnd/crm-cli/internal/db"
	crmmcp "github.com/jdanielnd/crm-cli/internal/mcp"
	gomcp "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestServer(t *testing.T) *server.MCPServer {
	t.Helper()
	d, err := db.Open(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { d.Close() })
	return crmmcp.NewServer(d, "test")
}

func callTool(t *testing.T, s *server.MCPServer, name string, args map[string]any) gomcp.JSONRPCMessage {
	t.Helper()
	req := map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/call",
		"params": map[string]any{
			"name":      name,
			"arguments": args,
		},
	}
	raw, err := json.Marshal(req)
	require.NoError(t, err)

	// Need to initialize first
	initReq := map[string]any{
		"jsonrpc": "2.0",
		"id":      0,
		"method":  "initialize",
		"params": map[string]any{
			"protocolVersion": "2025-03-26",
			"capabilities":    map[string]any{},
			"clientInfo": map[string]any{
				"name":    "test",
				"version": "1.0",
			},
		},
	}
	initRaw, _ := json.Marshal(initReq)
	s.HandleMessage(context.Background(), initRaw)

	return s.HandleMessage(context.Background(), raw)
}

func TestServerCreation(t *testing.T) {
	s := setupTestServer(t)
	assert.NotNil(t, s)

	// Verify tools are registered
	tools := s.ListTools()
	assert.Contains(t, toolNames(tools), "crm_person_search")
	assert.Contains(t, toolNames(tools), "crm_person_create")
	assert.Contains(t, toolNames(tools), "crm_stats")
	assert.Contains(t, toolNames(tools), "crm_context")
}

func toolNames(tools map[string]*server.ServerTool) []string {
	names := make([]string, 0, len(tools))
	for name := range tools {
		names = append(names, name)
	}
	return names
}

func TestToolCount(t *testing.T) {
	s := setupTestServer(t)
	tools := s.ListTools()
	// Should have all registered tools
	assert.GreaterOrEqual(t, len(tools), 15, "expected at least 15 tools registered")
}

func TestPersonCreateViaMessage(t *testing.T) {
	s := setupTestServer(t)

	resp := callTool(t, s, "crm_person_create", map[string]any{
		"first_name": "Jane",
		"last_name":  "Smith",
	})

	// Parse response
	raw, err := json.Marshal(resp)
	require.NoError(t, err)

	var result struct {
		Result struct {
			Content []struct {
				Text string `json:"text"`
			} `json:"content"`
			IsError bool `json:"isError"`
		} `json:"result"`
	}
	require.NoError(t, json.Unmarshal(raw, &result))
	assert.False(t, result.Result.IsError)

	var person map[string]any
	require.NoError(t, json.Unmarshal([]byte(result.Result.Content[0].Text), &person))
	assert.Equal(t, "Jane", person["first_name"])
	assert.Equal(t, "Smith", person["last_name"])
}
