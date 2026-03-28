package backup

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"

	"github.com/jdanielnd/crm-cli/internal/db/repo"
	"github.com/jdanielnd/crm-cli/internal/model"
)

// ApplyResult tracks what happened during import.
type ApplyResult struct {
	Applied  int
	Skipped  int
	Errors   []string
}

// ApplyChanges writes validated FieldChanges back to the local database.
// Only People, Organizations, Deals, and Tasks are writable.
// Interactions and Tags are read-only from sheets.
func ApplyChanges(ctx context.Context, db *sql.DB, changes []FieldChange) ApplyResult {
	var result ApplyResult

	// Group changes by (SheetName, RowID) so we can batch updates per entity
	type entityKey struct {
		sheet string
		id    int64
	}
	grouped := make(map[entityKey][]FieldChange)

	for _, c := range changes {
		// Reject read-only sheets
		if c.SheetName == "Interactions" || c.SheetName == "Tags" {
			result.Skipped++
			continue
		}

		id, err := strconv.ParseInt(c.RowID, 10, 64)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("[%s] invalid ID %q", c.SheetName, c.RowID))
			continue
		}

		key := entityKey{sheet: c.SheetName, id: id}
		grouped[key] = append(grouped[key], c)
	}

	pr := repo.NewPersonRepo(db)
	or := repo.NewOrgRepo(db)
	dr := repo.NewDealRepo(db)
	tr := repo.NewTaskRepo(db)

	for key, fields := range grouped {
		var err error
		switch key.sheet {
		case "People":
			err = applyPerson(ctx, pr, key.id, fields)
		case "Organizations":
			err = applyOrg(ctx, or, key.id, fields)
		case "Deals":
			err = applyDeal(ctx, dr, key.id, fields)
		case "Tasks":
			err = applyTask(ctx, tr, key.id, fields)
		default:
			result.Skipped += len(fields)
			continue
		}

		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("[%s] #%d: %v", key.sheet, key.id, err))
		} else {
			result.Applied += len(fields)
		}
	}

	return result
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func int64Ptr(s string) *int64 {
	if s == "" {
		return nil
	}
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return nil
	}
	return &v
}

func float64Ptr(s string) *float64 {
	if s == "" {
		return nil
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil
	}
	return &v
}

func applyPerson(ctx context.Context, pr *repo.PersonRepo, id int64, fields []FieldChange) error {
	input := model.UpdatePersonInput{}
	for _, f := range fields {
		switch f.Column {
		case "FirstName":
			input.FirstName = &f.RemoteValue
		case "LastName":
			input.LastName = strPtr(f.RemoteValue)
		case "Email":
			input.Email = strPtr(f.RemoteValue)
		case "Phone":
			input.Phone = strPtr(f.RemoteValue)
		case "Title":
			input.Title = strPtr(f.RemoteValue)
		case "Company":
			input.Company = strPtr(f.RemoteValue)
		case "Location":
			input.Location = strPtr(f.RemoteValue)
		case "Notes":
			input.Notes = strPtr(f.RemoteValue)
		case "Summary":
			input.Summary = strPtr(f.RemoteValue)
		case "OrgID":
			input.OrgID = int64Ptr(f.RemoteValue)
		}
	}
	_, err := pr.Update(ctx, id, input)
	return err
}

func applyOrg(ctx context.Context, or *repo.OrgRepo, id int64, fields []FieldChange) error {
	input := model.UpdateOrgInput{}
	for _, f := range fields {
		switch f.Column {
		case "Name":
			input.Name = &f.RemoteValue
		case "Domain":
			input.Domain = strPtr(f.RemoteValue)
		case "Industry":
			input.Industry = strPtr(f.RemoteValue)
		case "Notes":
			input.Notes = strPtr(f.RemoteValue)
		case "Summary":
			input.Summary = strPtr(f.RemoteValue)
		}
	}
	_, err := or.Update(ctx, id, input)
	return err
}

func applyDeal(ctx context.Context, dr *repo.DealRepo, id int64, fields []FieldChange) error {
	input := model.UpdateDealInput{}
	for _, f := range fields {
		switch f.Column {
		case "Title":
			input.Title = &f.RemoteValue
		case "Value":
			input.Value = float64Ptr(f.RemoteValue)
		case "Stage":
			input.Stage = &f.RemoteValue
		case "PersonID":
			input.PersonID = int64Ptr(f.RemoteValue)
		case "OrgID":
			input.OrgID = int64Ptr(f.RemoteValue)
		case "Notes":
			input.Notes = strPtr(f.RemoteValue)
		case "ClosedAt":
			input.ClosedAt = strPtr(f.RemoteValue)
		}
	}
	_, err := dr.Update(ctx, id, input)
	return err
}

func applyTask(ctx context.Context, tr *repo.TaskRepo, id int64, fields []FieldChange) error {
	input := model.UpdateTaskInput{}
	for _, f := range fields {
		switch f.Column {
		case "Title":
			input.Title = &f.RemoteValue
		case "Description":
			input.Description = strPtr(f.RemoteValue)
		case "PersonID":
			input.PersonID = int64Ptr(f.RemoteValue)
		case "DealID":
			input.DealID = int64Ptr(f.RemoteValue)
		case "DueAt":
			input.DueAt = strPtr(f.RemoteValue)
		case "Priority":
			input.Priority = &f.RemoteValue
		}
	}
	_, err := tr.Update(ctx, id, input)
	return err
}
