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
