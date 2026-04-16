package cli

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/FerdiKT/revenuecat-cli/internal/config"
	"github.com/FerdiKT/revenuecat-cli/internal/exitcode"
	"github.com/FerdiKT/revenuecat-cli/internal/output"
	"github.com/FerdiKT/revenuecat-cli/internal/rcapi"
	"github.com/spf13/cobra"
)

func addContextCommands(root *cobra.Command, app *App) {
	cmd := &cobra.Command{
		Use:   "contexts",
		Short: "Manage API key contexts",
	}

	cmd.AddCommand(
		newContextsAddCommand(app),
		newContextsListCommand(app),
		newContextsUseCommand(app),
		newContextsShowCommand(app),
		newContextsRemoveCommand(app),
		newContextsVerifyCommand(app),
	)

	root.AddCommand(cmd)
}

func newContextsAddCommand(app *App) *cobra.Command {
	var apiKey string
	var projectID string
	var projectName string
	var apiBaseURL string
	var makeActive bool

	cmd := &cobra.Command{
		Use:   "add <alias>",
		Short: "Add or update a named context",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := app.loadConfig()
			if err != nil {
				return err
			}

			ctx := config.Context{
				Alias:       args[0],
				APIKey:      strings.TrimSpace(apiKey),
				ProjectID:   projectID,
				ProjectName: projectName,
				APIBaseURL:  apiBaseURL,
			}
			if err := ctx.Validate(); err != nil {
				return &CLIError{Code: exitcode.Usage, Message: err.Error()}
			}

			cfg.UpsertContext(ctx)
			if makeActive || cfg.ActiveContext == "" {
				cfg.ActiveContext = ctx.Alias
			}
			if err := app.saveConfig(cfg); err != nil {
				return err
			}

			if app.globalFlags.JSON {
				return output.PrintJSON(os.Stdout, output.Success(
					&output.ContextSummary{Alias: ctx.Alias, ProjectID: ctx.ProjectID},
					map[string]any{
						"alias":        ctx.Alias,
						"project_id":   ctx.ProjectID,
						"project_name": ctx.ProjectName,
						"api_base_url": ctx.APIBaseURL,
						"api_key":      config.MaskAPIKey(ctx.APIKey),
					},
					output.Meta{},
				))
			}

			fmt.Fprintf(os.Stdout, "context saved: %s\n", ctx.Alias)
			return nil
		},
	}

	cmd.Flags().StringVar(&apiKey, "api-key", "", "RevenueCat v2 secret API key")
	cmd.Flags().StringVar(&projectID, "project-id", "", "Project ID for the context")
	cmd.Flags().StringVar(&projectName, "project-name", "", "Optional project name cache")
	cmd.Flags().StringVar(&apiBaseURL, "api-base-url", config.DefaultAPIBaseURL, "Override the RevenueCat API base URL")
	cmd.Flags().BoolVar(&makeActive, "active", false, "Set this context as the active context")
	_ = cmd.MarkFlagRequired("api-key")

	return cmd
}

func newContextsListCommand(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List configured contexts",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := app.loadConfig()
			if err != nil {
				return err
			}

			rows := make([]map[string]string, 0, len(cfg.Contexts))
			for _, ctx := range cfg.Contexts {
				rows = append(rows, map[string]string{
					"alias":        ctx.Alias,
					"active":       fmt.Sprintf("%t", strings.EqualFold(cfg.ActiveContext, ctx.Alias)),
					"project_id":   ctx.ProjectID,
					"project_name": ctx.ProjectName,
					"api_base_url": ctx.APIBaseURL,
					"api_key":      config.MaskAPIKey(ctx.APIKey),
				})
			}

			if shouldTable(app.globalFlags) {
				return output.PrintTable(os.Stdout, rows)
			}

			return output.PrintJSON(os.Stdout, output.Success(nil, map[string]any{
				"active_context": cfg.ActiveContext,
				"contexts":       rows,
			}, output.Meta{}))
		},
	}
}

func newContextsUseCommand(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "use <alias>",
		Short: "Set the active context",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := app.loadConfig()
			if err != nil {
				return err
			}

			if _, ok := cfg.FindContext(args[0]); !ok {
				return &CLIError{Code: exitcode.Context, Message: fmt.Sprintf("context %q not found", args[0])}
			}

			cfg.ActiveContext = args[0]
			if err := app.saveConfig(cfg); err != nil {
				return err
			}

			fmt.Fprintf(os.Stdout, "active context: %s\n", args[0])
			return nil
		},
	}
}

