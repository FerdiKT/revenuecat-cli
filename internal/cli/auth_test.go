package cli

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/FerdiKT/revenuecat-cli/internal/config"
	"github.com/FerdiKT/revenuecat-cli/internal/credentials"
	"github.com/zalando/go-keyring"
)

type writeCapture chan string

func (w writeCapture) Write(data []byte) (int, error) {
	w <- string(data)
	return len(data), nil
}

func TestRunOAuthLoginUsesPKCEAndStoresTokenResponse(t *testing.T) {
	var tokenRequest url.Values
	oauthServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/oauth2/token" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm: %v", err)
		}
		tokenRequest = r.Form
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token":  "atk_test",
			"refresh_token": "rtk_test",
			"token_type":    "Bearer",
			"expires_in":    3600,
			"scope":         "project_configuration:projects:read",
		})
	}))
	defer oauthServer.Close()

	redirectURI := freeRedirectURI(t)
	outputs := make(writeCapture, 8)
	resultCh := make(chan struct {
		token *oauthTokenResponse
		err   error
	}, 1)
	go func() {
		token, err := runOAuthLogin(context.Background(), oauthLoginOptions{
			ClientID:     "client_test",
			RedirectURI:  redirectURI,
			Scopes:       []string{"project_configuration:projects:read"},
			OAuthBaseURL: oauthServer.URL,
			NoOpen:       true,
			Timeout:      5 * time.Second,
			Stdout:       outputs,
		})
		resultCh <- struct {
			token *oauthTokenResponse
			err   error
		}{token: token, err: err}
	}()

	state := stateFromAuthorizeOutput(t, <-outputs)
	resp, err := http.Get(redirectURI + "?code=code_test&state=" + url.QueryEscape(state))
	if err != nil {
		t.Fatalf("callback: %v", err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("callback status = %d", resp.StatusCode)
	}

	result := <-resultCh
	if result.err != nil {
		t.Fatalf("runOAuthLogin: %v", result.err)
	}
	if result.token.AccessToken != "atk_test" {
		t.Fatalf("access token = %q", result.token.AccessToken)
	}
	if tokenRequest.Get("client_id") != "client_test" {
		t.Fatalf("client_id = %q", tokenRequest.Get("client_id"))
	}
	if tokenRequest.Get("code") != "code_test" {
		t.Fatalf("code = %q", tokenRequest.Get("code"))
	}
	if tokenRequest.Get("code_verifier") == "" {
		t.Fatal("expected code_verifier")
	}
	if tokenRequest.Get("client_secret") != "" {
		t.Fatal("did not expect client_secret for public client flow")
	}
}

func TestAuthStatusOutputsJSON(t *testing.T) {
	keyring.MockInit()

	tempConfig := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tempConfig)

	store, err := config.NewStore(filepath.Join(tempConfig, "revenuecat", "config.json"))
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	if err := store.Save(&config.Config{
		ActiveContext: "prod",
		Contexts: []config.Context{
			{Alias: "prod", APIKey: "sk_test", ProjectID: "proj_123"},
		},
	}); err != nil {
		t.Fatalf("Save: %v", err)
	}

	cmd, _ := newRootCommand()
	stdout, _, err := executeCommand(t, cmd, []string{"auth", "status"})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("Unmarshal: %v\n%s", err, stdout)
	}
	if payload["ok"] != true {
		t.Fatalf("payload = %#v", payload)
	}
}

func TestLoadConfigMigratesAPIKeysToCredentialStore(t *testing.T) {
	keyring.MockInit()

	tempConfig := t.TempDir()
	configPath := filepath.Join(tempConfig, "revenuecat", "config.json")
	store, err := config.NewStore(configPath)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	if err := store.Save(&config.Config{
		ActiveContext: "prod",
		Contexts: []config.Context{{
			Alias:     "prod",
			APIKey:    "sk_legacy",
			ProjectID: "proj_123",
		}},
	}); err != nil {
		t.Fatalf("Save: %v", err)
	}

	app := &App{store: store}
	cfg, err := app.loadConfig()
	if err != nil {
		t.Fatalf("loadConfig: %v", err)
	}
	if cfg.Contexts[0].APIKey != "sk_legacy" {
		t.Fatalf("hydrated APIKey = %q", cfg.Contexts[0].APIKey)
	}
	if cfg.Contexts[0].APIKeyStore != credentials.StoreName {
		t.Fatalf("APIKeyStore = %q", cfg.Contexts[0].APIKeyStore)
	}

	stored, found, err := credentials.LoadAPIKey("prod")
	if err != nil {
		t.Fatalf("LoadAPIKey: %v", err)
	}
	if !found || stored != "sk_legacy" {
		t.Fatalf("stored API key = %q, found=%t", stored, found)
	}

	raw, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if strings.Contains(string(raw), "sk_legacy") {
		t.Fatalf("config still contains API key: %s", raw)
	}
}

