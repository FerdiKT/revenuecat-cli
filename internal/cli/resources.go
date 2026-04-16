package cli

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/FerdiKT/revenuecat-cli/internal/config"
	"github.com/FerdiKT/revenuecat-cli/internal/exitcode"
	"github.com/FerdiKT/revenuecat-cli/internal/output"
	"github.com/FerdiKT/revenuecat-cli/internal/rcapi"
	"github.com/spf13/cobra"
)

type resourceDefinition struct {
	Plural            string
	Singular          string
	ListPath          func(projectID string, scope pathScope) string
	GetPath           func(projectID, id string, scope pathScope) string
	CreatePath        func(projectID string, scope pathScope) string
	UpdatePath        func(projectID, id string, scope pathScope) string
	ArchivePath       func(projectID, id string, scope pathScope) string
	UnarchivePath     func(projectID, id string, scope pathScope) string
	AttachProducts    func(projectID, id string, scope pathScope) string
	DetachProducts    func(projectID, id string, scope pathScope) string
	SupportsArchive   bool
	SupportsAttach    bool
	NeedsOffering     bool
	NeedsCustomerList bool
	ReadOnly          bool
}

type pathScope struct {
	OfferingID string
	CustomerID string
}

func addResourceCommands(root *cobra.Command, app *App) {
	root.AddCommand(
		newStandardResourceCommand(app, resourceDefinition{
			Plural:   "apps",
			Singular: "app",
			ListPath: func(projectID string, _ pathScope) string { return fmt.Sprintf("projects/%s/apps", projectID) },
			GetPath: func(projectID, id string, _ pathScope) string {
				return fmt.Sprintf("projects/%s/apps/%s", projectID, id)
			},
			CreatePath: func(projectID string, _ pathScope) string {
				return fmt.Sprintf("projects/%s/apps", projectID)
			},
			UpdatePath: func(projectID, id string, _ pathScope) string {
				return fmt.Sprintf("projects/%s/apps/%s", projectID, id)
			},
		}),
		newStandardResourceCommand(app, resourceDefinition{
			Plural:   "entitlements",
			Singular: "entitlement",
			ListPath: func(projectID string, _ pathScope) string { return fmt.Sprintf("projects/%s/entitlements", projectID) },
			GetPath: func(projectID, id string, _ pathScope) string {
				return fmt.Sprintf("projects/%s/entitlements/%s", projectID, id)
			},
			CreatePath: func(projectID string, _ pathScope) string {
				return fmt.Sprintf("projects/%s/entitlements", projectID)
			},
			UpdatePath: func(projectID, id string, _ pathScope) string {
				return fmt.Sprintf("projects/%s/entitlements/%s", projectID, id)
			},
			ArchivePath: func(projectID, id string, _ pathScope) string {
				return fmt.Sprintf("projects/%s/entitlements/%s/actions/archive", projectID, id)
			},
			UnarchivePath: func(projectID, id string, _ pathScope) string {
				return fmt.Sprintf("projects/%s/entitlements/%s/actions/unarchive", projectID, id)
			},
			AttachProducts: func(projectID, id string, _ pathScope) string {
				return fmt.Sprintf("projects/%s/entitlements/%s/actions/attach_products", projectID, id)
			},
			DetachProducts: func(projectID, id string, _ pathScope) string {
				return fmt.Sprintf("projects/%s/entitlements/%s/actions/detach_products", projectID, id)
			},
			SupportsArchive: true,
			SupportsAttach:  true,
		}),
		newStandardResourceCommand(app, resourceDefinition{
			Plural:   "products",
			Singular: "product",
			ListPath: func(projectID string, _ pathScope) string { return fmt.Sprintf("projects/%s/products", projectID) },
			GetPath: func(projectID, id string, _ pathScope) string {
				return fmt.Sprintf("projects/%s/products/%s", projectID, id)
			},
			CreatePath: func(projectID string, _ pathScope) string {
				return fmt.Sprintf("projects/%s/products", projectID)
			},
			UpdatePath: func(projectID, id string, _ pathScope) string {
				return fmt.Sprintf("projects/%s/products/%s", projectID, id)
			},
			ArchivePath: func(projectID, id string, _ pathScope) string {
				return fmt.Sprintf("projects/%s/products/%s/actions/archive", projectID, id)
			},
			UnarchivePath: func(projectID, id string, _ pathScope) string {
				return fmt.Sprintf("projects/%s/products/%s/actions/unarchive", projectID, id)
			},
			SupportsArchive: true,
		}),
		newStandardResourceCommand(app, resourceDefinition{
			Plural:   "offerings",
			Singular: "offering",
			ListPath: func(projectID string, _ pathScope) string { return fmt.Sprintf("projects/%s/offerings", projectID) },
			GetPath: func(projectID, id string, _ pathScope) string {
				return fmt.Sprintf("projects/%s/offerings/%s", projectID, id)
			},
			CreatePath: func(projectID string, _ pathScope) string {
				return fmt.Sprintf("projects/%s/offerings", projectID)
			},
			UpdatePath: func(projectID, id string, _ pathScope) string {
				return fmt.Sprintf("projects/%s/offerings/%s", projectID, id)
			},
			ArchivePath: func(projectID, id string, _ pathScope) string {
				return fmt.Sprintf("projects/%s/offerings/%s/actions/archive", projectID, id)
			},
			UnarchivePath: func(projectID, id string, _ pathScope) string {
				return fmt.Sprintf("projects/%s/offerings/%s/actions/unarchive", projectID, id)
			},
			SupportsArchive: true,
		}),
		newStandardResourceCommand(app, resourceDefinition{
			Plural:        "packages",
			Singular:      "package",
			NeedsOffering: true,
			ListPath: func(projectID string, scope pathScope) string {
				return fmt.Sprintf("projects/%s/offerings/%s/packages", projectID, scope.OfferingID)
			},
			GetPath: func(projectID, id string, scope pathScope) string {
				return fmt.Sprintf("projects/%s/offerings/%s/packages/%s", projectID, scope.OfferingID, id)
			},
			CreatePath: func(projectID string, scope pathScope) string {
				return fmt.Sprintf("projects/%s/offerings/%s/packages", projectID, scope.OfferingID)
			},
			UpdatePath: func(projectID, id string, scope pathScope) string {
				return fmt.Sprintf("projects/%s/offerings/%s/packages/%s", projectID, scope.OfferingID, id)
			},
			AttachProducts: func(projectID, id string, scope pathScope) string {
				return fmt.Sprintf("projects/%s/offerings/%s/packages/%s/actions/attach_products", projectID, scope.OfferingID, id)
			},
			DetachProducts: func(projectID, id string, scope pathScope) string {
				return fmt.Sprintf("projects/%s/offerings/%s/packages/%s/actions/detach_products", projectID, scope.OfferingID, id)
			},
			SupportsAttach: true,
		}),
		newStandardResourceCommand(app, resourceDefinition{
			Plural:   "customers",
			Singular: "customer",
			ReadOnly: true,
			ListPath: func(projectID string, _ pathScope) string { return fmt.Sprintf("projects/%s/customers", projectID) },
			GetPath: func(projectID, id string, _ pathScope) string {
				return fmt.Sprintf("projects/%s/customers/%s", projectID, id)
			},
		}),
		newStandardResourceCommand(app, resourceDefinition{
			Plural:            "subscriptions",
			Singular:          "subscription",
			ReadOnly:          true,
			NeedsCustomerList: true,
			ListPath: func(projectID string, scope pathScope) string {
				return fmt.Sprintf("projects/%s/customers/%s/subscriptions", projectID, scope.CustomerID)
			},
			GetPath: func(projectID, id string, _ pathScope) string {
				return fmt.Sprintf("projects/%s/subscriptions/%s", projectID, id)
			},
		}),
		newStandardResourceCommand(app, resourceDefinition{
			Plural:            "purchases",
			Singular:          "purchase",
			ReadOnly:          true,
			NeedsCustomerList: true,
			ListPath: func(projectID string, scope pathScope) string {
				return fmt.Sprintf("projects/%s/customers/%s/purchases", projectID, scope.CustomerID)
			},
			GetPath: func(projectID, id string, _ pathScope) string {
				return fmt.Sprintf("projects/%s/purchases/%s", projectID, id)
			},
		}),
		newMetricsCommand(app),
		newPullCommand(app),
	)
}

