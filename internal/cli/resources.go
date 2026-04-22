package cli

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
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
	DeletePath        func(projectID, id string, scope pathScope) string
	ArchivePath       func(projectID, id string, scope pathScope) string
	UnarchivePath     func(projectID, id string, scope pathScope) string
	AttachProducts    func(projectID, id string, scope pathScope) string
	DetachProducts    func(projectID, id string, scope pathScope) string
	SupportsDelete    bool
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
	appsCmd := newStandardResourceCommand(app, resourceDefinition{
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
		DeletePath: func(projectID, id string, _ pathScope) string {
			return fmt.Sprintf("projects/%s/apps/%s", projectID, id)
		},
		SupportsDelete: true,
	})
	appsCmd.AddCommand(
		newAppsResolveCommand(app),
		newAppsPublicKeysCommand(app),
		newAppsStoreKitConfigCommand(app),
	)

	root.AddCommand(
		newProjectsCommand(app),
		appsCmd,
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
			Plural:   "paywalls",
			Singular: "paywall",
			ListPath: func(projectID string, _ pathScope) string { return fmt.Sprintf("projects/%s/paywalls", projectID) },
			GetPath: func(projectID, id string, _ pathScope) string {
				return fmt.Sprintf("projects/%s/paywalls/%s", projectID, id)
			},
			CreatePath: func(projectID string, _ pathScope) string {
				return fmt.Sprintf("projects/%s/paywalls", projectID)
			},
			DeletePath: func(projectID, id string, _ pathScope) string {
				return fmt.Sprintf("projects/%s/paywalls/%s", projectID, id)
			},
			SupportsDelete: true,
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
		if def.CreatePath != nil {
			cmd.AddCommand(newCreateCommand(app, def))
		}
		if def.UpdatePath != nil {
			cmd.AddCommand(newUpdateCommand(app, def))
		}
		if def.SupportsArchive {
			cmd.AddCommand(newArchiveCommand(app, def, true), newArchiveCommand(app, def, false))
		}
		if def.SupportsDelete {
			cmd.AddCommand(newDeleteCommand(app, def))
		}
		if def.SupportsAttach {
			cmd.AddCommand(newAttachDetachCommand(app, def, true), newAttachDetachCommand(app, def, false))
		}
	}

	return cmd
}

func newProjectsCommand(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "projects",
		Short: "Manage account-level projects with OAuth",
	}
	cmd.AddCommand(newProjectsListCommand(app), newProjectsGetCommand(app), newProjectsCreateCommand(app))
	return cmd
}

func newProjectsListCommand(app *App) *cobra.Command {
	var flags requestFlags
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List projects available to the OAuth user",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := app.loadConfigMetadataOnly()
			if err != nil {
				return err
			}
			query, err := parseQuery(flags)
			if err != nil {
				return err
			}
			client, err := app.oauthClient(context.Background(), cfg)
			if err != nil {
				return err
			}
			result, err := client.Do(context.Background(), rcapi.Request{
				Method:    http.MethodGet,
				Path:      "projects",
				Query:     query,
				RetryMode: app.requestRetryMode(http.MethodGet),
			})
			if err != nil {
				var apiErr *rcapi.APIError
				if ok := errorAs(err, &apiErr); ok {
					return app.mapAPIError(nil, apiErr, "")
				}
				return &CLIError{Code: exitcode.Internal, Message: err.Error()}
			}
			return app.renderAccountRead(result)
		},
	}
	addReadFlags(cmd, &flags)
	return cmd
}

func newProjectsGetCommand(app *App) *cobra.Command {
	var flags requestFlags
	cmd := &cobra.Command{
		Use:   "get <project_id>",
		Short: "Get a project available to the OAuth user",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := app.loadConfigMetadataOnly()
			if err != nil {
				return err
			}
			query, err := parseQuery(flags)
			if err != nil {
				return err
			}
			client, err := app.oauthClient(context.Background(), cfg)
			if err != nil {
				return err
			}
			result, err := client.Do(context.Background(), rcapi.Request{
				Method:    http.MethodGet,
				Path:      fmt.Sprintf("projects/%s", args[0]),
				Query:     query,
				RetryMode: app.requestRetryMode(http.MethodGet),
			})
			if err != nil {
				var apiErr *rcapi.APIError
				if ok := errorAs(err, &apiErr); ok {
					return app.mapAPIError(nil, apiErr, "")
				}
				return &CLIError{Code: exitcode.Internal, Message: err.Error()}
			}
			return app.renderAccountRead(result)
		},
	}
	addReadFlags(cmd, &flags)
	return cmd
}

