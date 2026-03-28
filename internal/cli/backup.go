package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/jdanielnd/crm-cli/internal/backup"
	"github.com/spf13/cobra"
)

func registerBackupCommands(rootCmd *cobra.Command) {
	backupCmd := &cobra.Command{
		Use:   "backup",
		Short: "Backup CRM data to CSV or Google Sheets",
	}

	backupCmd.AddCommand(backupCSVCmd())
	backupCmd.AddCommand(backupSheetsCmd())
	backupCmd.AddCommand(backupPullCmd())

	rootCmd.AddCommand(backupCmd)
}

func backupCSVCmd() *cobra.Command {
	var dir string

	cmd := &cobra.Command{
		Use:   "csv",
		Short: "Export CRM data as CSV files",
		RunE: func(cmd *cobra.Command, args []string) error {
			if dir == "" {
				dir = fmt.Sprintf("crm-backup-%s", time.Now().Format("2006-01-02"))
			}

			db, err := openDB()
			if err != nil {
				return err
			}
			defer db.Close()

			totalRows, err := backup.ExportCSV(cmd.Context(), db, dir)
			if err != nil {
				return err
			}

			fmt.Fprintf(os.Stderr, "Exported %d rows to %s/\n", totalRows, dir)
			return nil
		},
	}

	cmd.Flags().StringVar(&dir, "dir", "", "output directory (default: crm-backup-YYYY-MM-DD)")
	return cmd
}

