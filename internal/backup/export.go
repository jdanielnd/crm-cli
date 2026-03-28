package backup

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/jdanielnd/crm-cli/internal/db/repo"
	"github.com/jdanielnd/crm-cli/internal/model"
)

// SheetData represents one sheet's worth of data (headers + rows).
type SheetData struct {
	Name    string
	Headers []string
	Rows    [][]string
}

// ExportAll extracts all CRM data as SheetData slices.
func ExportAll(ctx context.Context, db *sql.DB) ([]SheetData, error) {
	var sheets []SheetData

	people, err := exportPeople(ctx, db)
	if err != nil {
		return nil, err
	}
	sheets = append(sheets, people)

	orgs, err := exportOrgs(ctx, db)
	if err != nil {
		return nil, err
	}
	sheets = append(sheets, orgs)

	deals, err := exportDeals(ctx, db)
	if err != nil {
		return nil, err
	}
	sheets = append(sheets, deals)

	tasks, err := exportTasks(ctx, db)
	if err != nil {
		return nil, err
	}
	sheets = append(sheets, tasks)

	interactions, err := exportInteractions(ctx, db)
	if err != nil {
		return nil, err
	}
	sheets = append(sheets, interactions)

	tags, err := exportTags(ctx, db)
	if err != nil {
		return nil, err
	}
	sheets = append(sheets, tags)

	return sheets, nil
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func derefInt64(n *int64) string {
	if n == nil {
		return ""
	}
	return strconv.FormatInt(*n, 10)
}

func derefFloat64(f *float64) string {
	if f == nil {
		return ""
	}
	return strconv.FormatFloat(*f, 'f', -1, 64)
}

func boolStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

func exportPeople(ctx context.Context, db *sql.DB) (SheetData, error) {
	pr := repo.NewPersonRepo(db)
	people, err := pr.FindAll(ctx, model.PersonFilters{})
	if err != nil {
		return SheetData{}, fmt.Errorf("export people: %w", err)
	}

	sd := SheetData{
		Name:    "People",
		Headers: []string{"ID", "UUID", "FirstName", "LastName", "Email", "Phone", "Title", "Company", "Location", "Notes", "Summary", "OrgID", "CreatedAt", "UpdatedAt"},
	}
	for _, p := range people {
		sd.Rows = append(sd.Rows, []string{
			strconv.FormatInt(p.ID, 10), p.UUID, p.FirstName,
			deref(p.LastName), deref(p.Email), deref(p.Phone),
			deref(p.Title), deref(p.Company), deref(p.Location),
			deref(p.Notes), deref(p.Summary), derefInt64(p.OrgID),
			p.CreatedAt, p.UpdatedAt,
		})
	}
	return sd, nil
}

func exportOrgs(ctx context.Context, db *sql.DB) (SheetData, error) {
	or := repo.NewOrgRepo(db)
	orgs, err := or.FindAll(ctx, 0)
	if err != nil {
		return SheetData{}, fmt.Errorf("export organizations: %w", err)
	}

	sd := SheetData{
		Name:    "Organizations",
		Headers: []string{"ID", "UUID", "Name", "Domain", "Industry", "Notes", "Summary", "CreatedAt", "UpdatedAt"},
	}
	for _, o := range orgs {
		sd.Rows = append(sd.Rows, []string{
			strconv.FormatInt(o.ID, 10), o.UUID, o.Name,
			deref(o.Domain), deref(o.Industry), deref(o.Notes),
			deref(o.Summary), o.CreatedAt, o.UpdatedAt,
		})
	}
	return sd, nil
}

func exportDeals(ctx context.Context, db *sql.DB) (SheetData, error) {
	dr := repo.NewDealRepo(db)
	deals, err := dr.FindAll(ctx, model.DealFilters{})
	if err != nil {
		return SheetData{}, fmt.Errorf("export deals: %w", err)
	}

	sd := SheetData{
		Name:    "Deals",
		Headers: []string{"ID", "UUID", "Title", "Value", "Stage", "PersonID", "OrgID", "Notes", "ClosedAt", "CreatedAt", "UpdatedAt"},
	}
	for _, d := range deals {
		sd.Rows = append(sd.Rows, []string{
			strconv.FormatInt(d.ID, 10), d.UUID, d.Title,
			derefFloat64(d.Value), d.Stage, derefInt64(d.PersonID),
			derefInt64(d.OrgID), deref(d.Notes), deref(d.ClosedAt),
			d.CreatedAt, d.UpdatedAt,
		})
	}
	return sd, nil
}

func exportTasks(ctx context.Context, db *sql.DB) (SheetData, error) {
	tr := repo.NewTaskRepo(db)
	tasks, err := tr.FindAll(ctx, model.TaskFilters{IncludeCompleted: true})
	if err != nil {
		return SheetData{}, fmt.Errorf("export tasks: %w", err)
	}

	sd := SheetData{
		Name:    "Tasks",
		Headers: []string{"ID", "UUID", "Title", "Description", "PersonID", "DealID", "DueAt", "Priority", "Completed", "CompletedAt", "CreatedAt", "UpdatedAt"},
	}
	for _, t := range tasks {
		sd.Rows = append(sd.Rows, []string{
			strconv.FormatInt(t.ID, 10), t.UUID, t.Title,
			deref(t.Description), derefInt64(t.PersonID), derefInt64(t.DealID),
			deref(t.DueAt), t.Priority, boolStr(t.Completed),
			deref(t.CompletedAt), t.CreatedAt, t.UpdatedAt,
		})
	}
	return sd, nil
}

func exportInteractions(ctx context.Context, db *sql.DB) (SheetData, error) {
	ir := repo.NewInteractionRepo(db)
	interactions, err := ir.FindAll(ctx, model.InteractionFilters{})
	if err != nil {
		return SheetData{}, fmt.Errorf("export interactions: %w", err)
	}

	sd := SheetData{
		Name:    "Interactions",
		Headers: []string{"ID", "UUID", "Type", "Subject", "Content", "Direction", "OccurredAt", "PersonIDs", "CreatedAt", "UpdatedAt"},
	}
	for _, i := range interactions {
		pids := make([]string, len(i.PersonIDs))
		for j, pid := range i.PersonIDs {
			pids[j] = strconv.FormatInt(pid, 10)
		}
		sd.Rows = append(sd.Rows, []string{
			strconv.FormatInt(i.ID, 10), i.UUID, i.Type,
			deref(i.Subject), deref(i.Content), deref(i.Direction),
			i.OccurredAt, strings.Join(pids, ","),
			i.CreatedAt, i.UpdatedAt,
		})
	}
	return sd, nil
}

func exportTags(ctx context.Context, db *sql.DB) (SheetData, error) {
	tagr := repo.NewTagRepo(db)
	tags, err := tagr.FindAll(ctx)
	if err != nil {
		return SheetData{}, fmt.Errorf("export tags: %w", err)
	}

	sd := SheetData{
		Name:    "Tags",
		Headers: []string{"TagID", "TagName", "EntityType", "EntityID"},
	}

	for _, tag := range tags {
		for _, entityType := range model.EntityTypes {
			entityIDs, err := tagr.GetEntities(ctx, tag.Name, entityType)
			if err != nil {
				continue
			}
			for _, eid := range entityIDs {
				sd.Rows = append(sd.Rows, []string{
					strconv.FormatInt(tag.ID, 10), tag.Name,
					entityType, strconv.FormatInt(eid, 10),
				})
			}
		}
	}
	return sd, nil
}