func newProjectsCreateCommand(app *App) *cobra.Command {
	var flags requestFlags
	var name string
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a project with OAuth",
		RunE: func(cmd *cobra.Command, args []string) error {
			if app.globalFlags.AllContexts || strings.TrimSpace(app.globalFlags.ContextAlias) != "" || strings.TrimSpace(app.globalFlags.ProjectID) != "" {
				return &CLIError{Code: exitcode.Usage, Message: "projects create is account-level OAuth only; do not pass --context, --all-contexts, or --project-id"}
			}
			cfg, err := app.loadConfigMetadataOnly()
			if err != nil {
				return err
			}
			body, err := projectCreateBody(name, flags)
			if err != nil {
				return err
			}
			client, err := app.oauthClient(context.Background(), cfg)
			if err != nil {
				return err
			}
			result, err := client.Do(context.Background(), rcapi.Request{
				Method:    http.MethodPost,
				Path:      "projects",
				Body:      body,
				RetryMode: app.requestRetryMode(http.MethodPost),
			})
			if err != nil {
				var apiErr *rcapi.APIError
				if ok := errorAs(err, &apiErr); ok {
					return app.mapAPIError(nil, apiErr, "")
				}
				return &CLIError{Code: exitcode.Internal, Message: err.Error()}
			}
			return app.renderAccountMutation(result, "project created")
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Project name")
	addRequestBodyFlags(cmd, &flags)
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

func newDeleteCommand(app *App, def resourceDefinition) *cobra.Command {
	scope := pathScope{}
	var confirm string

	cmd := &cobra.Command{
		Use:   fmt.Sprintf("delete <%s_id>", def.Singular),
		Short: fmt.Sprintf("Delete a %s", def.Singular),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			if confirm != id {
				return &CLIError{
					Code:    exitcode.Usage,
					Message: fmt.Sprintf("destructive delete requires --confirm %s", id),
				}
			}
			return runMutation(app, fmt.Sprintf("%s deleted", def.Singular), http.MethodDelete, nil, func(ctx config.Context) (string, error) {
				projectID, err := app.ensureProjectID(ctx)
				if err != nil {
					return "", err
				}
				pathScope, err := validateScope(def, scope, false)
				if err != nil {
					return "", err
				}
				return def.DeletePath(projectID, id, pathScope), nil
			})
		},
	}
	cmd.Flags().StringVar(&confirm, "confirm", "", "Required exact resource ID confirmation for destructive delete")
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
	cmd.AddCommand(newMetricsOverviewCommand(app), newMetricsChartCommand(app), newMetricsCountriesCommand(app), newMetricsOptionsCommand(app))
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
					Path:      fmt.Sprintf("projects/%s/charts/%s", projectID, normalizeChartName(args[0])),
					Query:     query,
					RetryMode: app.requestRetryMode(http.MethodGet),
				})
			})
		},
	}
	addMetricsQueryFlags(cmd, &flags)
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
					Path:      fmt.Sprintf("projects/%s/charts/%s/options", projectID, normalizeChartName(args[0])),
					RetryMode: app.requestRetryMode(http.MethodGet),
				})
			})
		},
	}
	return cmd
}

