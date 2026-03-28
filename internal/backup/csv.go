package backup

import (
	"context"
	"database/sql"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
)

// ExportCSV writes all CRM data as CSV files to the given directory.
func ExportCSV(ctx context.Context, db *sql.DB, dir string) (int, error) {
	sheets, err := ExportAll(ctx, db)
	if err != nil {
		return 0, err
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return 0, fmt.Errorf("create directory %s: %w", dir, err)
	}

	totalRows := 0
	for _, sheet := range sheets {
		path := filepath.Join(dir, sheet.Name+".csv")
		f, err := os.Create(path)
		if err != nil {
			return 0, fmt.Errorf("create %s: %w", path, err)
		}

		w := csv.NewWriter(f)
		if err := w.Write(sheet.Headers); err != nil {
			f.Close()
			return 0, fmt.Errorf("write headers for %s: %w", sheet.Name, err)
		}

		for _, row := range sheet.Rows {
			if err := w.Write(row); err != nil {
				f.Close()
				return 0, fmt.Errorf("write row in %s: %w", sheet.Name, err)
			}
		}

		w.Flush()
		if err := w.Error(); err != nil {
			f.Close()
			return 0, fmt.Errorf("flush %s: %w", sheet.Name, err)
		}
		f.Close()

		totalRows += len(sheet.Rows)
	}

	return totalRows, nil
}
