package format

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"golang.org/x/term"
)

// Format represents an output format.
type Format string

const (
	FormatTable Format = "table"
	FormatJSON  Format = "json"
	FormatCSV   Format = "csv"
	FormatTSV   Format = "tsv"
)

// Resolve determines the output format from flag value.
// If empty, defaults to table for TTY, json for pipes.
func Resolve(flag string) Format {
	switch strings.ToLower(flag) {
	case "json":
		return FormatJSON
	case "csv":
		return FormatCSV
	case "tsv":
		return FormatTSV
	case "table":
		return FormatTable
	case "":
		if term.IsTerminal(int(os.Stdout.Fd())) {
			return FormatTable
		}
		return FormatJSON
	default:
		return FormatTable
	}
}

// ColumnDef defines a column for table output.
type ColumnDef struct {
	Header string
	Field  string
}

// Output writes data in the specified format to w.
func Output(w io.Writer, format Format, data []map[string]any, columns []ColumnDef, quiet bool) error {
	if quiet {
		return outputQuiet(w, data)
	}

	switch format {
	case FormatJSON:
		return outputJSON(w, data)
	case FormatCSV:
		return outputCSV(w, data, columns, ',')
	case FormatTSV:
		return outputCSV(w, data, columns, '\t')
	default:
		return outputTable(w, data, columns)
	}
}

func outputQuiet(w io.Writer, data []map[string]any) error {
	for _, row := range data {
		if id, ok := row["id"]; ok {
			fmt.Fprintf(w, "%v\n", id)
		}
	}
	return nil
}

func outputJSON(w io.Writer, data []map[string]any) error {
	enc := json.NewEncoder(w)
	if term.IsTerminal(int(os.Stdout.Fd())) {
		enc.SetIndent("", "  ")
	}
	return enc.Encode(data)
}

func outputTable(w io.Writer, data []map[string]any, columns []ColumnDef) error {
	if len(data) == 0 {
		fmt.Fprintln(w, "No results.")
		return nil
	}

	t := table.NewWriter()
	t.SetOutputMirror(w)
	t.SetStyle(table.StyleLight)

	header := make(table.Row, len(columns))
	for i, col := range columns {
		header[i] = col.Header
	}
	t.AppendHeader(header)

	for _, row := range data {
		vals := make(table.Row, len(columns))
		for i, col := range columns {
			v := row[col.Field]
			if v == nil {
				vals[i] = ""
			} else {
				vals[i] = fmt.Sprintf("%v", v)
			}
		}
		t.AppendRow(vals)
	}

	t.Render()
	return nil
}

func outputCSV(w io.Writer, data []map[string]any, columns []ColumnDef, sep rune) error {
	cw := csv.NewWriter(w)
	cw.Comma = sep

	headers := make([]string, len(columns))
	for i, col := range columns {
		headers[i] = col.Field
	}
	if err := cw.Write(headers); err != nil {
		return err
	}

	for _, row := range data {
		record := make([]string, len(columns))
		for i, col := range columns {
			v := row[col.Field]
			if v == nil {
				record[i] = ""
			} else {
				record[i] = fmt.Sprintf("%v", v)
			}
		}
		if err := cw.Write(record); err != nil {
			return err
		}
	}

	cw.Flush()
	return cw.Error()
}