func newAppsResolveCommand(app *App) *cobra.Command {
	var bundleID string
	cmd := &cobra.Command{
		Use:   "resolve",
		Short: "Resolve an app ID by bundle identifier",
		RunE: func(cmd *cobra.Command, args []string) error {
			if strings.TrimSpace(bundleID) == "" {
				return &CLIError{Code: exitcode.Usage, Message: "--bundle-id is required"}
			}

			cfg, err := app.loadConfig()
			if err != nil {
				return err
			}
			contexts, err := app.resolveReadContexts(cfg)
			if err != nil {
				return err
			}

			matches := make([]map[string]any, 0)
			requestIDs := make([]string, 0)
			for _, ctx := range contexts {
				projectID, err := app.ensureProjectID(ctx)
				if err != nil {
					return err
				}
				items, reqIDs, err := fetchAllListItems(context.Background(), app.clientFor(ctx), fmt.Sprintf("projects/%s/apps", projectID), url.Values{"limit": []string{"100"}})
				if err != nil {
					var apiErr *rcapi.APIError
					if errorAs(err, &apiErr) {
						return app.mapAPIError(&ctx, apiErr, "")
					}
					return &CLIError{Code: exitcode.Internal, Message: err.Error()}
				}
				requestIDs = append(requestIDs, reqIDs...)
				for _, item := range annotateItems(ctx, projectID, items) {
					object, ok := item.(map[string]any)
					if !ok || !appMatchesBundleID(object, bundleID) {
						continue
					}
					matches = append(matches, object)
				}
			}

			if len(matches) == 0 {
				return &CLIError{Code: exitcode.NotFound, Message: fmt.Sprintf("no app found for bundle id %q", bundleID)}
			}

			if len(matches) == 1 && !app.globalFlags.JSON && !strings.EqualFold(app.globalFlags.Format, "json") && !shouldTable(app.globalFlags) {
				_, err := fmt.Fprintln(os.Stdout, matches[0]["id"])
				return err
			}

			if !app.globalFlags.JSON && !strings.EqualFold(app.globalFlags.Format, "json") {
				return output.PrintTable(os.Stdout, toTableRows(any(matches)))
			}

			payload := map[string]any{
				"bundle_id": bundleID,
				"matches":   matches,
			}
			var summary *output.ContextSummary
			if len(contexts) == 1 {
				summary = app.outputContext(contexts[0])
			}
			return output.PrintJSON(os.Stdout, output.Success(summary, payload, output.Meta{
				RequestID: strings.Join(requestIDs, ","),
			}))
		},
	}
	cmd.Flags().StringVar(&bundleID, "bundle-id", "", "Bundle identifier to resolve")
	return cmd
}

type countriesFlags struct {
	appID          string
	startDate      string
	endDate        string
	top            int
	includeOther   bool
	csv            bool
	store          string
	appleClaimType string
	organicOnly    bool
}

type chartOptionSupport struct {
	Segments map[string]struct{}
	Filters  map[string]struct{}
}

type segmentInfo struct {
	ID          string
	DisplayName string
}

func newAppsPublicKeysCommand(app *App) *cobra.Command {
	var flags requestFlags
	cmd := &cobra.Command{
		Use:     "public-keys <app_id>",
		Aliases: []string{"api-keys"},
		Short:   "List public API keys for an app",
		Args:    cobra.ExactArgs(1),
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
			appID := args[0]
			return app.runReadAcrossContexts(contexts, func(ctx config.Context) (*rcapi.Result, error) {
				projectID, err := app.ensureProjectID(ctx)
				if err != nil {
					return nil, err
				}
				return app.clientFor(ctx).Do(context.Background(), rcapi.Request{
					Method:    http.MethodGet,
					Path:      fmt.Sprintf("projects/%s/apps/%s/public_api_keys", projectID, appID),
					Query:     query,
					RetryMode: app.requestRetryMode(http.MethodGet),
				})
			})
		},
	}
	addReadFlags(cmd, &flags)
	return cmd
}

func newAppsStoreKitConfigCommand(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "storekit-config <app_id>",
		Aliases: []string{"store-config", "storekit"},
		Short:   "Get the StoreKit configuration for an app",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := app.loadConfig()
			if err != nil {
				return err
			}
			contexts, err := app.resolveReadContexts(cfg)
			if err != nil {
				return err
			}
			appID := args[0]
			return app.runReadAcrossContexts(contexts, func(ctx config.Context) (*rcapi.Result, error) {
				projectID, err := app.ensureProjectID(ctx)
				if err != nil {
					return nil, err
				}
				return app.clientFor(ctx).Do(context.Background(), rcapi.Request{
					Method:    http.MethodGet,
					Path:      fmt.Sprintf("projects/%s/apps/%s/store_kit_config", projectID, appID),
					RetryMode: app.requestRetryMode(http.MethodGet),
				})
			})
		},
	}
	return cmd
}

