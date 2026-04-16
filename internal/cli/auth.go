package cli

import (
	"fmt"
	"os"

	"github.com/FerdiKT/revenuecat-cli/internal/exitcode"
	"github.com/FerdiKT/revenuecat-cli/internal/output"
	"github.com/spf13/cobra"
)

func addAuthCommands(root *cobra.Command, app *App) {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Authentication status and future OAuth surface",
	}

	cmd.AddCommand(
		newAuthStatusCommand(app),
		newAuthLoginCommand(),
		newAuthLogoutCommand(),
	)

	root.AddCommand(cmd)
}

func newAuthStatusCommand(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show current authentication mode",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := app.loadConfig()
			if err != nil {
				return err
			}

			return output.PrintJSON(os.Stdout, output.Success(nil, map[string]any{
				"mode":            "api_key",
				"active_context":  cfg.ActiveContext,
				"context_count":   len(cfg.Contexts),
				"oauth_available": false,
				"oauth_message":   "Coming soon: waiting on RevenueCat OAuth client approval from support.",
			}, output.Meta{}))
		},
	}
}

func newAuthLoginCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "login",
		Short: "OAuth login placeholder",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(os.Stderr, "OAuth login is coming soon; waiting on RevenueCat OAuth client approval from support. Use `revenuecat contexts add` with a v2 secret API key for v1.")
			return &CLIError{Code: exitcode.NotAvailable, Message: ""}
		},
	}
}

func newAuthLogoutCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "OAuth logout placeholder",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(os.Stdout, "OAuth is not enabled in v1. Manage API key contexts with `revenuecat contexts`.")
			return nil
		},
	}
}
