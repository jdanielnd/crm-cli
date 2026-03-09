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

// parseToolResult extracts the content text and isError from a tool call response.
func parseToolResult(t *testing.T, resp gomcp.JSONRPCMessage) (text string, isError bool) {
	t.Helper()
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
	if len(result.Result.Content) > 0 {
		return result.Result.Content[0].Text, result.Result.IsError
	}
	return "", result.Result.IsError
}

// parseJSON unmarshals the content text from a tool response into a map.
func parseJSONResult(t *testing.T, resp gomcp.JSONRPCMessage) map[string]any {
	t.Helper()
	text, isError := parseToolResult(t, resp)
	require.False(t, isError, "expected success but got error: %s", text)
	var m map[string]any
	require.NoError(t, json.Unmarshal([]byte(text), &m))
	return m
}

func parseJSONArrayResult(t *testing.T, resp gomcp.JSONRPCMessage) []map[string]any {
	t.Helper()
	text, isError := parseToolResult(t, resp)
	require.False(t, isError, "expected success but got error: %s", text)
	var arr []map[string]any
	require.NoError(t, json.Unmarshal([]byte(text), &arr))
	return arr
}

func toolNames(tools map[string]*server.ServerTool) []string {
	names := make([]string, 0, len(tools))
	for name := range tools {
		names = append(names, name)
	}
	return names
}

func TestServerCreation(t *testing.T) {
	s := setupTestServer(t)
	assert.NotNil(t, s)

	tools := s.ListTools()
	assert.Contains(t, toolNames(tools), "crm_person_search")
	assert.Contains(t, toolNames(tools), "crm_person_create")
	assert.Contains(t, toolNames(tools), "crm_stats")
	assert.Contains(t, toolNames(tools), "crm_context")
}

func TestToolCount(t *testing.T) {
	s := setupTestServer(t)
	tools := s.ListTools()
	assert.GreaterOrEqual(t, len(tools), 15, "expected at least 15 tools registered")
}

// --- Person tool tests ---

func TestPersonCreate(t *testing.T) {
	s := setupTestServer(t)

	resp := callTool(t, s, "crm_person_create", map[string]any{
		"first_name": "Jane",
		"last_name":  "Smith",
		"email":      "jane@example.com",
	})

	person := parseJSONResult(t, resp)
	assert.Equal(t, "Jane", person["first_name"])
	assert.Equal(t, "Smith", person["last_name"])
	assert.Equal(t, "jane@example.com", person["email"])
	assert.NotEmpty(t, person["uuid"])
}

func TestPersonCreate_DuplicateEmail(t *testing.T) {
	s := setupTestServer(t)

	callTool(t, s, "crm_person_create", map[string]any{
		"first_name": "Jane",
		"email":      "jane@example.com",
	})
	resp := callTool(t, s, "crm_person_create", map[string]any{
		"first_name": "Janet",
		"email":      "jane@example.com",
	})

	text, isError := parseToolResult(t, resp)
	assert.True(t, isError)
	assert.Contains(t, text, "conflict")
}

func TestPersonGet(t *testing.T) {
	s := setupTestServer(t)

	createResp := callTool(t, s, "crm_person_create", map[string]any{
		"first_name": "Bob",
	})
	created := parseJSONResult(t, createResp)

	resp := callTool(t, s, "crm_person_get", map[string]any{
		"id": created["id"],
	})
	person := parseJSONResult(t, resp)
	assert.Equal(t, "Bob", person["first_name"])
}

func TestPersonGet_NotFound(t *testing.T) {
	s := setupTestServer(t)

	resp := callTool(t, s, "crm_person_get", map[string]any{
		"id": float64(9999),
	})
	text, isError := parseToolResult(t, resp)
	assert.True(t, isError)
	assert.Contains(t, text, "not found")
}

func TestPersonUpdate(t *testing.T) {
	s := setupTestServer(t)

	createResp := callTool(t, s, "crm_person_create", map[string]any{
		"first_name": "Alice",
	})
	created := parseJSONResult(t, createResp)

	resp := callTool(t, s, "crm_person_update", map[string]any{
		"id":         created["id"],
		"first_name": "Alicia",
		"email":      "alicia@test.com",
	})
	updated := parseJSONResult(t, resp)
	assert.Equal(t, "Alicia", updated["first_name"])
	assert.Equal(t, "alicia@test.com", updated["email"])
}

func TestPersonUpdate_NotFound(t *testing.T) {
	s := setupTestServer(t)

	resp := callTool(t, s, "crm_person_update", map[string]any{
		"id":         float64(9999),
		"first_name": "Nobody",
	})
	text, isError := parseToolResult(t, resp)
	assert.True(t, isError)
	assert.Contains(t, text, "not found")
}

func TestPersonDelete(t *testing.T) {
	s := setupTestServer(t)

	createResp := callTool(t, s, "crm_person_create", map[string]any{
		"first_name": "ToDelete",
	})
	created := parseJSONResult(t, createResp)

	resp := callTool(t, s, "crm_person_delete", map[string]any{
		"id": created["id"],
	})
	text, isError := parseToolResult(t, resp)
	assert.False(t, isError)
	assert.Contains(t, text, "deleted")

	// Verify deleted
	getResp := callTool(t, s, "crm_person_get", map[string]any{
		"id": created["id"],
	})
	_, isError = parseToolResult(t, getResp)
	assert.True(t, isError)
}