func newMetricsCountriesCommand(app *App) *cobra.Command {
	var queryFlags requestFlags
	var flags countriesFlags
	cmd := &cobra.Command{
		Use:   "countries <chart_name>",
		Short: "Render a country breakdown table for a chart",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if flags.csv && app.globalFlags.JSON {
				return &CLIError{Code: exitcode.Usage, Message: "use either --csv or --json, not both"}
			}
			if flags.organicOnly && strings.TrimSpace(flags.appleClaimType) != "" {
				return &CLIError{Code: exitcode.Usage, Message: "use either --organic-only or --apple-claim-type"}
			}

			cfg, err := app.loadConfig()
			if err != nil {
				return err
			}
			contexts, err := app.resolveReadContexts(cfg)
			if err != nil {
				return err
			}

			chartName := normalizeChartName(args[0])
			baseQuery, err := parseQuery(queryFlags)
			if err != nil {
				return err
			}

			rows := make([]map[string]string, 0)
			requestIDs := make([]string, 0)
			for _, ctx := range contexts {
				projectID, err := app.ensureProjectID(ctx)
				if err != nil {
					return err
				}
				support, optionsRequestID, err := loadChartOptionSupport(context.Background(), app.clientFor(ctx), projectID, chartName)
				if err != nil {
					var apiErr *rcapi.APIError
					if errorAs(err, &apiErr) {
						return app.mapAPIError(&ctx, apiErr, "")
					}
					return &CLIError{Code: exitcode.Internal, Message: err.Error()}
				}
				if optionsRequestID != "" {
					requestIDs = append(requestIDs, optionsRequestID)
				}
				if err := validateCountriesSupport(chartName, ctx.Alias, support, flags); err != nil {
					return err
				}
				query, err := buildCountriesQuery(baseQuery, flags)
				if err != nil {
					return err
				}
				result, err := app.clientFor(ctx).Do(context.Background(), rcapi.Request{
					Method:    http.MethodGet,
					Path:      fmt.Sprintf("projects/%s/charts/%s", projectID, chartName),
					Query:     query,
					RetryMode: app.requestRetryMode(http.MethodGet),
				})
				if err != nil {
					var apiErr *rcapi.APIError
					if errorAs(err, &apiErr) {
						return app.mapAPIError(&ctx, apiErr, "")
					}
					return &CLIError{Code: exitcode.Internal, Message: err.Error()}
				}
				requestIDs = append(requestIDs, result.RequestID)
				if err := validateUnsupportedChartParams(chartName, ctx.Alias, result.Payload, flags); err != nil {
					return err
				}

				contextRows, err := extractCountryRows(result.Payload, chartName)
				if err != nil {
					return &CLIError{Code: exitcode.Internal, Message: err.Error()}
				}
				for _, row := range contextRows {
					if !flags.includeOther && strings.EqualFold(strings.TrimSpace(row["country"]), "other") {
						continue
					}
					if len(contexts) > 1 {
						row["context_alias"] = ctx.Alias
						row["project_id"] = projectID
					}
					rows = append(rows, row)
				}
			}

			if len(rows) == 0 {
				return &CLIError{Code: exitcode.NotFound, Message: fmt.Sprintf("no country rows returned for chart %q", chartName)}
			}

			if flags.csv {
				return printCSV(orderedCountryHeaders(rows), rows)
			}

			if app.globalFlags.JSON || strings.EqualFold(app.globalFlags.Format, "json") {
				payload := map[string]any{
					"chart_name": chartName,
					"rows":       rows,
				}
				var summary *output.ContextSummary
				if len(contexts) == 1 {
					summary = app.outputContext(contexts[0])
				}
				return output.PrintJSON(os.Stdout, output.Success(summary, payload, output.Meta{
					RequestID: strings.Join(requestIDs, ","),
				}))
			}

			return output.PrintTable(os.Stdout, rows)
		},
	}
	addMetricsQueryFlags(cmd, &queryFlags)
	cmd.Flags().StringVar(&flags.appID, "app", "", "Filter to a specific RevenueCat app ID")
	cmd.Flags().StringVar(&flags.startDate, "start", "", "Start date in YYYY-MM-DD format")
	cmd.Flags().StringVar(&flags.endDate, "end", "", "End date in YYYY-MM-DD format")
	cmd.Flags().IntVar(&flags.top, "top", 0, "Limit the country breakdown to the top N segments")
	cmd.Flags().BoolVar(&flags.includeOther, "include-other", false, "Keep the aggregated Other row when segment limiting is used")
	cmd.Flags().BoolVar(&flags.csv, "csv", false, "Print CSV instead of a table")
	cmd.Flags().StringVar(&flags.store, "store", "", "Filter to a specific store, e.g. app_store")
	cmd.Flags().StringVar(&flags.appleClaimType, "apple-claim-type", "", "Filter by Apple claim type, e.g. Organic")
	cmd.Flags().BoolVar(&flags.organicOnly, "organic-only", false, "Shortcut for --apple-claim-type Organic")
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
	paywalls, reqW, err := fetchAllListItems(context.Background(), client, fmt.Sprintf("projects/%s/paywalls", projectID), baseQuery)
	if err != nil {
		return nil, nil, err
	}
	packagesByOffering := map[string]any{}
	requestIDs := append(append(append(append(append(reqA, reqE...), reqP...), reqO...), reqW...))
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
		"paywalls":     annotateItems(ctx, projectID, paywalls),
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

