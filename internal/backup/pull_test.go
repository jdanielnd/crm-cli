package backup

import (
	"context"
	"fmt"
	"testing"

	"github.com/jdanielnd/crm-cli/internal/db"
	"github.com/jdanielnd/crm-cli/internal/db/repo"
	"github.com/jdanielnd/crm-cli/internal/model"
)

func TestPullPipeline(t *testing.T) {
	// Setup in-memory DB with test data
	d, err := db.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer d.Close()
	ctx := context.Background()

	pr := repo.NewPersonRepo(d)
	or := repo.NewOrgRepo(d)
	dr := repo.NewDealRepo(d)

	org, err := or.Create(ctx, model.CreateOrgInput{Name: "Acme Corp"})
	if err != nil {
		t.Fatal(err)
	}

	notes := "Initial notes"
	person, err := pr.Create(ctx, model.CreatePersonInput{
		FirstName: "Jane",
		LastName:  strp("Smith"),
		Email:     strp("jane@acme.com"),
		Notes:     &notes,
		OrgID:     &org.ID,
	})
	if err != nil {
		t.Fatal(err)
	}

	dealNotes := "Old deal notes"
	val := 15000.0
	deal, err := dr.Create(ctx, model.CreateDealInput{
		Title:    "Website Redesign",
		Value:    &val,
		Stage:    "proposal",
		PersonID: &person.ID,
		Notes:    &dealNotes,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Export local
	local, err := ExportAll(ctx, d)
	if err != nil {
		t.Fatal(err)
	}

	// Simulate remote with modifications
	remote := cloneSheets(local)
	setCell(remote, "People", person.ID, "Notes", "Updated: great call today")
	setCell(remote, "Deals", deal.ID, "Notes", "=EVIL_FORMULA()")

	// Diff
	changes := DiffSheetsDetailed(local, remote)
	if len(changes) != 2 {
		t.Fatalf("expected 2 changes, got %d", len(changes))
	}
	t.Logf("Changes: %d", len(changes))
	for _, c := range changes {
		t.Logf("  %s", c)
	}

	// Sanitize
	changes, report := SanitizeChanges(changes)
	t.Logf("After sanitize: %d clean, %d rejected, %d warnings", len(changes), report.RejectedRows, len(report.Warnings))
	for _, w := range report.Warnings {
		t.Logf("  WARNING: %s", w)
	}

	// The deal Notes had a formula — should be stripped but not rejected (Notes is freetext)
	if len(changes) != 2 {
		t.Fatalf("expected 2 clean changes, got %d", len(changes))
	}
	// Check formula was stripped
	for _, c := range changes {
		if c.SheetName == "Deals" && c.Column == "Notes" {
			if c.RemoteValue == "=EVIL_FORMULA()" {
				t.Fatal("formula should have been stripped")
			}
			if c.RemoteValue != "EVIL_FORMULA()" {
				t.Fatalf("expected 'EVIL_FORMULA()', got %q", c.RemoteValue)
			}
		}
	}

	// Apply
	result := ApplyChanges(ctx, d, changes)
	t.Logf("Applied: %d, Skipped: %d, Errors: %v", result.Applied, result.Skipped, result.Errors)
	if result.Applied != 2 {
		t.Fatalf("expected 2 applied, got %d", result.Applied)
	}
	if len(result.Errors) != 0 {
		t.Fatalf("expected 0 errors, got %v", result.Errors)
	}

	// Verify person was updated
	updatedPerson, err := pr.FindByID(ctx, person.ID)
	if err != nil {
		t.Fatal(err)
	}
	if updatedPerson.Notes == nil || *updatedPerson.Notes != "Updated: great call today" {
		t.Fatalf("expected updated notes, got %v", updatedPerson.Notes)
	}

	// Verify deal was updated (with stripped formula)
	updatedDeal, err := dr.FindByID(ctx, deal.ID)
	if err != nil {
		t.Fatal(err)
	}
	if updatedDeal.Notes == nil || *updatedDeal.Notes != "EVIL_FORMULA()" {
		t.Fatalf("expected stripped formula notes, got %v", updatedDeal.Notes)
	}
}

func TestSanitizeEnumRejection(t *testing.T) {
	changes := []FieldChange{
		{SheetName: "Deals", RowID: "1", Column: "Stage", RemoteValue: "invalid_stage"},
		{SheetName: "Tasks", RowID: "1", Column: "Priority", RemoteValue: "critical"},
		{SheetName: "People", RowID: "1", Column: "Notes", RemoteValue: "normal text"},
	}

	clean, report := SanitizeChanges(changes)
	if len(clean) != 1 {
		t.Fatalf("expected 1 clean change, got %d", len(clean))
	}
	if report.RejectedRows != 2 {
		t.Fatalf("expected 2 rejected, got %d", report.RejectedRows)
	}
	if clean[0].Column != "Notes" {
		t.Fatalf("expected Notes to survive, got %s", clean[0].Column)
	}
}

func TestSanitizeNullBytes(t *testing.T) {
	changes := []FieldChange{
		{SheetName: "People", RowID: "1", Column: "Notes", RemoteValue: "hello\x00world"},
	}
	clean, report := SanitizeChanges(changes)
	if len(clean) != 1 {
		t.Fatal("expected 1 clean change")
	}
	if clean[0].RemoteValue != "helloworld" {
		t.Fatalf("expected null stripped, got %q", clean[0].RemoteValue)
	}
	if len(report.Warnings) != 1 {
		t.Fatalf("expected 1 warning, got %d", len(report.Warnings))
	}
}

func TestDiffIgnoresNewAndRemovedRows(t *testing.T) {
	local := []SheetData{{
		Name:    "People",
		Headers: []string{"ID", "UUID", "FirstName"},
		Rows:    [][]string{{"1", "abc", "Jane"}},
	}}
	remote := []SheetData{{
		Name:    "People",
		Headers: []string{"ID", "UUID", "FirstName"},
		Rows: [][]string{
			{"1", "abc", "Jane"},
			{"99", "xyz", "NewPerson"}, // new row — should be ignored
		},
	}}

	changes := DiffSheetsDetailed(local, remote)
	if len(changes) != 0 {
		t.Fatalf("expected 0 changes (new rows ignored), got %d", len(changes))
	}
}

// helpers

func strp(s string) *string { return &s }

func cloneSheets(src []SheetData) []SheetData {
	dst := make([]SheetData, len(src))
	for i, s := range src {
		dst[i] = SheetData{Name: s.Name, Headers: s.Headers}
		for _, row := range s.Rows {
			rowCopy := make([]string, len(row))
			copy(rowCopy, row)
			dst[i].Rows = append(dst[i].Rows, rowCopy)
		}
	}
	return dst
}

func setCell(sheets []SheetData, sheetName string, id int64, column, value string) {
	idStr := fmt.Sprintf("%d", id)
	for si, s := range sheets {
		if s.Name != sheetName {
			continue
		}
		colIdx := -1
		for i, h := range s.Headers {
			if h == column {
				colIdx = i
				break
			}
		}
		if colIdx < 0 {
			return
		}
		for ri, row := range s.Rows {
			if len(row) > 0 && row[0] == idStr {
				sheets[si].Rows[ri][colIdx] = value
				return
			}
		}
	}
}