func newStandardResourceCommand(app *App, def resourceDefinition) *cobra.Command {
	cmd := &cobra.Command{
		Use:   def.Plural,
		Short: fmt.Sprintf("Manage %s", def.Plural),
	}

	cmd.AddCommand(
		newListCommand(app, def),
		newGetCommand(app, def),
	)
	if !def.ReadOnly {
		cmd.AddCommand(newCreateCommand(app, def), newUpdateCommand(app, def))
		if def.SupportsArchive {
			cmd.AddCommand(newArchiveCommand(app, def, true), newArchiveCommand(app, def, false))
		}
		if def.SupportsAttach {
			cmd.AddCommand(newAttachDetachCommand(app, def, true), newAttachDetachCommand(app, def, false))
		}
	}

	return cmd
}

func newListCommand(app *App, def resourceDefinition) *cobra.Command {
	var flags requestFlags
	scope := pathScope{}
	cmd := &cobra.Command{
		Use:   "list",
		Short: fmt.Sprintf("List %s", def.Plural),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := app.loadConfig()
			if err != nil {
				return err
			}
			contexts, err := app.resolveReadContexts(cfg)
			if err != nil {
				return err
			}

			query, err := parseQuery(flags)
			if err != nil {
				return err
			}

			return app.runReadAcrossContexts(contexts, func(ctx config.Context) (*rcapi.Result, error) {
				projectID, err := app.ensureProjectID(ctx)
				if err != nil {
					return nil, err
				}
				pathScope, err := validateScope(def, scope, true)
				if err != nil {
					return nil, err
				}
				return app.clientFor(ctx).Do(context.Background(), rcapi.Request{
					Method:    http.MethodGet,
					Path:      def.ListPath(projectID, pathScope),
					Query:     query,
					RetryMode: app.requestRetryMode(http.MethodGet),
				})
			})
		},
	}
	addReadFlags(cmd, &flags)
	addScopeFlags(cmd, def, &scope)
	return cmd
}