func backupSheetsCmd() *cobra.Command {
	var sheetID, credentials string

	cmd := &cobra.Command{
		Use:   "sheets",
		Short: "Backup CRM data to Google Sheets",
		Long:  "Export all CRM data to a Google Spreadsheet. With --sheet-id, syncs to an existing sheet (local data is source of truth). Without it, creates a new spreadsheet.",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Resolve credentials path
			if credentials == "" {
				credentials = os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
			}
			if credentials == "" {
				return fmt.Errorf("--credentials flag or GOOGLE_APPLICATION_CREDENTIALS env var is required")
			}

			db, err := openDB()
			if err != nil {
				return err
			}
			defer db.Close()

			// Extract all CRM data
			data, err := backup.ExportAll(cmd.Context(), db)
			if err != nil {
				return err
			}

			// Create Sheets client
			client, err := backup.NewSheetsClient(cmd.Context(), credentials)
			if err != nil {
				return err
			}

			if sheetID != "" {
				// Sync mode: read existing, diff, overwrite
				fmt.Fprintf(os.Stderr, "Reading existing spreadsheet...\n")
				remote, err := client.ReadSpreadsheet(cmd.Context(), sheetID)
				if err != nil {
					return fmt.Errorf("read spreadsheet: %w", err)
				}

				// Show diff
				diffs := backup.DiffSheets(data, remote)
				for _, d := range diffs {
					fmt.Fprintf(os.Stderr, "  %s\n", d)
				}

				// Overwrite with local data
				fmt.Fprintf(os.Stderr, "Syncing local data to sheet...\n")
				if err := client.UpdateSpreadsheet(cmd.Context(), sheetID, data); err != nil {
					return err
				}

				fmt.Fprintf(os.Stderr, "Sync complete.\n")
			} else {
				// Create new spreadsheet
				title := fmt.Sprintf("CRM Backup %s", time.Now().Format("2006-01-02"))
				id, url, err := client.CreateSpreadsheet(cmd.Context(), title, data)
				if err != nil {
					return err
				}

				fmt.Fprintf(os.Stderr, "Created spreadsheet: %s\n", url)
				fmt.Fprintf(os.Stderr, "Sheet ID: %s\n", id)
				fmt.Fprintf(os.Stderr, "Note: Share the spreadsheet with your Google account to access it.\n")
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&sheetID, "sheet-id", "", "existing Google Sheet ID to sync with")
	cmd.Flags().StringVar(&credentials, "credentials", "", "path to Google service account JSON (or set GOOGLE_APPLICATION_CREDENTIALS)")
	return cmd
}

func backupPullCmd() *cobra.Command {
	var sheetID, clientID, clientSecret, refreshToken string
	var yes bool

	cmd := &cobra.Command{
		Use:   "pull",
		Short: "Pull changes from a Google Sheet into the local CRM",
		Long: `Read an existing Google Sheet, compare against local data, and apply
changes back to the local CRM database. All incoming data is sanitized
and validated before import. The local DB is auto-backed up before changes
are applied.

Only modifications to existing People, Organizations, Deals, and Tasks rows
are imported. New rows, deleted rows, and Interactions/Tags are ignored.

Auth uses OAuth2 (client_id/client_secret/refresh_token). Use with scrt:
  scrt run 'crm backup pull --sheet-id <id> --client-id $env[...] ...'`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if sheetID == "" {
				return fmt.Errorf("--sheet-id is required")
			}

			// Resolve OAuth2 credentials from flags or env
			if clientID == "" {
				clientID = os.Getenv("GOOGLE_CLIENT_ID")
			}
			if clientSecret == "" {
				clientSecret = os.Getenv("GOOGLE_CLIENT_SECRET")
			}
			if refreshToken == "" {
				refreshToken = os.Getenv("GOOGLE_REFRESH_TOKEN")
			}
			if clientID == "" || clientSecret == "" || refreshToken == "" {
				return fmt.Errorf("OAuth2 credentials required: --client-id, --client-secret, --refresh-token (or GOOGLE_CLIENT_ID, GOOGLE_CLIENT_SECRET, GOOGLE_REFRESH_TOKEN env vars)")
			}

			// Open local DB
			db, err := openDB()
			if err != nil {
				return err
			}
			defer db.Close()

			// Create OAuth2 Sheets client
			fmt.Fprintf(os.Stderr, "Authenticating with Google Sheets...\n")
			client, err := backup.NewSheetsClientOAuth2(cmd.Context(), clientID, clientSecret, refreshToken)
			if err != nil {
				return err
			}

			// Read remote sheet
			fmt.Fprintf(os.Stderr, "Reading remote spreadsheet...\n")
			remote, err := client.ReadSpreadsheet(cmd.Context(), sheetID)
			if err != nil {
				return fmt.Errorf("read spreadsheet: %w", err)
			}

			// Export local data
			fmt.Fprintf(os.Stderr, "Exporting local data for comparison...\n")
			local, err := backup.ExportAll(cmd.Context(), db)
			if err != nil {
				return err
			}

			// Compute field-level diff (remote changes vs local)
			changes := backup.DiffSheetsDetailed(local, remote)
			if len(changes) == 0 {
				fmt.Fprintf(os.Stderr, "No changes detected. Local data matches the sheet.\n")
				return nil
			}

			// Sanitize all incoming values
			fmt.Fprintf(os.Stderr, "\nSanitizing %d incoming changes...\n", len(changes))
			changes, report := backup.SanitizeChanges(changes)

			// Print sanitization warnings
			for _, w := range report.Warnings {
				fmt.Fprintf(os.Stderr, "  WARNING: %s\n", w)
			}
			if report.RejectedRows > 0 {
				fmt.Fprintf(os.Stderr, "  %d changes rejected due to validation errors\n", report.RejectedRows)
			}

			if len(changes) == 0 {
				fmt.Fprintf(os.Stderr, "No valid changes remaining after sanitization.\n")
				return nil
			}

			// Print detailed changes
			fmt.Fprintf(os.Stderr, "\nChanges to apply (%d):\n", len(changes))
			for _, c := range changes {
				fmt.Fprintf(os.Stderr, "  %s\n", c)
			}

			// Confirm unless --yes
			if !yes {
				fmt.Fprintf(os.Stderr, "\nApply %d changes to local CRM? [y/N] ", len(changes))
				reader := bufio.NewReader(os.Stdin)
				answer, err := reader.ReadString('\n')
				if err != nil && err != io.EOF {
					return fmt.Errorf("read confirmation: %w", err)
				}
				answer = strings.TrimSpace(strings.ToLower(answer))
				if answer != "y" && answer != "yes" {
					fmt.Fprintf(os.Stderr, "Aborted.\n")
					return nil
				}
			}

			// Auto-backup the database before applying
			dbPath := flagDB
			if dbPath == "" {
				dbPath = os.Getenv("CRM_DB")
			}
			if dbPath == "" {
				home, _ := os.UserHomeDir()
				dbPath = home + "/.crm/crm.db"
			}
			bakPath := fmt.Sprintf("%s.bak-%s", dbPath, time.Now().Format("20060102-150405"))
			if err := copyFile(dbPath, bakPath); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: could not backup DB: %v\n", err)
			} else {
				fmt.Fprintf(os.Stderr, "Backed up database to %s\n", bakPath)
			}

			// Apply changes
			fmt.Fprintf(os.Stderr, "Applying changes...\n")
			result := backup.ApplyChanges(cmd.Context(), db, changes)

			fmt.Fprintf(os.Stderr, "\nDone: %d applied, %d skipped\n", result.Applied, result.Skipped)
			for _, e := range result.Errors {
				fmt.Fprintf(os.Stderr, "  ERROR: %s\n", e)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&sheetID, "sheet-id", "", "Google Sheet ID to pull from (required)")
	cmd.Flags().StringVar(&clientID, "client-id", "", "Google OAuth2 client ID (or GOOGLE_CLIENT_ID env)")
	cmd.Flags().StringVar(&clientSecret, "client-secret", "", "Google OAuth2 client secret (or GOOGLE_CLIENT_SECRET env)")
	cmd.Flags().StringVar(&refreshToken, "refresh-token", "", "Google OAuth2 refresh token (or GOOGLE_REFRESH_TOKEN env)")
	cmd.Flags().BoolVar(&yes, "yes", false, "skip confirmation prompt")
	return cmd
}

// copyFile copies src to dst. Used for pre-import DB backup.
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