func TestProjectsListRefreshesOAuthToken(t *testing.T) {
	keyring.MockInit()

	refreshSeen := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth2/token":
			if err := r.ParseForm(); err != nil {
				t.Fatalf("ParseForm: %v", err)
			}
			refreshSeen = true
			if got := r.Form.Get("grant_type"); got != "refresh_token" {
				t.Fatalf("grant_type = %q", got)
			}
			if got := r.Form.Get("refresh_token"); got != "rtk_old" {
				t.Fatalf("refresh_token = %q", got)
			}
			if got := r.Form.Get("client_secret"); got != "" {
				t.Fatalf("client_secret = %q, want empty", got)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"access_token":  "atk_fresh",
				"refresh_token": "rtk_new",
				"token_type":    "Bearer",
				"expires_in":    3600,
				"scope":         "project_configuration:projects:read",
			})
		case "/v2/projects":
			if got := r.Header.Get("Authorization"); got != "Bearer atk_fresh" {
				t.Fatalf("Authorization = %q", got)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"object": "list",
				"items": []any{
					map[string]any{"id": "proj_1", "name": "Project One"},
				},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	tempConfig := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tempConfig)
	store, err := config.NewStore(filepath.Join(tempConfig, "revenuecat", "config.json"))
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	if err := store.Save(&config.Config{
		OAuth: config.OAuth{
			ClientID:     "client_test",
			OAuthBaseURL: server.URL,
			APIBaseURL:   server.URL + "/v2",
			TokenStore:   credentials.StoreName,
		},
	}); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if err := credentials.SaveOAuth(credentials.OAuthToken{
		AccessToken:  "atk_old",
		RefreshToken: "rtk_old",
		TokenType:    "Bearer",
		ExpiresAt:    time.Now().Add(-time.Minute).UTC().Format(time.RFC3339),
		Scopes:       []string{"project_configuration:projects:read"},
	}); err != nil {
		t.Fatalf("SaveOAuth: %v", err)
	}

	cmd, _ := newRootCommand()
	stdout, _, err := executeCommand(t, cmd, []string{"projects", "list"})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !refreshSeen {
		t.Fatal("expected token refresh")
	}
	if !strings.Contains(stdout, `"id": "proj_1"`) {
		t.Fatalf("stdout = %s", stdout)
	}

	token, found, err := credentials.LoadOAuth()
	if err != nil {
		t.Fatalf("LoadOAuth: %v", err)
	}
	if !found || token.AccessToken != "atk_fresh" || token.RefreshToken != "rtk_new" {
		t.Fatalf("token = %#v, found=%t", token, found)
	}
}

func TestResourceCommandSupportsOAuthProjectID(t *testing.T) {
	keyring.MockInit()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/projects/proj_123/apps" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer atk_valid" {
			t.Fatalf("Authorization = %q", got)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"object": "list",
			"items": []any{
				map[string]any{"id": "app_1", "name": "Headsup"},
			},
		})
	}))
	defer server.Close()

	tempConfig := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tempConfig)
	store, err := config.NewStore(filepath.Join(tempConfig, "revenuecat", "config.json"))
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	if err := store.Save(&config.Config{
		OAuth: config.OAuth{
			ClientID:   "client_test",
			APIBaseURL: server.URL + "/v2",
			TokenStore: credentials.StoreName,
		},
	}); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if err := credentials.SaveOAuth(credentials.OAuthToken{
		AccessToken:  "atk_valid",
		RefreshToken: "rtk_valid",
		TokenType:    "Bearer",
		ExpiresAt:    time.Now().Add(time.Hour).UTC().Format(time.RFC3339),
		Scopes:       []string{"project_configuration:apps:read"},
	}); err != nil {
		t.Fatalf("SaveOAuth: %v", err)
	}

	cmd, _ := newRootCommand()
	stdout, _, err := executeCommand(t, cmd, []string{"apps", "list", "--project-id", "proj_123"})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !strings.Contains(stdout, `"id": "app_1"`) {
		t.Fatalf("stdout = %s", stdout)
	}
}

