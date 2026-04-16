package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/FerdiKT/revenuecat-cli/internal/config"
	"github.com/FerdiKT/revenuecat-cli/internal/exitcode"
	"github.com/FerdiKT/revenuecat-cli/internal/output"
	"github.com/FerdiKT/revenuecat-cli/internal/rcapi"
	"github.com/spf13/cobra"
)

type requestFlags struct {
	data          string
	dataFile      string
	expand        []string
	params        []string
	limit         int
	startingAfter string
}

type fanoutResult struct {
	ContextAlias string `json:"context_alias"`
	ProjectID    string `json:"project_id,omitempty"`
	OK           bool   `json:"ok"`
	RequestID    string `json:"request_id,omitempty"`
	Data         any    `json:"data,omitempty"`
	Error        any    `json:"error,omitempty"`
}

func (a *App) loadConfig() (*config.Config, error) {
	cfg, err := a.store.Load()
	if err != nil {
		return nil, &CLIError{Code: exitcode.Internal, Message: err.Error()}
	}
	return cfg, nil
}

func (a *App) saveConfig(cfg *config.Config) error {
	if err := a.store.Save(cfg); err != nil {
		return &CLIError{Code: exitcode.Internal, Message: err.Error()}
	}
	return nil
}

func (a *App) resolveSingleContext(cfg *config.Config) (*config.Context, error) {
	if a.globalFlags.AllContexts {
		return nil, &CLIError{Code: exitcode.Usage, Message: "--all-contexts is only valid for read commands"}
	}

	alias := strings.TrimSpace(a.globalFlags.ContextAlias)
	if alias == "" {
		alias = strings.TrimSpace(cfg.ActiveContext)
	}
	if alias == "" {
		return nil, &CLIError{Code: exitcode.Context, Message: "no context selected; use `revenuecat contexts use <alias>` or pass --context"}
	}

	ctx, ok := cfg.FindContext(alias)
	if !ok {
		return nil, &CLIError{Code: exitcode.Context, Message: fmt.Sprintf("context %q not found", alias)}
	}
	return ctx, nil
}

func (a *App) resolveReadContexts(cfg *config.Config) ([]config.Context, error) {
	if a.globalFlags.AllContexts {
		if len(cfg.Contexts) == 0 {
			return nil, &CLIError{Code: exitcode.Context, Message: "no contexts configured"}
		}
		return append([]config.Context(nil), cfg.Contexts...), nil
	}

	ctx, err := a.resolveSingleContext(cfg)
	if err != nil {
		return nil, err
	}
	return []config.Context{*ctx}, nil
}

func (a *App) clientFor(ctx config.Context) *rcapi.Client {
	return rcapi.NewClient(ctx.APIKey, ctx.APIBaseURL)
}

func (a *App) requestRetryMode(method string) rcapi.RetryMode {
	if method == http.MethodGet {
		return rcapi.RetryDefault
	}
	if a.globalFlags.Retry {
		return rcapi.RetryForced
	}
	return rcapi.RetryDisabled
}

func (a *App) outputContext(ctx config.Context) *output.ContextSummary {
	return &output.ContextSummary{
		Alias:     ctx.Alias,
		ProjectID: ctx.ProjectID,
	}
}

func (a *App) ensureProjectID(ctx config.Context) (string, error) {
	if strings.TrimSpace(ctx.ProjectID) == "" {
		return "", &CLIError{
			Code:    exitcode.Context,
			Message: fmt.Sprintf("context %q has no project_id; set it during `contexts add` or run `revenuecat contexts verify %s`", ctx.Alias, ctx.Alias),
		}
	}
	return ctx.ProjectID, nil
}

func (a *App) renderRead(ctx config.Context, result *rcapi.Result) error {
	meta := output.Meta{RequestID: result.RequestID}
	if pagination := extractPagination(result.Payload); pagination != nil {
		meta.Pagination = pagination
	}
	envelope := output.Success(a.outputContext(ctx), result.Payload, meta)

	if shouldTable(a.globalFlags) {
		rows := toTableRows(result.Payload)
		return output.PrintTable(os.Stdout, rows)
	}

	return output.PrintJSON(os.Stdout, envelope)
}