func buildCountriesQuery(baseQuery url.Values, flags countriesFlags) (url.Values, error) {
	query := cloneValues(baseQuery)
	query.Set("aggregate", "total")
	query.Set("segment", "country")
	if strings.TrimSpace(flags.startDate) != "" {
		query.Set("start_date", strings.TrimSpace(flags.startDate))
	}
	if strings.TrimSpace(flags.endDate) != "" {
		query.Set("end_date", strings.TrimSpace(flags.endDate))
	}
	if flags.top > 0 {
		query.Set("limit_num_segments", strconv.Itoa(flags.top))
	}

	filters, err := parseJSONArray(query.Get("filters"))
	if err != nil {
		return nil, &CLIError{Code: exitcode.Usage, Message: fmt.Sprintf("invalid filters payload: %v", err)}
	}
	if strings.TrimSpace(flags.appID) != "" {
		filters = upsertFilter(filters, "app_id", []string{strings.TrimSpace(flags.appID)})
	}
	if strings.TrimSpace(flags.store) != "" {
		filters = upsertFilter(filters, "store", []string{strings.TrimSpace(flags.store)})
	}
	if flags.organicOnly {
		filters = upsertFilter(filters, "apple_claim_type", []string{"Organic"})
	}
	if strings.TrimSpace(flags.appleClaimType) != "" {
		filters = upsertFilter(filters, "apple_claim_type", []string{strings.TrimSpace(flags.appleClaimType)})
	}
	if len(filters) > 0 {
		encoded, err := encodedJSON(filters)
		if err != nil {
			return nil, &CLIError{Code: exitcode.Internal, Message: fmt.Sprintf("encode filters: %v", err)}
		}
		query.Set("filters", encoded)
	}

	selectors, err := parseJSONObject(query.Get("selectors"))
	if err != nil {
		return nil, &CLIError{Code: exitcode.Usage, Message: fmt.Sprintf("invalid selectors payload: %v", err)}
	}
	if len(selectors) > 0 {
		encoded, err := encodedJSON(selectors)
		if err != nil {
			return nil, &CLIError{Code: exitcode.Internal, Message: fmt.Sprintf("encode selectors: %v", err)}
		}
		query.Set("selectors", encoded)
	}

	return query, nil
}

func extractCountryRows(payload any, chartName string) ([]map[string]string, error) {
	root, ok := payload.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("unexpected %s chart payload", chartName)
	}

	segments, segmentNames := extractSegmentInfo(root["segments"])

	measures := measureIDs(root["measures"])
	if rows := rowsFromSummaryTotal(root["summary"], segments, segmentNames, measures); len(rows) > 0 {
		return rows, nil
	}
	if rows := rowsFromSummary(root["summary"], segmentNames, measures); len(rows) > 0 {
		return rows, nil
	}

	return nil, fmt.Errorf("could not extract segmented country rows from %s chart summary", chartName)
}

func rowsFromSummaryTotal(summary any, segments []segmentInfo, segmentNames map[string]string, measures []string) []map[string]string {
	object, ok := summary.(map[string]any)
	if !ok {
		return nil
	}
	total, ok := object["total"]
	if !ok {
		return nil
	}

	switch current := total.(type) {
	case []any:
		return rowsFromSummaryTotalArray(current, segments, measures)
	case map[string]any:
		if rows := rowsFromMeasureArrays(current, segments, measures); len(rows) > 0 {
			return rows
		}
		return rowsFromSummary(current, segmentNames, measures)
	default:
		return nil
	}
}