func TestPersonSearch(t *testing.T) {
	s := setupTestServer(t)

	callTool(t, s, "crm_person_create", map[string]any{
		"first_name": "Searchable",
		"last_name":  "Person",
	})

	resp := callTool(t, s, "crm_person_search", map[string]any{
		"query": "Searchable",
	})
	results := parseJSONArrayResult(t, resp)
	assert.GreaterOrEqual(t, len(results), 1)
	assert.Equal(t, "Searchable", results[0]["first_name"])
}

// --- Deal tool tests ---

func TestDealCreate(t *testing.T) {
	s := setupTestServer(t)

	resp := callTool(t, s, "crm_deal_create", map[string]any{
		"title": "Big Deal",
		"value": float64(50000),
		"stage": "proposal",
	})
	deal := parseJSONResult(t, resp)
	assert.Equal(t, "Big Deal", deal["title"])
	assert.Equal(t, "proposal", deal["stage"])
}

func TestDealCreate_InvalidStage(t *testing.T) {
	s := setupTestServer(t)

	resp := callTool(t, s, "crm_deal_create", map[string]any{
		"title": "Bad Deal",
		"stage": "invalid_stage",
	})
	text, isError := parseToolResult(t, resp)
	assert.True(t, isError)
	assert.Contains(t, text, "validation")
}

func TestDealUpdate(t *testing.T) {
	s := setupTestServer(t)

	createResp := callTool(t, s, "crm_deal_create", map[string]any{
		"title": "Update Me",
		"stage": "lead",
	})
	created := parseJSONResult(t, createResp)

	resp := callTool(t, s, "crm_deal_update", map[string]any{
		"id":    created["id"],
		"stage": "won",
	})
	updated := parseJSONResult(t, resp)
	assert.Equal(t, "won", updated["stage"])
}

// --- Task tool tests ---

func TestTaskCreate(t *testing.T) {
	s := setupTestServer(t)

	resp := callTool(t, s, "crm_task_create", map[string]any{
		"title":    "Follow up",
		"priority": "high",
	})
	task := parseJSONResult(t, resp)
	assert.Equal(t, "Follow up", task["title"])
	assert.Equal(t, "high", task["priority"])
}

func TestTaskCreate_InvalidPriority(t *testing.T) {
	s := setupTestServer(t)

	resp := callTool(t, s, "crm_task_create", map[string]any{
		"title":    "Bad Task",
		"priority": "critical",
	})
	text, isError := parseToolResult(t, resp)
	assert.True(t, isError)
	assert.Contains(t, text, "validation")
}

func TestTaskList(t *testing.T) {
	s := setupTestServer(t)

	callTool(t, s, "crm_task_create", map[string]any{
		"title":    "Task 1",
		"priority": "medium",
	})
	callTool(t, s, "crm_task_create", map[string]any{
		"title":    "Task 2",
		"priority": "low",
	})

	resp := callTool(t, s, "crm_task_list", map[string]any{})
	tasks := parseJSONArrayResult(t, resp)
	assert.GreaterOrEqual(t, len(tasks), 2)
}

// --- Interaction tool tests ---

func TestInteractionLog(t *testing.T) {
	s := setupTestServer(t)

	// Create a person first
	createResp := callTool(t, s, "crm_person_create", map[string]any{
		"first_name": "InteractPerson",
	})
	person := parseJSONResult(t, createResp)

	resp := callTool(t, s, "crm_interaction_log", map[string]any{
		"type":       "call",
		"subject":    "Quick check-in",
		"person_ids": []any{person["id"]},
	})
	interaction := parseJSONResult(t, resp)
	assert.Equal(t, "call", interaction["type"])
}

func TestInteractionLog_InvalidType(t *testing.T) {
	s := setupTestServer(t)

	resp := callTool(t, s, "crm_interaction_log", map[string]any{
		"type":    "invalid_type",
		"subject": "Test",
	})
	text, isError := parseToolResult(t, resp)
	assert.True(t, isError)
	assert.Contains(t, text, "validation")
}

func TestInteractionList(t *testing.T) {
	s := setupTestServer(t)

	callTool(t, s, "crm_interaction_log", map[string]any{
		"type":    "note",
		"subject": "Test note",
	})

	resp := callTool(t, s, "crm_interaction_list", map[string]any{})
	interactions := parseJSONArrayResult(t, resp)
	assert.GreaterOrEqual(t, len(interactions), 1)
}

// --- Tag tool tests ---

func TestTagApply(t *testing.T) {
	s := setupTestServer(t)

	createResp := callTool(t, s, "crm_person_create", map[string]any{
		"first_name": "Tagged",
	})
	person := parseJSONResult(t, createResp)

	resp := callTool(t, s, "crm_tag_apply", map[string]any{
		"entity_type": "person",
		"entity_id":   person["id"],
		"tag":         "vip",
	})
	text, isError := parseToolResult(t, resp)
	assert.False(t, isError)
	assert.Contains(t, text, "Tagged")
}

