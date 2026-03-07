package cli

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jdanielnd/crm-cli/internal/db"
	"github.com/jdanielnd/crm-cli/internal/format"
	"github.com/spf13/cobra"
)

var (
	flagFormat  string
	flagQuiet   bool
	flagVerbose bool
	flagDB      string
	flagNoColor bool
)

// Version is set by ldflags at build time.
var Version = "dev"

// NewRootCmd creates the root cobra command.
func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "crm",
		Short: "Local-first personal CRM for the terminal",
		Long:  "A local-first personal CRM for the terminal. Store contacts, organizations, interactions, deals, and tasks in a local SQLite database.",
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       Version,
	}

	rootCmd.PersistentFlags().StringVarP(&flagFormat, "format", "f", "", "output format: table, json, csv, tsv")
	rootCmd.PersistentFlags().BoolVarP(&flagQuiet, "quiet", "q", false, "minimal output (just IDs)")
	rootCmd.PersistentFlags().BoolVarP(&flagVerbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().StringVar(&flagDB, "db", "", "alternate database path")
	rootCmd.PersistentFlags().BoolVar(&flagNoColor, "no-color", false, "disable colors")

	registerPersonCommands(rootCmd)
	registerOrgCommands(rootCmd)
	registerLogCommands(rootCmd)
	registerTagCommands(rootCmd)
	registerDealCommands(rootCmd)
	registerTaskCommands(rootCmd)

	return rootCmd
}

// openDB resolves the database path and opens it.
func openDB() (*sql.DB, error) {
	dbPath := flagDB

	if dbPath == "" {
		dbPath = os.Getenv("CRM_DB")
	}

	if dbPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("get home directory: %w", err)
		}
		dir := filepath.Join(home, ".crm")
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("create data directory: %w", err)
		}
		dbPath = filepath.Join(dir, "crm.db")
	}

	// Ensure parent directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("create directory %s: %w", dir, err)
	}

	return db.Open(dbPath)
}

// resolveFormat returns the current output format.
func resolveFormat() format.Format {
	return format.Resolve(flagFormat)
}
