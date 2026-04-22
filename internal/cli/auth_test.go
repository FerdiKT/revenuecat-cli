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
