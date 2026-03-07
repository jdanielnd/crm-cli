package cli

import (
	crmmcp "github.com/jdanielnd/crm-cli/internal/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"
)

func registerMCPCommands(rootCmd *cobra.Command) {
	mcpCmd := &cobra.Command{
		Use:   "mcp",
		Short: "MCP server for AI agent integration",
	}

	mcpCmd.AddCommand(&cobra.Command{
		Use:   "serve",
		Short: "Start the MCP server (stdio transport)",
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := openDB()
			if err != nil {
				return err
			}
			defer db.Close()

			s := crmmcp.NewServer(db, Version)
			return server.ServeStdio(s)
		},
	})

	rootCmd.AddCommand(mcpCmd)
}
