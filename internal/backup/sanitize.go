package backup

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/jdanielnd/crm-cli/internal/model"
)

const maxCellLength = 10000

// SanitizeWarning describes a problem found during sanitization.
type SanitizeWarning struct {
	SheetName string
	RowID     string
	Column    string
	Message   string
}

func (w SanitizeWarning) String() string {
	return fmt.Sprintf("[%s] row %s, col %s: %s", w.SheetName, w.RowID, w.Column, w.Message)
}

// SanitizeReport holds all warnings and rejected rows from sanitization.
type SanitizeReport struct {
	Warnings     []SanitizeWarning
	RejectedRows int
}

// SanitizeChanges validates and sanitizes a set of FieldChanges.
// Returns the cleaned changes and a report. Changes that fail validation
// are removed from the returned slice and counted as rejected.
func SanitizeChanges(changes []FieldChange) ([]FieldChange, SanitizeReport) {
	var report SanitizeReport
	var clean []FieldChange

	for _, c := range changes {
		val := c.RemoteValue

		// Strip null bytes
		if strings.ContainsRune(val, '\x00') {
			val = strings.ReplaceAll(val, "\x00", "")
			report.Warnings = append(report.Warnings, SanitizeWarning{
				SheetName: c.SheetName, RowID: c.RowID, Column: c.Column,
				Message: "stripped null bytes",
			})
		}

		// Strip control characters (except newline, tab)
		val = strings.Map(func(r rune) rune {
			if unicode.IsControl(r) && r != '\n' && r != '\t' {
				return -1
			}
			return r
		}, val)

		// Length cap
		if len(val) > maxCellLength {
			val = val[:maxCellLength]
			report.Warnings = append(report.Warnings, SanitizeWarning{
				SheetName: c.SheetName, RowID: c.RowID, Column: c.Column,
				Message: fmt.Sprintf("truncated to %d chars", maxCellLength),
			})
		}

		// Formula injection detection — if a freetext field starts with
		// a formula trigger character, strip the leading char and warn.
		if isFormulaTrigger(val) && isFreeTextField(c.SheetName, c.Column) {
			val = val[1:]
			report.Warnings = append(report.Warnings, SanitizeWarning{
				SheetName: c.SheetName, RowID: c.RowID, Column: c.Column,
				Message: "stripped leading formula character",
			})
		}

		// Enum validation — reject row if an enum field has an invalid value
		if rejected, msg := validateEnum(c.SheetName, c.Column, val); rejected {
			report.Warnings = append(report.Warnings, SanitizeWarning{
				SheetName: c.SheetName, RowID: c.RowID, Column: c.Column,
				Message: msg,
			})
			report.RejectedRows++
			continue
		}

		c.RemoteValue = val
		clean = append(clean, c)
	}

	return clean, report
}

func isFormulaTrigger(s string) bool {
	if len(s) == 0 {
		return false
	}
	switch s[0] {
	case '=', '+', '@':
		return true
	case '-':
		// Only flag if followed by non-digit (to allow negative numbers)
		return len(s) > 1 && (s[1] < '0' || s[1] > '9')
	}
	return false
}

// isFreeTextField returns true for columns that accept arbitrary user text
// (where formula injection is a concern). ID, UUID, enum, and timestamp
// columns are not free text.
func isFreeTextField(sheetName, column string) bool {
	switch column {
	case "ID", "UUID", "CreatedAt", "UpdatedAt", "OrgID", "PersonID",
		"DealID", "TagID", "EntityID", "Completed", "CompletedAt",
		"PersonIDs", "OccurredAt", "ClosedAt", "Value":
		return false
	case "Stage", "Priority", "Direction", "Type", "EntityType":
		return false
	}
	return true
}

func validateEnum(sheetName, column, value string) (rejected bool, msg string) {
	if value == "" {
		return false, ""
	}

	switch {
	case sheetName == "Deals" && column == "Stage":
		if !model.ValidDealStage(value) {
			return true, fmt.Sprintf("invalid stage %q (allowed: %s)", value, strings.Join(model.DealStages, ", "))
		}
	case sheetName == "Tasks" && column == "Priority":
		if !model.ValidPriority(value) {
			return true, fmt.Sprintf("invalid priority %q (allowed: %s)", value, strings.Join(model.Priorities, ", "))
		}
	case sheetName == "Interactions" && column == "Direction":
		if !model.ValidInteractionDirection(value) {
			return true, fmt.Sprintf("invalid direction %q (allowed: %s)", value, strings.Join(model.InteractionDirections, ", "))
		}
	case sheetName == "Interactions" && column == "Type":
		if !model.ValidInteractionType(value) {
			return true, fmt.Sprintf("invalid type %q (allowed: %s)", value, strings.Join(model.InteractionTypes, ", "))
		}
	}

	return false, ""
}
