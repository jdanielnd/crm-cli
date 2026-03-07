package mcp

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/jdanielnd/crm-cli/internal/db/repo"
	"github.com/jdanielnd/crm-cli/internal/model"
	gomcp "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// NewServer creates an MCP server with all CRM tools registered.
func NewServer(db *sql.DB, version string) *server.MCPServer {
	s := server.NewMCPServer(
		"crm",
		version,
		server.WithToolCapabilities(false),
	)

	pr := repo.NewPersonRepo(db)
	or := repo.NewOrgRepo(db)
	ir := repo.NewInteractionRepo(db)
	dr := repo.NewDealRepo(db)
	tr := repo.NewTaskRepo(db)
	tagr := repo.NewTagRepo(db)

	// Person tools
	s.AddTool(
		gomcp.NewTool("crm_person_search",
			gomcp.WithDescription("Search people by name, email, tag, or org"),
			gomcp.WithString("query", gomcp.Required(), gomcp.Description("Search query")),
			gomcp.WithNumber("limit", gomcp.Description("Max results (default 20)")),
		),
		personSearchHandler(pr),
	)

	s.AddTool(
		gomcp.NewTool("crm_person_get",
			gomcp.WithDescription("Get full details for a person by ID"),
			gomcp.WithNumber("id", gomcp.Required(), gomcp.Description("Person ID")),
		),
		personGetHandler(pr),
	)

	s.AddTool(
		gomcp.NewTool("crm_person_create",
			gomcp.WithDescription("Create a new person"),
			gomcp.WithString("first_name", gomcp.Required(), gomcp.Description("First name")),
			gomcp.WithString("last_name", gomcp.Description("Last name")),
			gomcp.WithString("email", gomcp.Description("Email address")),
			gomcp.WithString("phone", gomcp.Description("Phone number")),
			gomcp.WithString("title", gomcp.Description("Job title")),
			gomcp.WithString("company", gomcp.Description("Company name")),
			gomcp.WithString("location", gomcp.Description("Location")),
			gomcp.WithString("notes", gomcp.Description("Notes")),
		),
		personCreateHandler(pr),
	)

	s.AddTool(
		gomcp.NewTool("crm_person_update",
			gomcp.WithDescription("Update person fields"),
			gomcp.WithNumber("id", gomcp.Required(), gomcp.Description("Person ID")),
			gomcp.WithString("first_name", gomcp.Description("First name")),
			gomcp.WithString("last_name", gomcp.Description("Last name")),
			gomcp.WithString("email", gomcp.Description("Email address")),
			gomcp.WithString("phone", gomcp.Description("Phone number")),
			gomcp.WithString("title", gomcp.Description("Job title")),
			gomcp.WithString("company", gomcp.Description("Company name")),
			gomcp.WithString("location", gomcp.Description("Location")),
			gomcp.WithString("notes", gomcp.Description("Notes")),
			gomcp.WithString("summary", gomcp.Description("AI-maintained summary/dossier")),
		),
		personUpdateHandler(pr),
	)

	s.AddTool(
		gomcp.NewTool("crm_person_delete",
			gomcp.WithDescription("Delete (archive) a person by ID"),
			gomcp.WithNumber("id", gomcp.Required(), gomcp.Description("Person ID")),
		),
		personDeleteHandler(pr),
	)

	// Organization tools
	s.AddTool(
		gomcp.NewTool("crm_org_search",
			gomcp.WithDescription("Search organizations"),
			gomcp.WithString("query", gomcp.Required(), gomcp.Description("Search query")),
			gomcp.WithNumber("limit", gomcp.Description("Max results (default 20)")),
		),
		orgSearchHandler(or),
	)

	s.AddTool(
		gomcp.NewTool("crm_org_get",
			gomcp.WithDescription("Get organization details with people"),
			gomcp.WithNumber("id", gomcp.Required(), gomcp.Description("Organization ID")),
		),
		orgGetHandler(or),
	)

	// Interaction tools
	s.AddTool(
		gomcp.NewTool("crm_interaction_log",
			gomcp.WithDescription("Log an interaction (call, email, meeting, note, message)"),
			gomcp.WithString("type", gomcp.Required(), gomcp.Description("Interaction type: call, email, meeting, note, message")),
			gomcp.WithString("subject", gomcp.Description("Subject/title")),
			gomcp.WithString("content", gomcp.Description("Content/body")),
			gomcp.WithString("direction", gomcp.Description("Direction: inbound or outbound")),
			gomcp.WithString("occurred_at", gomcp.Description("When it occurred (ISO 8601)")),
			gomcp.WithArray("person_ids", gomcp.Description("Array of person IDs to link this interaction to")),
		),
		interactionLogHandler(ir),
	)

	s.AddTool(
		gomcp.NewTool("crm_interaction_list",
			gomcp.WithDescription("List interactions, optionally filtered by person"),
			gomcp.WithNumber("person_id", gomcp.Description("Filter by person ID")),
			gomcp.WithString("type", gomcp.Description("Filter by type")),
			gomcp.WithNumber("limit", gomcp.Description("Max results (default 20)")),
		),
		interactionListHandler(ir),
	)

	// Search
	s.AddTool(
		gomcp.NewTool("crm_search",
			gomcp.WithDescription("Cross-entity full-text search across people, organizations, interactions, and deals"),
			gomcp.WithString("query", gomcp.Required(), gomcp.Description("Search query")),
			gomcp.WithString("type", gomcp.Description("Filter by entity type: person, organization, interaction, deal")),
			gomcp.WithNumber("limit", gomcp.Description("Max results per type (default 20)")),
		),
		searchHandler(pr, or, ir, dr),
	)

	// Context
	s.AddTool(
		gomcp.NewTool("crm_context",
			gomcp.WithDescription("Full context briefing for a person: profile, org, interactions, deals, tasks, relationships, tags"),
			gomcp.WithNumber("person_id", gomcp.Required(), gomcp.Description("Person ID")),
		),
		contextHandler(pr, or, ir, dr, tr, repo.NewRelationshipRepo(db), tagr),
	)

	// Deal tools
	s.AddTool(
		gomcp.NewTool("crm_deal_create",
			gomcp.WithDescription("Create a new deal"),
			gomcp.WithString("title", gomcp.Required(), gomcp.Description("Deal title")),
			gomcp.WithNumber("value", gomcp.Description("Deal value")),
			gomcp.WithString("stage", gomcp.Description("Stage: lead, prospect, proposal, negotiation, won, lost (default: lead)")),
			gomcp.WithNumber("person_id", gomcp.Description("Associated person ID")),
			gomcp.WithNumber("org_id", gomcp.Description("Associated organization ID")),
			gomcp.WithString("notes", gomcp.Description("Notes")),
		),
		dealCreateHandler(dr),
	)

	s.AddTool(
		gomcp.NewTool("crm_deal_update",
			gomcp.WithDescription("Update a deal (stage, value, etc.)"),
			gomcp.WithNumber("id", gomcp.Required(), gomcp.Description("Deal ID")),
			gomcp.WithString("title", gomcp.Description("Deal title")),
			gomcp.WithNumber("value", gomcp.Description("Deal value")),
			gomcp.WithString("stage", gomcp.Description("Stage: lead, prospect, proposal, negotiation, won, lost")),
			gomcp.WithString("notes", gomcp.Description("Notes")),
			gomcp.WithString("closed_at", gomcp.Description("Date the deal was closed")),
		),
		dealUpdateHandler(dr),
	)

	// Task tools
	s.AddTool(
		gomcp.NewTool("crm_task_create",
			gomcp.WithDescription("Create a follow-up task"),
			gomcp.WithString("title", gomcp.Required(), gomcp.Description("Task title")),
			gomcp.WithString("description", gomcp.Description("Description")),
			gomcp.WithNumber("person_id", gomcp.Description("Associated person ID")),
			gomcp.WithNumber("deal_id", gomcp.Description("Associated deal ID")),
			gomcp.WithString("due", gomcp.Description("Due date (ISO 8601)")),
			gomcp.WithString("priority", gomcp.Description("Priority: low, medium, high (default: medium)")),
		),
		taskCreateHandler(tr),
	)

	s.AddTool(
		gomcp.NewTool("crm_task_list",
			gomcp.WithDescription("List open tasks"),
			gomcp.WithNumber("person_id", gomcp.Description("Filter by person ID")),
			gomcp.WithBoolean("overdue", gomcp.Description("Show only overdue tasks")),
			gomcp.WithBoolean("include_completed", gomcp.Description("Include completed tasks")),
			gomcp.WithNumber("limit", gomcp.Description("Max results")),
		),
		taskListHandler(tr),
	)

	// Tag tools
	s.AddTool(
		gomcp.NewTool("crm_tag_apply",
			gomcp.WithDescription("Apply a tag to an entity"),
			gomcp.WithString("entity_type", gomcp.Required(), gomcp.Description("Entity type: person, organization, deal, interaction")),
			gomcp.WithNumber("entity_id", gomcp.Required(), gomcp.Description("Entity ID")),
			gomcp.WithString("tag", gomcp.Required(), gomcp.Description("Tag name")),
		),
		tagApplyHandler(tagr),
	)

	// Relationships
	rr := repo.NewRelationshipRepo(db)
	s.AddTool(
		gomcp.NewTool("crm_person_relate",
			gomcp.WithDescription("Create a relationship between two people"),
			gomcp.WithNumber("person_id", gomcp.Required(), gomcp.Description("Person ID")),
			gomcp.WithNumber("related_person_id", gomcp.Required(), gomcp.Description("Related person ID")),
			gomcp.WithString("type", gomcp.Required(), gomcp.Description("Relationship type: colleague, friend, manager, mentor, referred-by")),
			gomcp.WithString("notes", gomcp.Description("Optional notes about the relationship")),
		),
		relateHandler(rr),
	)

	// Stats
	s.AddTool(
		gomcp.NewTool("crm_stats",
			gomcp.WithDescription("CRM summary statistics"),
		),
		statsHandler(db),
	)

	return s
}