func newContextsShowCommand(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "show [alias]",
		Short: "Show one context or the active context",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := app.loadConfig()
			if err != nil {
				return err
			}

			alias := cfg.ActiveContext
			if len(args) == 1 {
				alias = args[0]
			}
			if alias == "" {
				return &CLIError{Code: exitcode.Context, Message: "no active context set"}
			}

			ctx, ok := cfg.FindContext(alias)
			if !ok {
				return &CLIError{Code: exitcode.Context, Message: fmt.Sprintf("context %q not found", alias)}
			}

			payload := map[string]any{
				"alias":           ctx.Alias,
				"project_id":      ctx.ProjectID,
				"project_name":    ctx.ProjectName,
				"api_base_url":    ctx.APIBaseURL,
				"api_key":         config.MaskAPIKey(ctx.APIKey),
				"cached_metadata": ctx.CachedMetadata,
			}
			return output.PrintJSON(os.Stdout, output.Success(app.outputContext(*ctx), payload, output.Meta{}))
		},
	}
}

func newContextsRemoveCommand(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "remove <alias>",
		Short: "Remove a configured context",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := app.loadConfig()
			if err != nil {
				return err
			}

			if !cfg.RemoveContext(args[0]) {
				return &CLIError{Code: exitcode.Context, Message: fmt.Sprintf("context %q not found", args[0])}
			}
			if err := app.saveConfig(cfg); err != nil {
				return err
			}

			fmt.Fprintf(os.Stdout, "context removed: %s\n", args[0])
			return nil
		},
	}
}

func newContextsVerifyCommand(app *App) *cobra.Command {
	var verifyAll bool

	cmd := &cobra.Command{
		Use:   "verify [alias]",
		Short: "Verify one or more contexts and cache project metadata when possible",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := app.loadConfig()
			if err != nil {
				return err
			}

			targets := make([]*config.Context, 0)
			switch {
			case verifyAll:
				for i := range cfg.Contexts {
					targets = append(targets, &cfg.Contexts[i])
				}
			case len(args) == 1:
				ctx, ok := cfg.FindContext(args[0])
				if !ok {
					return &CLIError{Code: exitcode.Context, Message: fmt.Sprintf("context %q not found", args[0])}
				}
				targets = append(targets, ctx)
			default:
				ctx, err := app.resolveSingleContext(cfg)
				if err != nil {
					return err
				}
				targets = append(targets, ctx)
			}

			results := make([]fanoutResult, 0, len(targets))
			for _, ctx := range targets {
				result := fanoutResult{
					ContextAlias: ctx.Alias,
					ProjectID:    ctx.ProjectID,
				}
				client := app.clientFor(*ctx)
				projectInfo, err := verifyContext(context.Background(), client, *ctx)
				if err != nil {
					if apiErr, ok := err.(*rcapi.APIError); ok {
						result.Error = map[string]any{
							"type":        apiErr.Type,
							"message":     apiErr.Message,
							"status_code": apiErr.StatusCode,
						}
					} else {
						result.Error = map[string]any{"message": err.Error()}
					}
					results = append(results, result)
					continue
				}

				ctx.CachedMetadata = projectInfo
				if id, ok := projectInfo["id"].(string); ok && id != "" {
					ctx.ProjectID = id
				}
				ctx.ProjectName = projectNameFromPayload(projectInfo)

				result.OK = true
				result.ProjectID = ctx.ProjectID
				result.Data = map[string]any{
					"project_id":   ctx.ProjectID,
					"project_name": ctx.ProjectName,
				}
				results = append(results, result)
			}

			if err := app.saveConfig(cfg); err != nil {
				return err
			}

			return output.PrintJSON(os.Stdout, output.Success(nil, map[string]any{"results": results}, output.Meta{}))
		},
	}

	cmd.Flags().BoolVar(&verifyAll, "all", false, "Verify every configured context")
	return cmd
}

func verifyContext(ctx context.Context, client *rcapi.Client, stored config.Context) (map[string]any, error) {
	if stored.ProjectID != "" {
		result, err := client.Do(ctx, rcapi.Request{
			Method:    http.MethodGet,
			Path:      fmt.Sprintf("projects/%s/apps", stored.ProjectID),
			Query:     url.Values{"limit": []string{"1"}},
			RetryMode: rcapi.RetryDefault,
		})
		if err != nil {
			return nil, err
		}

		return map[string]any{
			"id":                stored.ProjectID,
			"name":              stored.ProjectName,
			"verification_hint": "apps_list",
			"request_id":        result.RequestID,
		}, nil
	}

	project, err := discoverProject(ctx, client)
	if err != nil {
		return nil, err
	}
	return project, nil
}