func rowsFromSummary(summary any, segmentNames map[string]string, measures []string) []map[string]string {
	switch value := summary.(type) {
	case map[string]any:
		if nested, ok := value["segments"]; ok {
			if rows := rowsFromSummary(nested, segmentNames, measures); len(rows) > 0 {
				return rows
			}
		}

		rows := make([]map[string]string, 0)
		for segmentID, displayName := range segmentNames {
			entry, ok := value[segmentID]
			if !ok {
				continue
			}
			row := map[string]string{"country": displayName}
			switch current := entry.(type) {
			case map[string]any:
				scalars := scalarMeasureFields(current, measures)
				if len(scalars) == 0 {
					continue
				}
				for key, item := range scalars {
					row[key] = formatMetric(item)
				}
			default:
				row[firstMeasureOrValue(measures)] = formatMetric(current)
			}
			rows = append(rows, row)
		}
		if len(rows) > 0 {
			return rows
		}

		if len(segmentNames) == 0 {
			for key, item := range value {
				if object, ok := item.(map[string]any); ok {
					row := map[string]string{"country": key}
					scalars := scalarMeasureFields(object, measures)
					if len(scalars) == 0 {
						continue
					}
					for scalarKey, scalarValue := range scalars {
						row[scalarKey] = formatMetric(scalarValue)
					}
					rows = append(rows, row)
					continue
				}
				if isScalarMetric(item) {
					rows = append(rows, map[string]string{
						"country":                     key,
						firstMeasureOrValue(measures): formatMetric(item),
					})
				}
			}
		}
		return rows
	case []any:
		rows := make([]map[string]string, 0, len(value))
		for _, item := range value {
			object, ok := item.(map[string]any)
			if !ok {
				continue
			}
			row := rowFromMetricObject(object, segmentNames, measures)
			if len(row) == 0 {
				continue
			}
			rows = append(rows, row)
		}
		return rows
	default:
		return nil
	}
}

func rowsFromSummaryTotalArray(total []any, segments []segmentInfo, measures []string) []map[string]string {
	if len(total) == 0 || len(total) != len(segments) {
		return nil
	}

	rows := make([]map[string]string, 0, len(total))
	for index, item := range total {
		row := map[string]string{"country": segments[index].DisplayName}
		switch current := item.(type) {
		case map[string]any:
			scalars := scalarMeasureFields(current, measures)
			delete(scalars, "country")
			delete(scalars, "id")
			delete(scalars, "display_name")
			for key, value := range scalars {
				row[key] = formatMetric(value)
			}
		default:
			row[firstMeasureOrValue(measures)] = formatMetric(current)
		}
		if len(row) == 1 {
			continue
		}
		rows = append(rows, row)
	}
	return rows
}

func rowsFromMeasureArrays(total map[string]any, segments []segmentInfo, measures []string) []map[string]string {
	if len(segments) == 0 || len(total) == 0 {
		return nil
	}
	rows := make([]map[string]string, len(segments))
	for index, segment := range segments {
		rows[index] = map[string]string{"country": segment.DisplayName}
	}
	seenMeasure := false
	for key, value := range total {
		if len(measures) > 0 && !stringInSlice(key, measures) {
			continue
		}
		items, ok := value.([]any)
		if !ok || len(items) != len(segments) {
			return nil
		}
		for index, item := range items {
			rows[index][key] = formatMetric(item)
		}
		seenMeasure = true
	}
	if !seenMeasure {
		return nil
	}
	return rows
}

func extractSegmentInfo(raw any) ([]segmentInfo, map[string]string) {
	items, ok := raw.([]any)
	if !ok {
		return nil, map[string]string{}
	}
	segments := make([]segmentInfo, 0, len(items))
	names := make(map[string]string, len(items))
	for _, item := range items {
		object, ok := item.(map[string]any)
		if !ok {
			continue
		}
		id := firstString(object["id"], object["name"])
		if id == "" {
			continue
		}
		displayName := chooseNonEmpty(firstString(object["display_name"]), id)
		segments = append(segments, segmentInfo{ID: id, DisplayName: displayName})
		names[id] = displayName
	}
	return segments, names
}

func rowFromMetricObject(object map[string]any, segmentNames map[string]string, measures []string) map[string]string {
	segmentID := firstString(object["segment_id"], object["country"], object["country_id"], object["id"])
	if segmentID == "" {
		return nil
	}
	row := map[string]string{"country": chooseNonEmpty(segmentNames[segmentID], segmentID)}

	scalars := scalarMeasureFields(object, measures)
	delete(scalars, "segment_id")
	delete(scalars, "country")
	delete(scalars, "country_id")
	delete(scalars, "id")
	delete(scalars, "display_name")
	delete(scalars, "date")
	delete(scalars, "timestamp")
	delete(scalars, "time")
	for key, value := range scalars {
		row[key] = formatMetric(value)
	}
	if len(row) == 1 {
		return nil
	}
	return row
}