// Helper to convert any value to JSON text result
func jsonResult(v any) (*gomcp.CallToolResult, error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return gomcp.NewToolResultError(fmt.Sprintf("marshal error: %s", err)), nil
	}
	return gomcp.NewToolResultText(string(b)), nil
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// Safe type assertion helpers for MCP arguments.

func argString(args map[string]any, key string) (string, bool) {
	v, ok := args[key]
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}

func argFloat(args map[string]any, key string) (float64, bool) {
	v, ok := args[key]
	if !ok {
		return 0, false
	}
	switch n := v.(type) {
	case float64:
		return n, true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	default:
		return 0, false
	}
}

func argInt64(args map[string]any, key string) (int64, bool) {
	f, ok := argFloat(args, key)
	if !ok {
		return 0, false
	}
	return int64(f), true
}

// --- Person handlers ---

func personSearchHandler(pr *repo.PersonRepo) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		query := req.GetString("query", "")
		limit := req.GetInt("limit", 20)
		people, err := pr.Search(ctx, query, limit)
		if err != nil {
			return gomcp.NewToolResultError(err.Error()), nil
		}
		return jsonResult(people)
	}
}

func personGetHandler(pr *repo.PersonRepo) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		id := int64(req.GetInt("id", 0))
		person, err := pr.FindByID(ctx, id)
		if err != nil {
			return gomcp.NewToolResultError(err.Error()), nil
		}
		return jsonResult(person)
	}
}