func (a *App) renderMutation(ctx config.Context, result *rcapi.Result, action string) error {
	if a.globalFlags.JSON {
		meta := output.Meta{RequestID: result.RequestID}
		envelope := output.Success(a.outputContext(ctx), result.Payload, meta)
		return output.PrintJSON(os.Stdout, envelope)
	}

	if id := extractID(result.Payload); id != "" {
		_, err := fmt.Fprintf(os.Stdout, "%s: %s\n", action, id)
		return err
	}
	_, err := fmt.Fprintln(os.Stdout, action)
	return err
}

func shouldTable(flags GlobalFlags) bool {
	return strings.EqualFold(flags.Format, "table")
}

func parseBody(data, file string) (any, error) {
	switch {
	case data != "" && file != "":
		return nil, &CLIError{Code: exitcode.Usage, Message: "use only one of --data or --file"}
	case data == "" && file == "":
		return nil, &CLIError{Code: exitcode.Usage, Message: "one of --data or --file is required"}
	}

	raw := data
	if file != "" {
		bytes, err := os.ReadFile(file)
		if err != nil {
			return nil, &CLIError{Code: exitcode.Usage, Message: fmt.Sprintf("read %s: %v", file, err)}
		}
		raw = string(bytes)
	}

	var payload any
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return nil, &CLIError{Code: exitcode.Usage, Message: fmt.Sprintf("invalid JSON body: %v", err)}
	}
	return payload, nil
}

func parseQuery(flags requestFlags) (url.Values, error) {
	query := url.Values{}
	if flags.limit > 0 {
		query.Set("limit", fmt.Sprintf("%d", flags.limit))
	}
	if flags.startingAfter != "" {
		query.Set("starting_after", flags.startingAfter)
	}
	for _, expand := range flags.expand {
		query.Add("expand", expand)
	}
	for _, param := range flags.params {
		key, value, ok := strings.Cut(param, "=")
		if !ok || strings.TrimSpace(key) == "" {
			return nil, &CLIError{Code: exitcode.Usage, Message: fmt.Sprintf("invalid --param %q, expected key=value", param)}
		}
		query.Add(strings.TrimSpace(key), strings.TrimSpace(value))
	}
	return query, nil
}

func (a *App) runReadAcrossContexts(contexts []config.Context, request func(config.Context) (*rcapi.Result, error)) error {
	if len(contexts) == 1 {
		result, err := request(contexts[0])
		if err != nil {
			var apiErr *rcapi.APIError
			if errors.As(err, &apiErr) {
				return a.mapAPIError(&contexts[0], apiErr, "")
			}
			return &CLIError{Code: exitcode.Internal, Message: err.Error()}
		}
		return a.renderRead(contexts[0], result)
	}

	results := make([]fanoutResult, 0, len(contexts))
	var firstAPIError *rcapi.APIError
	for _, current := range contexts {
		result, err := request(current)
		if err != nil {
			entry := fanoutResult{ContextAlias: current.Alias, ProjectID: current.ProjectID, OK: false}
			var apiErr *rcapi.APIError
			if errors.As(err, &apiErr) {
				if firstAPIError == nil {
					firstAPIError = apiErr
				}
				entry.Error = map[string]any{
					"type":        apiErr.Type,
					"message":     apiErr.Message,
					"status_code": apiErr.StatusCode,
					"retryable":   apiErr.Retryable,
				}
			} else {
				entry.Error = map[string]any{"message": err.Error()}
			}
			results = append(results, entry)
			continue
		}

		results = append(results, fanoutResult{
			ContextAlias: current.Alias,
			ProjectID:    current.ProjectID,
			OK:           true,
			RequestID:    result.RequestID,
			Data:         result.Payload,
		})
	}

	if firstAPIError != nil {
		envelope := output.Failure(nil, output.Meta{}, &output.ErrorPayload{
			Type:       "fanout_partial_failure",
			Message:    "one or more contexts failed",
			StatusCode: firstAPIError.StatusCode,
			Retryable:  firstAPIError.Retryable,
		})
		envelope.Data = map[string]any{"results": results}
		if err := output.PrintJSON(os.Stdout, envelope); err != nil {
			return &CLIError{Code: exitcode.Internal, Message: err.Error()}
		}
		return classifyAPIError(firstAPIError)
	}

	return output.PrintJSON(os.Stdout, output.Success(nil, map[string]any{"results": results}, output.Meta{}))
}