func scalarMeasureFields(object map[string]any, measures []string) map[string]any {
	fields := map[string]any{}
	for _, measure := range measures {
		if value, ok := object[measure]; ok && isScalarMetric(value) {
			fields[measure] = value
		}
	}
	if len(fields) > 0 {
		return fields
	}
	for key, value := range object {
		if !isScalarMetric(value) {
			continue
		}
		fields[key] = value
	}
	return fields
}

func measureIDs(raw any) []string {
	items, ok := raw.([]any)
	if !ok {
		return nil
	}
	measures := make([]string, 0, len(items))
	for _, item := range items {
		object, ok := item.(map[string]any)
		if !ok {
			continue
		}
		id := firstString(object["id"], object["name"], object["key"], object["display_name"])
		if id == "" {
			continue
		}
		measures = append(measures, id)
	}
	return measures
}

func loadChartOptionSupport(ctx context.Context, client *rcapi.Client, projectID, chartName string) (chartOptionSupport, string, error) {
	result, err := client.Do(ctx, rcapi.Request{
		Method:    http.MethodGet,
		Path:      fmt.Sprintf("projects/%s/charts/%s/options", projectID, chartName),
		RetryMode: rcapi.RetryDefault,
	})
	if err != nil {
		return chartOptionSupport{}, "", err
	}
	root, ok := result.Payload.(map[string]any)
	if !ok {
		return chartOptionSupport{}, result.RequestID, fmt.Errorf("unexpected chart options payload for %s", chartName)
	}
	return chartOptionSupport{
		Segments: collectOptionIDs(root["segments"]),
		Filters:  collectOptionIDs(root["filters"]),
	}, result.RequestID, nil
}

func collectOptionIDs(raw any) map[string]struct{} {
	items, ok := raw.([]any)
	if !ok {
		return map[string]struct{}{}
	}
	result := make(map[string]struct{}, len(items))
	for _, item := range items {
		object, ok := item.(map[string]any)
		if !ok {
			continue
		}
		id := firstString(object["id"], object["name"])
		if id == "" {
			continue
		}
		result[id] = struct{}{}
	}
	return result
}

func validateCountriesSupport(chartName, contextAlias string, support chartOptionSupport, flags countriesFlags) error {
	if len(support.Segments) > 0 {
		if _, ok := support.Segments["country"]; !ok {
			return &CLIError{Code: exitcode.Usage, Message: fmt.Sprintf("chart %q in context %q does not support country segmentation", chartName, contextAlias)}
		}
	}
	for _, validation := range []struct {
		flagName   string
		filterName string
		required   bool
	}{
		{flagName: "--app", filterName: "app_id", required: strings.TrimSpace(flags.appID) != ""},
		{flagName: "--store", filterName: "store", required: strings.TrimSpace(flags.store) != ""},
		{flagName: "--apple-claim-type", filterName: "apple_claim_type", required: strings.TrimSpace(flags.appleClaimType) != "" || flags.organicOnly},
	} {
		if !validation.required {
			continue
		}
		if len(support.Filters) == 0 {
			return &CLIError{Code: exitcode.Usage, Message: fmt.Sprintf("chart %q in context %q did not expose filter schema; cannot validate %s", chartName, contextAlias, validation.flagName)}
		}
		if _, ok := support.Filters[validation.filterName]; !ok {
			return &CLIError{Code: exitcode.Usage, Message: fmt.Sprintf("chart %q in context %q does not support %s", chartName, contextAlias, validation.flagName)}
		}
	}
	return nil
}