func personCreateHandler(pr *repo.PersonRepo) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		input := model.CreatePersonInput{
			FirstName: req.GetString("first_name", ""),
			LastName:  strPtr(req.GetString("last_name", "")),
			Email:     strPtr(req.GetString("email", "")),
			Phone:     strPtr(req.GetString("phone", "")),
			Title:     strPtr(req.GetString("title", "")),
			Company:   strPtr(req.GetString("company", "")),
			Location:  strPtr(req.GetString("location", "")),
			Notes:     strPtr(req.GetString("notes", "")),
		}
		person, err := pr.Create(ctx, input)
		if err != nil {
			return gomcp.NewToolResultError(err.Error()), nil
		}
		return jsonResult(person)
	}
}

func personUpdateHandler(pr *repo.PersonRepo) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		id := int64(req.GetInt("id", 0))
		args := req.GetArguments()
		input := model.UpdatePersonInput{}

		if s, ok := argString(args, "first_name"); ok {
			input.FirstName = &s
		}
		if s, ok := argString(args, "last_name"); ok {
			input.LastName = &s
		}
		if s, ok := argString(args, "email"); ok {
			input.Email = &s
		}
		if s, ok := argString(args, "phone"); ok {
			input.Phone = &s
		}
		if s, ok := argString(args, "title"); ok {
			input.Title = &s
		}
		if s, ok := argString(args, "company"); ok {
			input.Company = &s
		}
		if s, ok := argString(args, "location"); ok {
			input.Location = &s
		}
		if s, ok := argString(args, "notes"); ok {
			input.Notes = &s
		}
		if s, ok := argString(args, "summary"); ok {
			input.Summary = &s
		}

		person, err := pr.Update(ctx, id, input)
		if err != nil {
			return gomcp.NewToolResultError(err.Error()), nil
		}
		return jsonResult(person)
	}
}