func TestPullBundleBuildsAggregatedSnapshot(t *testing.T) {
	server := newRevenueCatFixtureServer()
	defer server.Close()

	app := &App{}
	ctx := config.Context{
		Alias:      "prod",
		APIKey:     "sk_test",
		ProjectID:  "proj_123",
		APIBaseURL: server.URL + "/v2",
	}

	bundle, requestIDs, err := app.pullBundle(ctx, []string{"trials"}, true, true, true)
	if err != nil {
		t.Fatalf("pullBundle: %v", err)
	}

	if len(requestIDs) == 0 {
		t.Fatal("expected request ids")
	}
	apps := bundle["apps"].([]any)
	firstApp := apps[0].(map[string]any)
	if firstApp["context_alias"] != "prod" || firstApp["project_id"] != "proj_123" {
		t.Fatalf("unexpected app annotation: %#v", firstApp)
	}

	if _, ok := bundle["charts"].(map[string]any)["trials"]; !ok {
		t.Fatalf("expected trials chart in bundle: %#v", bundle["charts"])
	}
	paywalls := bundle["paywalls"].([]any)
	firstPaywall := paywalls[0].(map[string]any)
	if firstPaywall["id"] != "paywall_1" || firstPaywall["context_alias"] != "prod" {
		t.Fatalf("unexpected paywall annotation: %#v", firstPaywall)
	}
	if _, ok := bundle["subscriptions"].(map[string]any)["cust_1"]; !ok {
		t.Fatalf("expected subscriptions keyed by customer: %#v", bundle["subscriptions"])
	}
	if _, ok := bundle["purchases"].(map[string]any)["cust_1"]; !ok {
		t.Fatalf("expected purchases keyed by customer: %#v", bundle["purchases"])
	}
}

func executeCommand(t *testing.T, cmd interface {
	SetArgs([]string)
	Execute() error
}, args []string) (string, string, error) {
	t.Helper()

	stdoutReader, stdoutWriter, err := os.Pipe()
	if err != nil {
		t.Fatalf("Pipe stdout: %v", err)
	}
	defer stdoutReader.Close()

	stderrReader, stderrWriter, err := os.Pipe()
	if err != nil {
		t.Fatalf("Pipe stderr: %v", err)
	}
	defer stderrReader.Close()

	oldStdout := os.Stdout
	oldStderr := os.Stderr
	os.Stdout = stdoutWriter
	os.Stderr = stderrWriter
	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	cmd.SetArgs(args)
	runErr := cmd.Execute()

	_ = stdoutWriter.Close()
	_ = stderrWriter.Close()
	stdoutBytes, _ := io.ReadAll(stdoutReader)
	stderrBytes, _ := io.ReadAll(stderrReader)
	return string(stdoutBytes), strings.TrimSpace(string(stderrBytes)), runErr
}

func errorAsCLI(err error, target **CLIError) bool {
	cliErr, ok := err.(*CLIError)
	if !ok {
		return false
	}
	*target = cliErr
	return true
}

func freeRedirectURI(t *testing.T) string {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Listen: %v", err)
	}
	addr := listener.Addr().String()
	if err := listener.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	return "http://" + addr + "/oauth/callback"
}

func stateFromAuthorizeOutput(t *testing.T, text string) string {
	t.Helper()

	start := strings.Index(text, "http")
	if start < 0 {
		t.Fatalf("authorize output missing URL: %q", text)
	}
	rawURL := strings.TrimSpace(text[start:])
	parsed, err := url.Parse(rawURL)
	if err != nil {
		t.Fatalf("Parse authorize URL: %v", err)
	}
	state := parsed.Query().Get("state")
	if state == "" {
		t.Fatalf("authorize URL missing state: %s", rawURL)
	}
	if parsed.Query().Get("code_challenge") == "" {
		t.Fatalf("authorize URL missing code_challenge: %s", rawURL)
	}
	return state
}