func addRequestBodyFlags(cmd *cobra.Command, flags *requestFlags) {
	cmd.Flags().StringVar(&flags.data, "data", "", "Inline JSON request body")
	cmd.Flags().StringVar(&flags.dataFile, "file", "", "Path to a JSON request body file")
}

func addReadFlags(cmd *cobra.Command, flags *requestFlags) {
	cmd.Flags().StringSliceVar(&flags.expand, "expand", nil, "Expandable fields to include")
	cmd.Flags().StringSliceVar(&flags.params, "param", nil, "Additional query parameter in key=value form")
	cmd.Flags().IntVar(&flags.limit, "limit", 100, "Maximum number of items to fetch")
	cmd.Flags().StringVar(&flags.startingAfter, "starting-after", "", "Pagination cursor")
}

func addCommonIDFlag(cmd *cobra.Command, name, usage string) *string {
	value := ""
	cmd.Flags().StringVar(&value, name, "", usage)
	_ = cmd.MarkFlagRequired(name)
	return &value
}

func (a *App) mapAPIError(ctx *config.Context, apiErr *rcapi.APIError, requestID string) error {
	if apiErr == nil {
		return nil
	}
	meta := output.Meta{RequestID: requestID}
	errPayload := &output.ErrorPayload{
		Type:       apiErr.Type,
		Message:    apiErr.Message,
		StatusCode: apiErr.StatusCode,
		Retryable:  apiErr.Retryable,
		DocURL:     apiErr.DocURL,
		BackoffMS:  apiErr.BackoffMS,
	}
	if a.globalFlags.JSON || shouldTable(a.globalFlags) == false {
		summary := (*output.ContextSummary)(nil)
		if ctx != nil {
			summary = a.outputContext(*ctx)
		}
		_ = output.PrintJSON(os.Stdout, output.Failure(summary, meta, errPayload))
	}

	return classifyAPIError(apiErr)
}

func extractPagination(payload any) any {
	if object, ok := payload.(map[string]any); ok {
		if object["object"] == "list" {
			result := map[string]any{}
			if nextPage, ok := object["next_page"]; ok {
				result["next_page"] = nextPage
			}
			if urlValue, ok := object["url"]; ok {
				result["url"] = urlValue
			}
			if len(result) > 0 {
				return result
			}
		}
	}
	return nil
}

func extractID(payload any) string {
	if object, ok := payload.(map[string]any); ok {
		if id, ok := object["id"].(string); ok {
			return id
		}
	}
	return ""
}

func toTableRows(payload any) []map[string]string {
	rows := make([]map[string]string, 0)
	switch value := payload.(type) {
	case map[string]any:
		if items, ok := value["items"].([]any); ok {
			for _, item := range items {
				rows = append(rows, scalarRow(item))
			}
			return rows
		}
		return []map[string]string{scalarRow(value)}
	case []any:
		for _, item := range value {
			rows = append(rows, scalarRow(item))
		}
		return rows
	default:
		return []map[string]string{{"value": fmt.Sprint(value)}}
	}
}