func personDeleteHandler(pr *repo.PersonRepo) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		id := int64(req.GetInt("id", 0))
		err := pr.Archive(ctx, id)
		if err != nil {
			return gomcp.NewToolResultError(err.Error()), nil
		}
		return gomcp.NewToolResultText(fmt.Sprintf("Person #%d deleted", id)), nil
	}
}

// --- Organization handlers ---

func orgSearchHandler(or *repo.OrgRepo) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		query := req.GetString("query", "")
		limit := req.GetInt("limit", 20)
		orgs, err := or.Search(ctx, query, limit)
		if err != nil {
			return gomcp.NewToolResultError(err.Error()), nil
		}
		return jsonResult(orgs)
	}
}

func orgGetHandler(or *repo.OrgRepo) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		id := int64(req.GetInt("id", 0))
		org, err := or.FindByID(ctx, id)
		if err != nil {
			return gomcp.NewToolResultError(err.Error()), nil
		}
		return jsonResult(org)
	}
}

// --- Interaction handlers ---

func interactionLogHandler(ir *repo.InteractionRepo) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		args := req.GetArguments()
		input := model.CreateInteractionInput{
			Type:    req.GetString("type", "note"),
			Subject: strPtr(req.GetString("subject", "")),
			Content: strPtr(req.GetString("content", "")),
		}

		if dir := req.GetString("direction", ""); dir != "" {
			input.Direction = &dir
		}
		if at := req.GetString("occurred_at", ""); at != "" {
			input.OccurredAt = &at
		}

		// Handle person_ids as an array of numbers
		if ids, ok := args["person_ids"]; ok {
			if v, ok := ids.([]any); ok {
				for _, item := range v {
					switch n := item.(type) {
					case float64:
						input.PersonIDs = append(input.PersonIDs, int64(n))
					case int:
						input.PersonIDs = append(input.PersonIDs, int64(n))
					case int64:
						input.PersonIDs = append(input.PersonIDs, n)
					}
				}
			}
		}

		interaction, err := ir.Create(ctx, input)
		if err != nil {
			return gomcp.NewToolResultError(err.Error()), nil
		}
		return jsonResult(interaction)
	}
}

func interactionListHandler(ir *repo.InteractionRepo) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		filters := model.InteractionFilters{
			Limit: req.GetInt("limit", 20),
		}
		if pid := int64(req.GetInt("person_id", 0)); pid > 0 {
			filters.PersonID = &pid
		}
		if t := req.GetString("type", ""); t != "" {
			filters.Type = &t
		}

		interactions, err := ir.FindAll(ctx, filters)
		if err != nil {
			return gomcp.NewToolResultError(err.Error()), nil
		}
		return jsonResult(interactions)
	}
}

// --- Search handler ---

