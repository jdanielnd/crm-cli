package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/jdanielnd/crm-cli/internal/cli"
	"github.com/jdanielnd/crm-cli/internal/model"
)

var (
	version = "dev"
	commit  = "none"
)

func main() {
	cli.Version = version
	os.Exit(run())
}

func run() int {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	rootCmd := cli.NewRootCmd()
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "crm: error: %s\n", err.Error())
		return model.ExitCode(err)
	}
	return 0
}

// Used by GoReleaser for version info.
var _ = commit