func newGetCommand(app *App, def resourceDefinition) *cobra.Command {
	var flags requestFlags
	scope := pathScope{}
	cmd := &cobra.Command{
		Use:   fmt.Sprintf("get <%s_id>", def.Singular),
		Short: fmt.Sprintf("Get a %s", def.Singular),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := app.loadConfig()
			if err != nil {
				return err
			}
			contexts, err := app.resolveReadContexts(cfg)
			if err != nil {
				return err
			}
			query, err := parseQuery(flags)
			if err != nil {
				return err
			}

			return app.runReadAcrossContexts(contexts, func(ctx config.Context) (*rcapi.Result, error) {
				projectID, err := app.ensureProjectID(ctx)
				if err != nil {
					return nil, err
				}
				pathScope, err := validateScope(def, scope, false)
				if err != nil {
					return nil, err
				}
				return app.clientFor(ctx).Do(context.Background(), rcapi.Request{
					Method:    http.MethodGet,
					Path:      def.GetPath(projectID, args[0], pathScope),
					Query:     query,
					RetryMode: app.requestRetryMode(http.MethodGet),
				})
			})
		},
	}
	addReadFlags(cmd, &flags)
	addScopeFlags(cmd, def, &scope)
	return cmd
}

func newCreateCommand(app *App, def resourceDefinition) *cobra.Command {
	var flags requestFlags
	scope := pathScope{}
	cmd := &cobra.Command{
		Use:   "create",
		Short: fmt.Sprintf("Create a %s", def.Singular),
		RunE: func(cmd *cobra.Command, args []string) error {
			body, err := parseBody(flags.data, flags.dataFile)
			if err != nil {
				return err
			}
			return runMutation(app, def.Singular+" created", http.MethodPost, body, func(ctx config.Context) (string, error) {
				projectID, err := app.ensureProjectID(ctx)
				if err != nil {
					return "", err
				}
				pathScope, err := validateScope(def, scope, false)
				if err != nil {
					return "", err
				}
				return def.CreatePath(projectID, pathScope), nil
			})
		},
	}
	addRequestBodyFlags(cmd, &flags)
	addScopeFlags(cmd, def, &scope)
	return cmd
}