func searchHandler(pr *repo.PersonRepo, or *repo.OrgRepo, ir *repo.InteractionRepo, dr *repo.DealRepo) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		query := req.GetString("query", "")
		entityType := req.GetString("type", "")
		limit := req.GetInt("limit", 20)

		results := map[string]any{}

		if entityType == "" || entityType == "person" {
			people, err := pr.Search(ctx, query, limit)
			if err != nil {
				return gomcp.NewToolResultError(err.Error()), nil
			}
			results["people"] = people
		}
		if entityType == "" || entityType == "organization" {
			orgs, err := or.Search(ctx, query, limit)
			if err != nil {
				return gomcp.NewToolResultError(err.Error()), nil
			}
			results["organizations"] = orgs
		}
		if entityType == "" || entityType == "interaction" {
			interactions, err := ir.Search(ctx, query, limit)
			if err != nil {
				return gomcp.NewToolResultError(err.Error()), nil
			}
			results["interactions"] = interactions
		}
		if entityType == "" || entityType == "deal" {
			deals, err := dr.Search(ctx, query, limit)
			if err != nil {
				return gomcp.NewToolResultError(err.Error()), nil
			}
			results["deals"] = deals
		}

		return jsonResult(results)
	}
}

// --- Context handler ---

func contextHandler(pr *repo.PersonRepo, or *repo.OrgRepo, ir *repo.InteractionRepo, dr *repo.DealRepo, tr *repo.TaskRepo, rr *repo.RelationshipRepo, tagr *repo.TagRepo) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		id := int64(req.GetInt("person_id", 0))
		person, err := pr.FindByID(ctx, id)
		if err != nil {
			return gomcp.NewToolResultError(err.Error()), nil
		}

		result := map[string]any{"person": person}

		if person.OrgID != nil {
			org, err := or.FindByID(ctx, *person.OrgID)
			if err != nil {
				return gomcp.NewToolResultError(err.Error()), nil
			}
			result["organization"] = org
		}

		interactions, err := ir.FindAll(ctx, model.InteractionFilters{PersonID: &person.ID, Limit: 10})
		if err != nil {
			return gomcp.NewToolResultError(err.Error()), nil
		}
		if len(interactions) > 0 {
			result["recent_interactions"] = interactions
		}

		deals, err := dr.FindAll(ctx, model.DealFilters{PersonID: &person.ID})
		if err != nil {
			return gomcp.NewToolResultError(err.Error()), nil
		}
		if len(deals) > 0 {
			result["deals"] = deals
		}

		tasks, err := tr.FindAll(ctx, model.TaskFilters{PersonID: &person.ID})
		if err != nil {
			return gomcp.NewToolResultError(err.Error()), nil
		}
		if len(tasks) > 0 {
			result["tasks"] = tasks
		}

		rels, err := rr.FindForPerson(ctx, person.ID)
		if err != nil {
			return gomcp.NewToolResultError(err.Error()), nil
		}
		if len(rels) > 0 {
			result["relationships"] = rels
		}

		tags, err := tagr.GetForEntity(ctx, "person", person.ID)
		if err != nil {
			return gomcp.NewToolResultError(err.Error()), nil
		}
		if len(tags) > 0 {
			result["tags"] = tags
		}

		return jsonResult(result)
	}
}

// --- Deal handlers ---

func dealCreateHandler(dr *repo.DealRepo) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		args := req.GetArguments()
		input := model.CreateDealInput{
			Title: req.GetString("title", ""),
			Stage: req.GetString("stage", "lead"),
			Notes: strPtr(req.GetString("notes", "")),
		}

		if f, ok := argFloat(args, "value"); ok {
			input.Value = &f
		}
		if pid, ok := argInt64(args, "person_id"); ok {
			input.PersonID = &pid
		}
		if oid, ok := argInt64(args, "org_id"); ok {
			input.OrgID = &oid
		}

		deal, err := dr.Create(ctx, input)
		if err != nil {
			return gomcp.NewToolResultError(err.Error()), nil
		}
		return jsonResult(deal)
	}
}

func dealUpdateHandler(dr *repo.DealRepo) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		id := int64(req.GetInt("id", 0))
		args := req.GetArguments()
		input := model.UpdateDealInput{}

		if s, ok := argString(args, "title"); ok {
			input.Title = &s
		}
		if f, ok := argFloat(args, "value"); ok {
			input.Value = &f
		}
		if s, ok := argString(args, "stage"); ok {
			input.Stage = &s
		}
		if s, ok := argString(args, "notes"); ok {
			input.Notes = &s
		}
		if s, ok := argString(args, "closed_at"); ok {
			input.ClosedAt = &s
		}

		deal, err := dr.Update(ctx, id, input)
		if err != nil {
			return gomcp.NewToolResultError(err.Error()), nil
		}
		return jsonResult(deal)
	}
}

