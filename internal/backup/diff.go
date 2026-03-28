package backup

import "fmt"

// DiffResult summarizes changes between local and remote data for one sheet.
type DiffResult struct {
	SheetName string
	Added     int
	Modified  int
	Removed   int
	Unchanged int
}

// String formats a DiffResult as a human-readable summary.
func (d DiffResult) String() string {
	return fmt.Sprintf("%s: %d added, %d modified, %d removed, %d unchanged",
		d.SheetName, d.Added, d.Modified, d.Removed, d.Unchanged)
}

// DiffSheets compares local SheetData against remote SheetData.
// Matching is by first column (ID). All values are compared as strings.
func DiffSheets(local, remote []SheetData) []DiffResult {
	remoteMap := make(map[string]SheetData, len(remote))
	for _, r := range remote {
		remoteMap[r.Name] = r
	}

	var results []DiffResult
	for _, l := range local {
		r, hasRemote := remoteMap[l.Name]
		if !hasRemote {
			results = append(results, DiffResult{
				SheetName: l.Name,
				Added:     len(l.Rows),
			})
			continue
		}

		results = append(results, diffSheet(l, r))
	}
	return results
}

func diffSheet(local, remote SheetData) DiffResult {
	result := DiffResult{SheetName: local.Name}

	// Build map of remote rows keyed by first column (ID)
	remoteRows := make(map[string][]string, len(remote.Rows))
	for _, row := range remote.Rows {
		if len(row) > 0 {
			remoteRows[row[0]] = row
		}
	}

	// Compare local against remote
	for _, localRow := range local.Rows {
		if len(localRow) == 0 {
			continue
		}
		key := localRow[0]
		remoteRow, exists := remoteRows[key]
		if !exists {
			result.Added++
			continue
		}

		if rowsEqual(localRow, remoteRow) {
			result.Unchanged++
		} else {
			result.Modified++
		}
		delete(remoteRows, key)
	}

	// Remaining remote rows are removals
	result.Removed = len(remoteRows)

	return result
}

func rowsEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// FieldChange describes a single cell that differs between local and remote.
type FieldChange struct {
	SheetName   string
	RowID       string // value of first column (entity ID)
	Column      string // header name
	LocalValue  string
	RemoteValue string
}

// String formats a FieldChange as a human-readable line.
func (fc FieldChange) String() string {
	local := truncateForDisplay(fc.LocalValue, 40)
	remote := truncateForDisplay(fc.RemoteValue, 40)
	return fmt.Sprintf("[%s] #%s %s: %q → %q", fc.SheetName, fc.RowID, fc.Column, local, remote)
}

// DiffSheetsDetailed compares remote sheet data against local and returns
// individual field-level changes. Only existing rows (matched by ID column)
// with modified cells are returned. New/removed rows are ignored — this is
// for pulling edits, not additions or deletions.
func DiffSheetsDetailed(local, remote []SheetData) []FieldChange {
	localMap := make(map[string]SheetData, len(local))
	for _, l := range local {
		localMap[l.Name] = l
	}

	var changes []FieldChange
	for _, r := range remote {
		l, exists := localMap[r.Name]
		if !exists {
			continue // sheet doesn't exist locally, skip
		}

		changes = append(changes, diffSheetDetailed(l, r)...)
	}
	return changes
}

func diffSheetDetailed(local, remote SheetData) []FieldChange {
	// Build header index for the local sheet
	localHeaders := make(map[int]string, len(local.Headers))
	for i, h := range local.Headers {
		localHeaders[i] = h
	}

	// Remote header index → local header name mapping
	remoteColToHeader := make(map[int]string, len(remote.Headers))
	for i, h := range remote.Headers {
		remoteColToHeader[i] = h
	}

	// Build map of local rows by ID (first column)
	localRows := make(map[string][]string, len(local.Rows))
	for _, row := range local.Rows {
		if len(row) > 0 {
			localRows[row[0]] = row
		}
	}

	// Build header-to-index map for local data
	localHeaderIdx := make(map[string]int, len(local.Headers))
	for i, h := range local.Headers {
		localHeaderIdx[h] = i
	}

	var changes []FieldChange
	for _, remoteRow := range remote.Rows {
		if len(remoteRow) == 0 {
			continue
		}

		rowID := remoteRow[0]
		localRow, exists := localRows[rowID]
		if !exists {
			continue // new row in remote, skip
		}

		// Compare each cell
		for ci, remoteVal := range remoteRow {
			header, ok := remoteColToHeader[ci]
			if !ok {
				continue
			}

			// Skip ID, UUID, timestamps — these are not user-editable
			if header == "ID" || header == "UUID" || header == "CreatedAt" || header == "UpdatedAt" {
				continue
			}

			// Find corresponding local value
			localIdx, ok := localHeaderIdx[header]
			if !ok {
				continue
			}
			localVal := ""
			if localIdx < len(localRow) {
				localVal = localRow[localIdx]
			}

			if remoteVal != localVal {
				changes = append(changes, FieldChange{
					SheetName:   local.Name,
					RowID:       rowID,
					Column:      header,
					LocalValue:  localVal,
					RemoteValue: remoteVal,
				})
			}
		}
	}

	return changes
}

func truncateForDisplay(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