func validateUnsupportedChartParams(chartName, contextAlias string, payload any, flags countriesFlags) error {
	root, ok := payload.(map[string]any)
	if !ok {
		return nil
	}
	unsupported, ok := root["unsupported_params"].(map[string]any)
	if !ok {
		return nil
	}
	filterList, ok := unsupported["filters"].([]any)
	if !ok || len(filterList) == 0 {
		return nil
	}
	unsupportedFilters := make(map[string]struct{}, len(filterList))
	for _, item := range filterList {
		name, ok := item.(string)
		if !ok || name == "" {
			continue
		}
		unsupportedFilters[name] = struct{}{}
	}
	for _, validation := range []struct {
		flagName   string
		filterName string
		required   bool
	}{
		{flagName: "--app", filterName: "app_id", required: strings.TrimSpace(flags.appID) != ""},
		{flagName: "--store", filterName: "store", required: strings.TrimSpace(flags.store) != ""},
		{flagName: "--apple-claim-type", filterName: "apple_claim_type", required: strings.TrimSpace(flags.appleClaimType) != "" || flags.organicOnly},
	} {
		if !validation.required {
			continue
		}
		if _, ok := unsupportedFilters[validation.filterName]; ok {
			return &CLIError{Code: exitcode.Usage, Message: fmt.Sprintf("chart %q in context %q rejected %s", chartName, contextAlias, validation.flagName)}
		}
	}
	return nil
}

func upsertFilter(filters []any, name string, values []string) []any {
	for index, item := range filters {
		object, ok := item.(map[string]any)
		if !ok {
			continue
		}
		currentName := firstString(object["name"], object["id"])
		if currentName != name {
			continue
		}
		object["name"] = name
		object["values"] = values
		filters[index] = object
		return filters
	}
	return append(filters, map[string]any{
		"name":   name,
		"values": values,
	})
}

func appMatchesBundleID(app map[string]any, bundleID string) bool {
	want := strings.TrimSpace(bundleID)
	candidates := []string{
		nestedString(app, "app_store", "bundle_id"),
		nestedString(app, "mac_app_store", "bundle_id"),
		nestedString(app, "play_store", "package_name"),
		nestedString(app, "amazon", "package_name"),
	}
	for _, candidate := range candidates {
		if candidate == want {
			return true
		}
	}
	return false
}

func nestedString(object map[string]any, keys ...string) string {
	current := any(object)
	for _, key := range keys {
		next, ok := current.(map[string]any)
		if !ok {
			return ""
		}
		current = next[key]
	}
	value, _ := current.(string)
	return value
}

func orderedCountryHeaders(rows []map[string]string) []string {
	priority := []string{"context_alias", "project_id", "country", "revenue", "transactions", "value"}
	seen := map[string]struct{}{}
	headers := make([]string, 0)
	for _, row := range rows {
		for _, key := range priority {
			if _, ok := row[key]; ok {
				if _, exists := seen[key]; !exists {
					headers = append(headers, key)
					seen[key] = struct{}{}
				}
			}
		}
		for key := range row {
			if _, exists := seen[key]; exists {
				continue
			}
			headers = append(headers, key)
			seen[key] = struct{}{}
		}
	}
	return headers
}

func stringInSlice(needle string, haystack []string) bool {
	for _, item := range haystack {
		if item == needle {
			return true
		}
	}
	return false
}

func firstMeasureOrValue(measures []string) string {
	if len(measures) == 0 {
		return "value"
	}
	return measures[0]
}

func firstString(values ...any) string {
	for _, value := range values {
		if text, ok := value.(string); ok && strings.TrimSpace(text) != "" {
			return text
		}
	}
	return ""
}

func chooseNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func isScalarMetric(value any) bool {
	switch value.(type) {
	case string, float64, float32, int, int64, int32, int16, int8, uint, uint64, uint32, uint16, uint8:
		return true
	default:
		return false
	}
}

func toFloat(value any) (float64, bool) {
	switch current := value.(type) {
	case float64:
		return current, true
	case float32:
		return float64(current), true
	case int:
		return float64(current), true
	case int64:
		return float64(current), true
	case int32:
		return float64(current), true
	case int16:
		return float64(current), true
	case int8:
		return float64(current), true
	case uint:
		return float64(current), true
	case uint64:
		return float64(current), true
	case uint32:
		return float64(current), true
	case uint16:
		return float64(current), true
	case uint8:
		return float64(current), true
	case string:
		parsed, err := strconv.ParseFloat(current, 64)
		if err != nil {
			return 0, false
		}
		return parsed, true
	default:
		return 0, false
	}
}

func formatMetric(value any) string {
	switch current := value.(type) {
	case string:
		return current
	default:
		number, ok := toFloat(current)
		if !ok {
			return fmt.Sprint(current)
		}
		if math.Mod(number, 1) == 0 {
			return strconv.FormatInt(int64(number), 10)
		}
		return strconv.FormatFloat(number, 'f', -1, 64)
	}
}