// --- Task handlers ---

func taskCreateHandler(tr *repo.TaskRepo) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		args := req.GetArguments()
		input := model.CreateTaskInput{
			Title:       req.GetString("title", ""),
			Description: strPtr(req.GetString("description", "")),
			Priority:    req.GetString("priority", "medium"),
			DueAt:       strPtr(req.GetString("due", "")),
		}

		if pid, ok := argInt64(args, "person_id"); ok {
			input.PersonID = &pid
		}
		if did, ok := argInt64(args, "deal_id"); ok {
			input.DealID = &did
		}

		task, err := tr.Create(ctx, input)
		if err != nil {
			return gomcp.NewToolResultError(err.Error()), nil
		}
		return jsonResult(task)
	}
}

func taskListHandler(tr *repo.TaskRepo) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		filters := model.TaskFilters{
			Limit:            req.GetInt("limit", 0),
			Overdue:          req.GetBool("overdue", false),
			IncludeCompleted: req.GetBool("include_completed", false),
		}

		if pid := int64(req.GetInt("person_id", 0)); pid > 0 {
			filters.PersonID = &pid
		}

		tasks, err := tr.FindAll(ctx, filters)
		if err != nil {
			return gomcp.NewToolResultError(err.Error()), nil
		}
		return jsonResult(tasks)
	}
}

// --- Tag handler ---

func tagApplyHandler(tagr *repo.TagRepo) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		entityType := req.GetString("entity_type", "")
		entityID := int64(req.GetInt("entity_id", 0))
		tagName := req.GetString("tag", "")

		err := tagr.Apply(ctx, entityType, entityID, tagName)
		if err != nil {
			return gomcp.NewToolResultError(err.Error()), nil
		}
		return gomcp.NewToolResultText(fmt.Sprintf("Tagged %s #%d with %q", entityType, entityID, tagName)), nil
	}
}

// --- Relationship handler ---

func relateHandler(rr *repo.RelationshipRepo) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		personID := int64(req.GetInt("person_id", 0))
		relatedID := int64(req.GetInt("related_person_id", 0))
		relType := req.GetString("type", "")
		notes := req.GetString("notes", "")

		var notesPtr *string
		if notes != "" {
			notesPtr = &notes
		}

		rel, err := rr.Create(ctx, personID, relatedID, relType, notesPtr)
		if err != nil {
			return gomcp.NewToolResultError(err.Error()), nil
		}
		return jsonResult(rel)
	}
}

// --- Stats handler ---

func statsHandler(db *sql.DB) server.ToolHandlerFunc {
	return func(ctx context.Context, req gomcp.CallToolRequest) (*gomcp.CallToolResult, error) {
		stats := map[string]any{}

		var n int
		_ = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM people WHERE archived = 0").Scan(&n)
		stats["contacts"] = n

		_ = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM organizations WHERE archived = 0").Scan(&n)
		stats["organizations"] = n

		var dealCount int
		var dealValue float64
		_ = db.QueryRowContext(ctx,
			"SELECT COUNT(*), COALESCE(SUM(value), 0) FROM deals WHERE archived = 0 AND stage NOT IN ('won', 'lost')").
			Scan(&dealCount, &dealValue)
		stats["open_deals"] = dealCount
		stats["deal_value"] = dealValue

		_ = db.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM tasks WHERE archived = 0 AND completed = 0").Scan(&n)
		stats["open_tasks"] = n

		_ = db.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM tasks WHERE archived = 0 AND completed = 0 AND due_at IS NOT NULL AND due_at < datetime('now')").Scan(&n)
		stats["overdue_tasks"] = n

		_ = db.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM interactions WHERE archived = 0 AND occurred_at >= datetime('now', '-7 days')").Scan(&n)
		stats["interactions_7days"] = n

		return jsonResult(stats)
	}
}