func newUpdateCommand(app *App, def resourceDefinition) *cobra.Command {
	var flags requestFlags
	scope := pathScope{}
	cmd := &cobra.Command{
		Use:   fmt.Sprintf("update <%s_id>", def.Singular),
		Short: fmt.Sprintf("Update a %s", def.Singular),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			body, err := parseBody(flags.data, flags.dataFile)
			if err != nil {
				return err
			}
			return runMutation(app, def.Singular+" updated", http.MethodPost, body, func(ctx config.Context) (string, error) {
				projectID, err := app.ensureProjectID(ctx)
				if err != nil {
					return "", err
				}
				pathScope, err := validateScope(def, scope, false)
				if err != nil {
					return "", err
				}
				return def.UpdatePath(projectID, args[0], pathScope), nil
			})
		},
	}
	addRequestBodyFlags(cmd, &flags)
	addScopeFlags(cmd, def, &scope)
	return cmd
}

func newArchiveCommand(app *App, def resourceDefinition, archive bool) *cobra.Command {
	scope := pathScope{}
	name := "archive"
	verb := "archived"
	builder := def.ArchivePath
	if !archive {
		name = "unarchive"
		verb = "unarchived"
		builder = def.UnarchivePath
	}

	cmd := &cobra.Command{
		Use:   fmt.Sprintf("%s <%s_id>", name, def.Singular),
		Short: fmt.Sprintf("%s a %s", strings.Title(name), def.Singular),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMutation(app, fmt.Sprintf("%s %s", def.Singular, verb), http.MethodPost, map[string]any{}, func(ctx config.Context) (string, error) {
				projectID, err := app.ensureProjectID(ctx)
				if err != nil {
					return "", err
				}
				pathScope, err := validateScope(def, scope, false)
				if err != nil {
					return "", err
				}
				return builder(projectID, args[0], pathScope), nil
			})
		},
	}
	addScopeFlags(cmd, def, &scope)
	return cmd
}

func newAttachDetachCommand(app *App, def resourceDefinition, attach bool) *cobra.Command {
	scope := pathScope{}
	var flags requestFlags
	name := "attach-products"
	verb := "products attached"
	builder := def.AttachProducts
	if !attach {
		name = "detach-products"
		verb = "products detached"
		builder = def.DetachProducts
	}
	cmd := &cobra.Command{
		Use:   fmt.Sprintf("%s <%s_id>", name, def.Singular),
		Short: fmt.Sprintf("%s for a %s", name, def.Singular),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			body, err := parseBody(flags.data, flags.dataFile)
			if err != nil {
				return err
			}
			return runMutation(app, fmt.Sprintf("%s: %s", def.Singular, verb), http.MethodPost, body, func(ctx config.Context) (string, error) {
				projectID, err := app.ensureProjectID(ctx)
				if err != nil {
					return "", err
				}
				pathScope, err := validateScope(def, scope, false)
				if err != nil {
					return "", err
				}
				return builder(projectID, args[0], pathScope), nil
			})
		},
	}
	addRequestBodyFlags(cmd, &flags)
	addScopeFlags(cmd, def, &scope)
	return cmd
}

