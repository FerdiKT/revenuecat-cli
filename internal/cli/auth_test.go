package cli

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/FerdiKT/revenuecat-cli/internal/config"
)

func TestAuthLoginComingSoon(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	cmd, _ := newRootCommand()
	stdout, stderr, err := executeCommand(t, cmd, []string{"auth", "login"})

	if stdout != "" {
		t.Fatalf("stdout = %q, want empty", stdout)
	}
	var cliErr *CLIError
	if !errorAsCLI(err, &cliErr) {
		t.Fatalf("err = %T, want *CLIError", err)
	}
	if cliErr.Code != 8 {
		t.Fatalf("cliErr.Code = %d, want 8", cliErr.Code)
	}
	if !strings.Contains(stderr, "coming soon") {
		t.Fatalf("stderr = %q, want coming soon message", stderr)
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
