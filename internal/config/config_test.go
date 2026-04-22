package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStoreSaveLoadAndPermissions(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "config.json")
	store, err := NewStore(path)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	cfg := &Config{
		ActiveContext: "prod",
		Contexts: []Context{
			{
				Alias:       "prod",
				APIKey:      "sk_test_1234",
				ProjectID:   "proj_123",
				ProjectName: "Production",
			},
		},
	}

	if err := store.Save(cfg); err != nil {
		t.Fatalf("Save: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if got, want := info.Mode().Perm(), os.FileMode(0o600); got != want {
		t.Fatalf("mode = %o, want %o", got, want)
	}

	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if loaded.ActiveContext != "prod" {
		t.Fatalf("ActiveContext = %q, want prod", loaded.ActiveContext)
	}
	if len(loaded.Contexts) != 1 || loaded.Contexts[0].ProjectID != "proj_123" {
		t.Fatalf("unexpected contexts: %#v", loaded.Contexts)
	}
	if loaded.Contexts[0].APIBaseURL != DefaultAPIBaseURL {
		t.Fatalf("APIBaseURL = %q, want %q", loaded.Contexts[0].APIBaseURL, DefaultAPIBaseURL)
	}
}

func TestConfigUpsertAndRemoveContext(t *testing.T) {
	t.Parallel()

	cfg := &Config{}
	cfg.UpsertContext(Context{Alias: "Beta", APIKey: "sk_beta"})
	cfg.UpsertContext(Context{Alias: "alpha", APIKey: "sk_alpha"})
	cfg.UpsertContext(Context{Alias: "beta", APIKey: "sk_beta_new", ProjectID: "proj_beta"})
	cfg.ActiveContext = "beta"

	if len(cfg.Contexts) != 2 {
		t.Fatalf("len(Contexts) = %d, want 2", len(cfg.Contexts))
	}
	if cfg.Contexts[0].Alias != "alpha" || cfg.Contexts[1].APIKey != "sk_beta_new" {
		t.Fatalf("unexpected contexts ordering/update: %#v", cfg.Contexts)
	}

	if removed := cfg.RemoveContext("BETA"); !removed {
		t.Fatal("expected context removal")
	}
	if cfg.ActiveContext != "" {
		t.Fatalf("ActiveContext = %q, want empty", cfg.ActiveContext)
	}
}

func TestStoreSaveStripsStoredAPIKeys(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "config.json")
	store, err := NewStore(path)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	cfg := &Config{
		ActiveContext: "prod",
		Contexts: []Context{{
			Alias:       "prod",
			APIKey:      "sk_test_1234",
			APIKeyStore: "os_credential_store",
			ProjectID:   "proj_123",
		}},
	}
	if err := store.Save(cfg); err != nil {
		t.Fatalf("Save: %v", err)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if strings.Contains(string(raw), "sk_test_1234") {
		t.Fatalf("stored config leaked API key: %s", raw)
	}

	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.Contexts[0].APIKey != "" {
		t.Fatalf("loaded APIKey = %q, want empty", loaded.Contexts[0].APIKey)
	}
	if loaded.Contexts[0].APIKeyStore != "os_credential_store" {
		t.Fatalf("APIKeyStore = %q", loaded.Contexts[0].APIKeyStore)
	}
}