func runMutation(app *App, action string, method string, body any, pathBuilder func(ctx config.Context) (string, error)) error {
	cfg, err := app.loadConfig()
	if err != nil {
		return err
	}
	ctx, err := app.resolveSingleContext(cfg)
	if err != nil {
		return err
	}

	requestPath, err := pathBuilder(*ctx)
	if err != nil {
		return err
	}

	result, err := app.clientFor(*ctx).Do(context.Background(), rcapi.Request{
		Method:    method,
		Path:      requestPath,
		Body:      body,
		RetryMode: app.requestRetryMode(method),
	})
	if err != nil {
		var apiErr *rcapi.APIError
		if ok := errorAs(err, &apiErr); ok {
			return app.mapAPIError(ctx, apiErr, "")
		}
		return &CLIError{Code: exitcode.Internal, Message: err.Error()}
	}

	return app.renderMutation(*ctx, result, action)
}

func errorAs(err error, target **rcapi.APIError) bool {
	apiErr, ok := err.(*rcapi.APIError)
	if !ok {
		return false
	}
	*target = apiErr
	return true
}

func addScopeFlags(cmd *cobra.Command, def resourceDefinition, scope *pathScope) {
	if def.NeedsOffering {
		cmd.Flags().StringVar(&scope.OfferingID, "offering-id", "", "Offering ID required for package operations")
		_ = cmd.MarkFlagRequired("offering-id")
	}
	if def.NeedsCustomerList {
		cmd.Flags().StringVar(&scope.CustomerID, "customer-id", "", "Customer ID required when listing this resource")
	}
}

func validateScope(def resourceDefinition, scope pathScope, isList bool) (pathScope, error) {
	if def.NeedsOffering && strings.TrimSpace(scope.OfferingID) == "" {
		return scope, &CLIError{Code: exitcode.Usage, Message: "--offering-id is required"}
	}
	if isList && def.NeedsCustomerList && strings.TrimSpace(scope.CustomerID) == "" {
		return scope, &CLIError{Code: exitcode.Usage, Message: "--customer-id is required"}
	}
	return scope, nil
}

func newMetricsCommand(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "metrics",
		Short: "Read RevenueCat metrics and charts",
	}
	cmd.AddCommand(newMetricsOverviewCommand(app), newMetricsChartCommand(app), newMetricsOptionsCommand(app))
	return cmd
}

func newMetricsOverviewCommand(app *App) *cobra.Command {
	var flags requestFlags
	cmd := &cobra.Command{
		Use:   "overview",
		Short: "Get overview metrics for the current project",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := app.loadConfig()
			if err != nil {
				return err
			}
			contexts, err := app.resolveReadContexts(cfg)
			if err != nil {
				return err
			}
			query, err := parseQuery(flags)
			if err != nil {
				return err
			}
			return app.runReadAcrossContexts(contexts, func(ctx config.Context) (*rcapi.Result, error) {
				projectID, err := app.ensureProjectID(ctx)
				if err != nil {
					return nil, err
				}
				return app.clientFor(ctx).Do(context.Background(), rcapi.Request{
					Method:    http.MethodGet,
					Path:      fmt.Sprintf("projects/%s/metrics/overview", projectID),
					Query:     query,
					RetryMode: app.requestRetryMode(http.MethodGet),
				})
			})
		},
	}
	addReadFlags(cmd, &flags)
	return cmd
}

