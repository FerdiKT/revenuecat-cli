package cli

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/FerdiKT/revenuecat-cli/internal/config"
	"github.com/zalando/go-keyring"
)

func TestMetricsChartSupportsJSONFlags(t *testing.T) {
	tempConfig := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tempConfig)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/projects/proj_123/charts/customers_new" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("filters"); got != `[{"name":"app_id","values":["appbafaf866e4"]},{"name":"store","values":["app_store"]}]` {
			t.Fatalf("filters = %q", got)
		}
		if got := r.URL.Query().Get("selectors"); got != `{"revenue_type":"revenue"}` {
			t.Fatalf("selectors = %q", got)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"chart_name": "customers_new",
			"values":     []any{map[string]any{"value": 12}},
		})
	}))
	defer server.Close()

	writeTestConfig(t, tempConfig, server.URL)

	cmd, _ := newRootCommand()
	stdout, _, err := executeCommand(t, cmd, []string{
		"metrics", "chart", "customers-new",
		"--context", "prod",
		"--filters-json", `[{"name":"app_id","values":["appbafaf866e4"]},{"name":"store","values":["app_store"]}]`,
		"--selectors-json", `{"revenue_type":"revenue"}`,
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !strings.Contains(stdout, `"chart_name": "customers_new"`) {
		t.Fatalf("stdout = %s", stdout)
	}
}

func TestAppsResolveByBundleID(t *testing.T) {
	server := newRevenueCatFixtureServer()
	defer server.Close()

	tempConfig := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tempConfig)
	writeTestConfig(t, tempConfig, server.URL)

	cmd, _ := newRootCommand()
	stdout, _, err := executeCommand(t, cmd, []string{
		"apps", "resolve",
		"--context", "prod",
		"--bundle-id", "app.ferdi.headson",
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if strings.TrimSpace(stdout) != "app_1" {
		t.Fatalf("stdout = %q, want app_1", stdout)
	}
}

func TestAppsDeleteRequiresExactConfirmation(t *testing.T) {
	tempConfig := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tempConfig)
	writeTestConfig(t, tempConfig, "https://example.invalid")

	cmd, _ := newRootCommand()
	_, _, err := executeCommand(t, cmd, []string{
		"apps", "delete", "app_1",
		"--context", "prod",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	var cliErr *CLIError
	if !errorAsCLI(err, &cliErr) {
		t.Fatalf("err = %T, want *CLIError", err)
	}
	if cliErr.Code != 2 {
		t.Fatalf("cliErr.Code = %d, want 2", cliErr.Code)
	}
	if !strings.Contains(cliErr.Message, "destructive delete requires --confirm app_1") {
		t.Fatalf("message = %q", cliErr.Message)
	}
}

func TestAppsDeleteCallsDeleteEndpointWithConfirmation(t *testing.T) {
	var deleted bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/projects/proj_123/apps/app_1" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodDelete {
			t.Fatalf("method = %s, want DELETE", r.Method)
		}
		deleted = true
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	tempConfig := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tempConfig)
	writeTestConfig(t, tempConfig, server.URL)

	cmd, _ := newRootCommand()
	stdout, _, err := executeCommand(t, cmd, []string{
		"apps", "delete", "app_1",
		"--context", "prod",
		"--confirm", "app_1",
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !deleted {
		t.Fatal("expected delete request")
	}
	if !strings.Contains(stdout, "app deleted") {
		t.Fatalf("stdout = %q", stdout)
	}
}

func TestAppsDeleteSupportsJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/projects/proj_123/apps/app_1" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	tempConfig := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tempConfig)
	writeTestConfig(t, tempConfig, server.URL)

	cmd, _ := newRootCommand()
	stdout, _, err := executeCommand(t, cmd, []string{
		"apps", "delete", "app_1",
		"--context", "prod",
		"--confirm", "app_1",
		"--json",
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !strings.Contains(stdout, `"ok": true`) {
		t.Fatalf("stdout = %s", stdout)
	}
}

func TestAppsPublicKeys(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/projects/proj_123/apps/app_1/public_api_keys" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"object": "list",
			"items": []any{
				map[string]any{
					"object":      "public_api_key",
					"id":          "apikey_1",
					"key":         "appl_123",
					"environment": "production",
					"app_id":      "app_1",
				},
			},
		})
	}))
	defer server.Close()

	tempConfig := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tempConfig)
	writeTestConfig(t, tempConfig, server.URL)

	cmd, _ := newRootCommand()
	stdout, _, err := executeCommand(t, cmd, []string{
		"apps", "public-keys", "app_1",
		"--context", "prod",
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !strings.Contains(stdout, `"key": "appl_123"`) {
		t.Fatalf("stdout = %s", stdout)
	}
}

func TestAppsStoreKitConfig(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/projects/proj_123/apps/app_1/store_kit_config" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"object": "store_kit_config_file",
			"contents": map[string]any{
				"identifier": "app_1",
			},
		})
	}))
	defer server.Close()

	tempConfig := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tempConfig)
	writeTestConfig(t, tempConfig, server.URL)

	cmd, _ := newRootCommand()
	stdout, _, err := executeCommand(t, cmd, []string{
		"apps", "storekit-config", "app_1",
		"--context", "prod",
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !strings.Contains(stdout, `"object": "store_kit_config_file"`) {
		t.Fatalf("stdout = %s", stdout)
	}
}

func TestPaywallsList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/projects/proj_123/paywalls" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Fatalf("method = %s, want GET", r.Method)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"object": "list",
			"items": []any{
				map[string]any{
					"id":           "paywall_1",
					"display_name": "Main Paywall",
				},
			},
		})
	}))
	defer server.Close()

	tempConfig := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tempConfig)
	writeTestConfig(t, tempConfig, server.URL)

	cmd, _ := newRootCommand()
	stdout, _, err := executeCommand(t, cmd, []string{
		"paywalls", "list",
		"--context", "prod",
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !strings.Contains(stdout, `"id": "paywall_1"`) {
		t.Fatalf("stdout = %s", stdout)
	}
}

func TestPaywallsCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/projects/proj_123/paywalls" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("Decode: %v", err)
		}
		if body["display_name"] != "Main Paywall" {
			t.Fatalf("body = %#v", body)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":           "paywall_1",
			"display_name": "Main Paywall",
		})
	}))
	defer server.Close()

	tempConfig := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tempConfig)
	writeTestConfig(t, tempConfig, server.URL)

	cmd, _ := newRootCommand()
	stdout, _, err := executeCommand(t, cmd, []string{
		"paywalls", "create",
		"--context", "prod",
		"--data", `{"display_name":"Main Paywall"}`,
		"--json",
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !strings.Contains(stdout, `"id": "paywall_1"`) {
		t.Fatalf("stdout = %s", stdout)
	}
}

func TestPaywallsDeleteRequiresExactConfirmation(t *testing.T) {
	tempConfig := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tempConfig)
	writeTestConfig(t, tempConfig, "https://example.invalid")

	cmd, _ := newRootCommand()
	_, _, err := executeCommand(t, cmd, []string{
		"paywalls", "delete", "paywall_1",
		"--context", "prod",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	var cliErr *CLIError
	if !errorAsCLI(err, &cliErr) {
		t.Fatalf("err = %T, want *CLIError", err)
	}
	if cliErr.Code != 2 {
		t.Fatalf("cliErr.Code = %d, want 2", cliErr.Code)
	}
	if !strings.Contains(cliErr.Message, "destructive delete requires --confirm paywall_1") {
		t.Fatalf("message = %q", cliErr.Message)
	}
}

func TestPaywallsDeleteCallsDeleteEndpointWithConfirmation(t *testing.T) {
	var deleted bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/projects/proj_123/paywalls/paywall_1" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodDelete {
			t.Fatalf("method = %s, want DELETE", r.Method)
		}
		deleted = true
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	tempConfig := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tempConfig)
	writeTestConfig(t, tempConfig, server.URL)

	cmd, _ := newRootCommand()
	stdout, _, err := executeCommand(t, cmd, []string{
		"paywalls", "delete", "paywall_1",
		"--context", "prod",
		"--confirm", "paywall_1",
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !deleted {
		t.Fatal("expected delete request")
	}
	if !strings.Contains(stdout, "paywall deleted") {
		t.Fatalf("stdout = %q", stdout)
	}
}

func TestMetricsCountriesOutputsTable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v2/projects/proj_123/charts/revenue/options":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"object": "chart_options",
				"segments": []any{
					map[string]any{"id": "country", "display_name": "Country"},
				},
				"filters": []any{
					map[string]any{"id": "app_id", "display_name": "App"},
					map[string]any{"id": "store", "display_name": "Store"},
					map[string]any{"id": "apple_claim_type", "display_name": "Apple Claim Type"},
				},
			})
		case "/v2/projects/proj_123/charts/revenue":
			if got := r.URL.Query().Get("aggregate"); got != "total" {
				t.Fatalf("aggregate = %q, want total", got)
			}
			if got := r.URL.Query().Get("segment"); got != "country" {
				t.Fatalf("segment = %q, want country", got)
			}
			if got := r.URL.Query().Get("filters"); got != `[{"name":"app_id","values":["app_1"]},{"name":"store","values":["app_store"]},{"name":"apple_claim_type","values":["Organic"]}]` {
				t.Fatalf("filters = %q", got)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"chart_name": "revenue",
				"segments": []any{
					map[string]any{"id": "US", "display_name": "United States"},
					map[string]any{"id": "TR", "display_name": "Turkey"},
					map[string]any{"id": "OTHER", "display_name": "Other"},
				},
				"measures": []any{
					map[string]any{"id": "revenue"},
					map[string]any{"id": "transactions"},
				},
				"summary": map[string]any{
					"total": []any{
						map[string]any{"revenue": 1234.5, "transactions": 18},
						map[string]any{"revenue": 87, "transactions": 4},
						map[string]any{"revenue": 11, "transactions": 1},
					},
				},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	tempConfig := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tempConfig)
	writeTestConfig(t, tempConfig, server.URL)

	cmd, _ := newRootCommand()
	stdout, _, err := executeCommand(t, cmd, []string{
		"metrics", "countries", "revenue",
		"--context", "prod",
		"--app", "app_1",
		"--store", "app_store",
		"--organic-only",
		"--start", "2026-01-01",
		"--end", "2026-04-16",
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !strings.Contains(stdout, "country") || !strings.Contains(stdout, "revenue") || !strings.Contains(stdout, "transactions") {
		t.Fatalf("stdout headers missing: %s", stdout)
	}
	if !strings.Contains(stdout, "United States") || !strings.Contains(stdout, "1234.5") {
		t.Fatalf("stdout missing expected row: %s", stdout)
	}
	if strings.Contains(stdout, "Other") {
		t.Fatalf("stdout should exclude Other by default: %s", stdout)
	}
}

func TestMetricsCountriesRejectsUnsupportedAppFilter(t *testing.T) {
	tempConfig := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tempConfig)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/projects/proj_123/charts/customers_active/options" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"object": "chart_options",
			"segments": []any{
				map[string]any{"id": "country", "display_name": "Country"},
			},
			"filters": []any{
				map[string]any{"id": "apple_claim_type", "display_name": "Apple Claim Type"},
			},
		})
	}))
	defer server.Close()

	writeTestConfig(t, tempConfig, server.URL)

	cmd, _ := newRootCommand()
	_, _, err := executeCommand(t, cmd, []string{
		"metrics", "countries", "customers-active",
		"--context", "prod",
		"--app", "app_1",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	var cliErr *CLIError
	if !errorAsCLI(err, &cliErr) {
		t.Fatalf("err = %T, want *CLIError", err)
	}
	if cliErr.Code != 2 {
		t.Fatalf("cliErr.Code = %d, want 2", cliErr.Code)
	}
	if !strings.Contains(cliErr.Message, "does not support --app") {
		t.Fatalf("message = %q", cliErr.Message)
	}
}

func writeTestConfig(t *testing.T, baseDir, serverURL string) {
	t.Helper()
	keyring.MockInit()

	store, err := config.NewStore(filepath.Join(baseDir, "revenuecat", "config.json"))
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	if err := store.Save(&config.Config{
		ActiveContext: "prod",
		Contexts: []config.Context{{
			Alias:      "prod",
			APIKey:     "sk_test",
			ProjectID:  "proj_123",
			APIBaseURL: serverURL + "/v2",
		}},
	}); err != nil {
		t.Fatalf("Save: %v", err)
	}
}