func TestTagApply_InvalidEntityType(t *testing.T) {
	s := setupTestServer(t)

	resp := callTool(t, s, "crm_tag_apply", map[string]any{
		"entity_type": "invalid_type",
		"entity_id":   float64(1),
		"tag":         "test",
	})
	text, isError := parseToolResult(t, resp)
	assert.True(t, isError)
	assert.Contains(t, text, "validation")
}

// --- Organization tool tests ---

func TestOrgGet_NotFound(t *testing.T) {
	s := setupTestServer(t)

	resp := callTool(t, s, "crm_org_get", map[string]any{
		"id": float64(9999),
	})
	text, isError := parseToolResult(t, resp)
	assert.True(t, isError)
	assert.Contains(t, text, "not found")
}

// --- Relationship tool tests ---

func TestPersonRelate(t *testing.T) {
	s := setupTestServer(t)

	resp1 := callTool(t, s, "crm_person_create", map[string]any{"first_name": "Person1"})
	p1 := parseJSONResult(t, resp1)
	resp2 := callTool(t, s, "crm_person_create", map[string]any{"first_name": "Person2"})
	p2 := parseJSONResult(t, resp2)

	resp := callTool(t, s, "crm_person_relate", map[string]any{
		"person_id":         p1["id"],
		"related_person_id": p2["id"],
		"type":              "colleague",
	})
	rel := parseJSONResult(t, resp)
	assert.Equal(t, "colleague", rel["type"])
}

func TestPersonRelate_InvalidType(t *testing.T) {
	s := setupTestServer(t)

	resp1 := callTool(t, s, "crm_person_create", map[string]any{"first_name": "A"})
	p1 := parseJSONResult(t, resp1)
	resp2 := callTool(t, s, "crm_person_create", map[string]any{"first_name": "B"})
	p2 := parseJSONResult(t, resp2)

	resp := callTool(t, s, "crm_person_relate", map[string]any{
		"person_id":         p1["id"],
		"related_person_id": p2["id"],
		"type":              "enemy",
	})
	text, isError := parseToolResult(t, resp)
	assert.True(t, isError)
	assert.Contains(t, text, "validation")
}

// --- Search tool tests ---

func TestCrossEntitySearch(t *testing.T) {
	s := setupTestServer(t)

	callTool(t, s, "crm_person_create", map[string]any{
		"first_name": "Searchme",
	})

	resp := callTool(t, s, "crm_search", map[string]any{
		"query": "Searchme",
		"type":  "person",
	})
	results := parseJSONResult(t, resp)
	people, ok := results["people"].([]any)
	assert.True(t, ok)
	assert.GreaterOrEqual(t, len(people), 1)
}

// --- Context tool tests ---

func TestContext(t *testing.T) {
	s := setupTestServer(t)

	createResp := callTool(t, s, "crm_person_create", map[string]any{
		"first_name": "ContextPerson",
	})
	person := parseJSONResult(t, createResp)

	resp := callTool(t, s, "crm_context", map[string]any{
		"person_id": person["id"],
	})
	result := parseJSONResult(t, resp)
	p, ok := result["person"].(map[string]any)
	assert.True(t, ok)
	assert.Equal(t, "ContextPerson", p["first_name"])
}

func TestContext_NotFound(t *testing.T) {
	s := setupTestServer(t)

	resp := callTool(t, s, "crm_context", map[string]any{
		"person_id": float64(9999),
	})
	text, isError := parseToolResult(t, resp)
	assert.True(t, isError)
	assert.Contains(t, text, "not found")
}

// --- Stats tool tests ---

func TestStats(t *testing.T) {
	s := setupTestServer(t)

	// Create some data
	callTool(t, s, "crm_person_create", map[string]any{"first_name": "StatsP"})
	callTool(t, s, "crm_deal_create", map[string]any{"title": "StatsD", "stage": "lead"})

	resp := callTool(t, s, "crm_stats", map[string]any{})
	stats := parseJSONResult(t, resp)
	assert.GreaterOrEqual(t, stats["contacts"], float64(1))
	assert.GreaterOrEqual(t, stats["open_deals"], float64(1))
}

// --- Invalid ID tests ---

func TestRequireID_InvalidID(t *testing.T) {
	s := setupTestServer(t)

	tests := []struct {
		name string
		tool string
		args map[string]any
	}{
		{"person_get_zero", "crm_person_get", map[string]any{"id": float64(0)}},
		{"person_get_negative", "crm_person_get", map[string]any{"id": float64(-1)}},
		{"person_delete_zero", "crm_person_delete", map[string]any{"id": float64(0)}},
		{"deal_update_zero", "crm_deal_update", map[string]any{"id": float64(0)}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := callTool(t, s, tt.tool, tt.args)
			text, isError := parseToolResult(t, resp)
			assert.True(t, isError, "expected error for %s", tt.name)
			assert.Contains(t, text, "positive integer")
		})
	}
}