func newMetricsChartCommand(app *App) *cobra.Command {
	var flags requestFlags
	cmd := &cobra.Command{
		Use:   "chart <chart_name>",
		Short: "Get chart data for a project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := app.loadConfig()
			if err != nil {
				return err
			}
			contexts, err := app.resolveReadContexts(cfg)
			if err != nil {
				return err
			}
			query, err := parseQuery(flags)
			if err != nil {
				return err
			}
			return app.runReadAcrossContexts(contexts, func(ctx config.Context) (*rcapi.Result, error) {
				projectID, err := app.ensureProjectID(ctx)
				if err != nil {
					return nil, err
				}
				return app.clientFor(ctx).Do(context.Background(), rcapi.Request{
					Method:    http.MethodGet,
					Path:      fmt.Sprintf("projects/%s/charts/%s", projectID, args[0]),
					Query:     query,
					RetryMode: app.requestRetryMode(http.MethodGet),
				})
			})
		},
	}
	addReadFlags(cmd, &flags)
	return cmd
}

func newMetricsOptionsCommand(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "options <chart_name>",
		Short: "Get available options for a chart",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := app.loadConfig()
			if err != nil {
				return err
			}
			contexts, err := app.resolveReadContexts(cfg)
			if err != nil {
				return err
			}
			return app.runReadAcrossContexts(contexts, func(ctx config.Context) (*rcapi.Result, error) {
				projectID, err := app.ensureProjectID(ctx)
				if err != nil {
					return nil, err
				}
				return app.clientFor(ctx).Do(context.Background(), rcapi.Request{
					Method:    http.MethodGet,
					Path:      fmt.Sprintf("projects/%s/charts/%s/options", projectID, args[0]),
					RetryMode: app.requestRetryMode(http.MethodGet),
				})
			})
		},
	}
	return cmd
}

func newPullCommand(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pull",
		Short: "Export normalized project snapshots for agents",
	}
	cmd.AddCommand(newPullProjectCommand(app), newPullAllCommand(app))
	return cmd
}

func newPullProjectCommand(app *App) *cobra.Command {
	var includeCharts []string
	var includeCustomers bool
	var includeSubscriptions bool
	var includePurchases bool
	cmd := &cobra.Command{
		Use:   "project",
		Short: "Pull a normalized snapshot for the active context",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := app.loadConfig()
			if err != nil {
				return err
			}
			ctx, err := app.resolveSingleContext(cfg)
			if err != nil {
				return err
			}

			bundle, requestIDs, err := app.pullBundle(*ctx, includeCharts, includeCustomers, includeSubscriptions, includePurchases)
			if err != nil {
				var apiErr *rcapi.APIError
				if ok := errorAs(err, &apiErr); ok {
					return app.mapAPIError(ctx, apiErr, "")
				}
				return err
			}

			payload := map[string]any{
				"schema_version": "v1alpha1",
				"generated_at":   time.Now().UTC().Format(time.RFC3339),
				"context": map[string]any{
					"alias":      ctx.Alias,
					"project_id": ctx.ProjectID,
				},
				"project_bundle": bundle,
			}
			return output.PrintJSON(os.Stdout, output.Success(app.outputContext(*ctx), payload, output.Meta{
				RequestID: strings.Join(requestIDs, ","),
			}))
		},
	}
	cmd.Flags().StringSliceVar(&includeCharts, "chart", nil, "Additional chart names to include")
	cmd.Flags().BoolVar(&includeCustomers, "include-customers", false, "Include all customers")
	cmd.Flags().BoolVar(&includeSubscriptions, "include-subscriptions", false, "Include subscriptions for every fetched customer")
	cmd.Flags().BoolVar(&includePurchases, "include-purchases", false, "Include purchases for every fetched customer")
	return cmd
}