func scalarRow(item any) map[string]string {
	row := map[string]string{}
	object, ok := item.(map[string]any)
	if !ok {
		row["value"] = fmt.Sprint(item)
		return row
	}

	keys := make([]string, 0, len(object))
	for key := range object {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		switch value := object[key].(type) {
		case nil:
			row[key] = ""
		case string, float64, bool:
			row[key] = fmt.Sprint(value)
		default:
			bytes, err := json.Marshal(value)
			if err != nil {
				row[key] = fmt.Sprint(value)
			} else {
				row[key] = string(bytes)
			}
		}
	}
	return row
}

func discoverProject(ctx context.Context, client *rcapi.Client) (map[string]any, error) {
	result, err := client.Do(ctx, rcapi.Request{
		Method:    http.MethodGet,
		Path:      "projects",
		Query:     url.Values{"limit": []string{"1"}},
		RetryMode: rcapi.RetryDefault,
	})
	if err != nil {
		return nil, err
	}

	root, ok := result.Payload.(map[string]any)
	if !ok {
		return nil, errors.New("unexpected projects response")
	}
	items, ok := root["items"].([]any)
	if !ok || len(items) == 0 {
		return nil, errors.New("projects list did not contain any items")
	}
	project, ok := items[0].(map[string]any)
	if !ok {
		return nil, errors.New("unexpected project payload")
	}
	return project, nil
}

func projectNameFromPayload(payload map[string]any) string {
	if value, ok := payload["name"].(string); ok {
		return value
	}
	if value, ok := payload["display_name"].(string); ok {
		return value
	}
	return ""
}

func classifyAPIError(apiErr *rcapi.APIError) error {
	switch apiErr.StatusCode {
	case http.StatusUnauthorized, http.StatusForbidden:
		return &CLIError{Code: exitcode.Auth, Message: ""}
	case http.StatusNotFound:
		return &CLIError{Code: exitcode.NotFound, Message: ""}
	case http.StatusConflict, http.StatusLocked:
		return &CLIError{Code: exitcode.Conflict, Message: ""}
	case http.StatusTooManyRequests, http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return &CLIError{Code: exitcode.Retryable, Message: ""}
	default:
		return &CLIError{Code: exitcode.Internal, Message: ""}
	}
}

func fetchAllListItems(ctx context.Context, client *rcapi.Client, initialPath string, query url.Values) ([]any, []string, error) {
	currentPath := initialPath
	currentQuery := cloneValues(query)
	items := make([]any, 0)
	requestIDs := make([]string, 0)

	for {
		result, err := client.Do(ctx, rcapi.Request{
			Method:    http.MethodGet,
			Path:      currentPath,
			Query:     currentQuery,
			RetryMode: rcapi.RetryDefault,
		})
		if err != nil {
			return nil, requestIDs, err
		}
		requestIDs = append(requestIDs, result.RequestID)

		root, ok := result.Payload.(map[string]any)
		if !ok {
			return nil, requestIDs, errors.New("expected list response object")
		}
		pageItems, ok := root["items"].([]any)
		if !ok {
			return nil, requestIDs, errors.New("expected list response items")
		}
		items = append(items, pageItems...)

		nextPage, ok := root["next_page"].(string)
		if !ok || nextPage == "" {
			break
		}
		nextPath, nextQuery, err := splitNextPage(nextPage)
		if err != nil {
			return nil, requestIDs, err
		}
		currentPath = nextPath
		currentQuery = nextQuery
	}

	return items, requestIDs, nil
}

func cloneValues(values url.Values) url.Values {
	if values == nil {
		return url.Values{}
	}
	cloned := url.Values{}
	for key, items := range values {
		cloned[key] = append([]string(nil), items...)
	}
	return cloned
}

func splitNextPage(nextPage string) (string, url.Values, error) {
	parsed, err := url.Parse(nextPage)
	if err != nil {
		return "", nil, fmt.Errorf("parse next_page: %w", err)
	}

	nextPath := strings.TrimPrefix(parsed.Path, "/")
	parts := strings.Split(nextPath, "/")
	if len(parts) > 0 && parts[0] == "v2" {
		nextPath = path.Join(parts[1:]...)
	}

	return nextPath, parsed.Query(), nil
}
