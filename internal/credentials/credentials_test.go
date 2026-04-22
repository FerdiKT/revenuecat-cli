package credentials

import (
	"testing"

	"github.com/zalando/go-keyring"
)

func TestOAuthTokenRoundTrip(t *testing.T) {
	keyring.MockInit()

	_, found, err := LoadOAuth()
	if err != nil {
		t.Fatalf("LoadOAuth before save: %v", err)
	}
	if found {
		t.Fatal("expected no token before save")
	}

	want := OAuthToken{
		AccessToken:  "atk_test",
		RefreshToken: "rtk_test",
		TokenType:    "Bearer",
		ExpiresAt:    "2026-04-22T21:31:33Z",
		Scopes:       []string{"project_configuration:projects:read"},
	}
	if err := SaveOAuth(want); err != nil {
		t.Fatalf("SaveOAuth: %v", err)
	}

	got, found, err := LoadOAuth()
	if err != nil {
		t.Fatalf("LoadOAuth after save: %v", err)
	}
	if !found {
		t.Fatal("expected token after save")
	}
	if got.AccessToken != want.AccessToken || got.RefreshToken != want.RefreshToken || got.Scopes[0] != want.Scopes[0] {
		t.Fatalf("token = %#v, want %#v", got, want)
	}

	if err := DeleteOAuth(); err != nil {
		t.Fatalf("DeleteOAuth: %v", err)
	}
	_, found, err = LoadOAuth()
	if err != nil {
		t.Fatalf("LoadOAuth after delete: %v", err)
	}
	if found {
		t.Fatal("expected no token after delete")
	}
}

func TestAPIKeyRoundTrip(t *testing.T) {
	keyring.MockInit()

	if err := SaveAPIKey("Prod", "sk_test"); err != nil {
		t.Fatalf("SaveAPIKey: %v", err)
	}

	got, found, err := LoadAPIKey("prod")
	if err != nil {
		t.Fatalf("LoadAPIKey: %v", err)
	}
	if !found {
		t.Fatal("expected API key after save")
	}
	if got != "sk_test" {
		t.Fatalf("api key = %q, want sk_test", got)
	}

	if err := DeleteAPIKey("PROD"); err != nil {
		t.Fatalf("DeleteAPIKey: %v", err)
	}
	_, found, err = LoadAPIKey("prod")
	if err != nil {
		t.Fatalf("LoadAPIKey after delete: %v", err)
	}
	if found {
		t.Fatal("expected no API key after delete")
	}
}