func newPullAllCommand(app *App) *cobra.Command {
	var includeCharts []string
	var includeCustomers bool
	var includeSubscriptions bool
	var includePurchases bool
	cmd := &cobra.Command{
		Use:   "all",
		Short: "Pull normalized snapshots across all configured contexts",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := app.loadConfig()
			if err != nil {
				return err
			}
			contexts := append([]config.Context(nil), cfg.Contexts...)
			if len(contexts) == 0 {
				return &CLIError{Code: exitcode.Context, Message: "no contexts configured"}
			}

			projects := make([]map[string]any, 0, len(contexts))
			failures := make([]fanoutResult, 0)
			for _, ctx := range contexts {
				bundle, requestIDs, err := app.pullBundle(ctx, includeCharts, includeCustomers, includeSubscriptions, includePurchases)
				if err != nil {
					entry := fanoutResult{ContextAlias: ctx.Alias, ProjectID: ctx.ProjectID}
					var apiErr *rcapi.APIError
					if ok := errorAs(err, &apiErr); ok {
						entry.Error = map[string]any{
							"type":        apiErr.Type,
							"message":     apiErr.Message,
							"status_code": apiErr.StatusCode,
						}
					} else {
						entry.Error = map[string]any{"message": err.Error()}
					}
					failures = append(failures, entry)
					continue
				}

				projects = append(projects, map[string]any{
					"context_alias": ctx.Alias,
					"project_id":    ctx.ProjectID,
					"request_ids":   requestIDs,
					"bundle":        bundle,
				})
			}

			payload := map[string]any{
				"schema_version": "v1alpha1",
				"generated_at":   time.Now().UTC().Format(time.RFC3339),
				"projects":       projects,
			}
			if len(failures) > 0 {
				envelope := output.Failure(nil, output.Meta{}, &output.ErrorPayload{
					Type:    "fanout_partial_failure",
					Message: "one or more contexts failed during pull all",
				})
				envelope.Data = map[string]any{
					"projects": projects,
					"errors":   failures,
				}
				if err := output.PrintJSON(os.Stdout, envelope); err != nil {
					return &CLIError{Code: exitcode.Internal, Message: err.Error()}
				}
				return &CLIError{Code: exitcode.Retryable, Message: ""}
			}

			return output.PrintJSON(os.Stdout, output.Success(nil, payload, output.Meta{}))
		},
	}
	cmd.Flags().StringSliceVar(&includeCharts, "chart", nil, "Additional chart names to include")
	cmd.Flags().BoolVar(&includeCustomers, "include-customers", false, "Include all customers")
	cmd.Flags().BoolVar(&includeSubscriptions, "include-subscriptions", false, "Include subscriptions for every fetched customer")
	cmd.Flags().BoolVar(&includePurchases, "include-purchases", false, "Include purchases for every fetched customer")
	return cmd
}

