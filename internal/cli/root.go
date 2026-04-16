package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/FerdiKT/revenuecat-cli/internal/buildinfo"
	"github.com/FerdiKT/revenuecat-cli/internal/config"
	"github.com/FerdiKT/revenuecat-cli/internal/exitcode"
	"github.com/spf13/cobra"
)

type App struct {
	store       *config.Store
	globalFlags GlobalFlags
}

type GlobalFlags struct {
	ContextAlias string
	AllContexts  bool
	JSON         bool
	Format       string
	Retry        bool
}

func Execute() int {
	cmd, app := newRootCommand()
	if err := cmd.Execute(); err != nil {
		var cliErr *CLIError
		if errors.As(err, &cliErr) {
			if cliErr.Message != "" {
				fmt.Fprintln(os.Stderr, cliErr.Message)
			}
			return cliErr.Code
		}
		fmt.Fprintln(os.Stderr, err.Error())
		return exitcode.Internal
	}

	_ = app
	return exitcode.Success
}

func newRootCommand() (*cobra.Command, *App) {
	store, err := config.NewStore("")
	if err != nil {
		panic(err)
	}

	app := &App{store: store}
	cmd := &cobra.Command{
		Use:           "revenuecat",
		Short:         "Agent-first RevenueCat CLI",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	cmd.Version = buildinfo.Version

	flags := cmd.PersistentFlags()
	flags.StringVar(&app.globalFlags.ContextAlias, "context", "", "Override the active context alias")
	flags.BoolVar(&app.globalFlags.AllContexts, "all-contexts", false, "Fan out read-only commands across all contexts")
	flags.BoolVar(&app.globalFlags.JSON, "json", false, "Force JSON output")
	flags.StringVar(&app.globalFlags.Format, "format", "", "Output format for read commands (json|table)")
	flags.BoolVar(&app.globalFlags.Retry, "retry", false, "Enable retries for mutating commands")

	addContextCommands(cmd, app)
	addAuthCommands(cmd, app)
	addAgentCommands(cmd)
	addResourceCommands(cmd, app)
	addVersionCommand(cmd)

	return cmd, app
}