func (a *App) pullBundle(ctx config.Context, includeCharts []string, includeCustomers bool, includeSubscriptions bool, includePurchases bool) (map[string]any, []string, error) {
	projectID, err := a.ensureProjectID(ctx)
	if err != nil {
		return nil, nil, err
	}
	client := a.clientFor(ctx)
	baseQuery := url.Values{"limit": []string{"100"}}

	apps, reqA, err := fetchAllListItems(context.Background(), client, fmt.Sprintf("projects/%s/apps", projectID), baseQuery)
	if err != nil {
		return nil, nil, err
	}
	entitlements, reqE, err := fetchAllListItems(context.Background(), client, fmt.Sprintf("projects/%s/entitlements", projectID), baseQuery)
	if err != nil {
		return nil, nil, err
	}
	products, reqP, err := fetchAllListItems(context.Background(), client, fmt.Sprintf("projects/%s/products", projectID), baseQuery)
	if err != nil {
		return nil, nil, err
	}
	offerings, reqO, err := fetchAllListItems(context.Background(), client, fmt.Sprintf("projects/%s/offerings", projectID), baseQuery)
	if err != nil {
		return nil, nil, err
	}
	packagesByOffering := map[string]any{}
	requestIDs := append(append(append(append(reqA, reqE...), reqP...), reqO...))
	for _, offering := range offerings {
		object, ok := offering.(map[string]any)
		if !ok {
			continue
		}
		offeringID, _ := object["id"].(string)
		if offeringID == "" {
			continue
		}
		items, reqs, err := fetchAllListItems(context.Background(), client, fmt.Sprintf("projects/%s/offerings/%s/packages", projectID, offeringID), baseQuery)
		if err != nil {
			return nil, requestIDs, err
		}
		requestIDs = append(requestIDs, reqs...)
		packagesByOffering[offeringID] = items
	}

	overview, err := client.Do(context.Background(), rcapi.Request{
		Method:    http.MethodGet,
		Path:      fmt.Sprintf("projects/%s/metrics/overview", projectID),
		Query:     url.Values{},
		RetryMode: rcapi.RetryDefault,
	})
	if err != nil {
		return nil, requestIDs, err
	}
	requestIDs = append(requestIDs, overview.RequestID)

	charts := map[string]any{}
	for _, chart := range includeCharts {
		result, err := client.Do(context.Background(), rcapi.Request{
			Method:    http.MethodGet,
			Path:      fmt.Sprintf("projects/%s/charts/%s", projectID, chart),
			RetryMode: rcapi.RetryDefault,
		})
		if err != nil {
			return nil, requestIDs, err
		}
		requestIDs = append(requestIDs, result.RequestID)
		charts[chart] = result.Payload
	}

	bundle := map[string]any{
		"apps":         annotateItems(ctx, projectID, apps),
		"entitlements": annotateItems(ctx, projectID, entitlements),
		"products":     annotateItems(ctx, projectID, products),
		"offerings":    annotateItems(ctx, projectID, offerings),
		"packages":     packagesByOffering,
		"overview":     overview.Payload,
	}
	if len(charts) > 0 {
		bundle["charts"] = charts
	}

	if includeCustomers || includeSubscriptions || includePurchases {
		customers, reqC, err := fetchAllListItems(context.Background(), client, fmt.Sprintf("projects/%s/customers", projectID), baseQuery)
		if err != nil {
			return nil, requestIDs, err
		}
		requestIDs = append(requestIDs, reqC...)
		if includeCustomers {
			bundle["customers"] = annotateItems(ctx, projectID, customers)
		}
		if includeSubscriptions {
			subscriptions := map[string]any{}
			for _, customer := range customers {
				customerObject, ok := customer.(map[string]any)
				if !ok {
					continue
				}
				customerID, _ := customerObject["id"].(string)
				if customerID == "" {
					continue
				}
				items, reqs, err := fetchAllListItems(context.Background(), client, fmt.Sprintf("projects/%s/customers/%s/subscriptions", projectID, customerID), baseQuery)
				if err != nil {
					return nil, requestIDs, err
				}
				requestIDs = append(requestIDs, reqs...)
				subscriptions[customerID] = items
			}
			bundle["subscriptions"] = subscriptions
		}
		if includePurchases {
			purchases := map[string]any{}
			for _, customer := range customers {
				customerObject, ok := customer.(map[string]any)
				if !ok {
					continue
				}
				customerID, _ := customerObject["id"].(string)
				if customerID == "" {
					continue
				}
				items, reqs, err := fetchAllListItems(context.Background(), client, fmt.Sprintf("projects/%s/customers/%s/purchases", projectID, customerID), baseQuery)
				if err != nil {
					return nil, requestIDs, err
				}
				requestIDs = append(requestIDs, reqs...)
				purchases[customerID] = items
			}
			bundle["purchases"] = purchases
		}
	}

	return bundle, requestIDs, nil
}

func annotateItems(ctx config.Context, projectID string, items []any) []any {
	for i, item := range items {
		object, ok := item.(map[string]any)
		if !ok {
			continue
		}
		object["context_alias"] = ctx.Alias
		if _, ok := object["project_id"]; !ok {
			object["project_id"] = projectID
		}
		items[i] = object
	}
	return items
}
