package credentials

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/zalando/go-keyring"
)

const (
	serviceName      = "revenuecat-cli"
	oauthAccountName = "oauth-token"
)

type OAuthToken struct {
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token,omitempty"`
	TokenType    string   `json:"token_type,omitempty"`
	ExpiresAt    string   `json:"expires_at,omitempty"`
	Scopes       []string `json:"scopes,omitempty"`
}

func SaveOAuth(token OAuthToken) error {
	data, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("encode OAuth token: %w", err)
	}
	if err := keyring.Set(serviceName, oauthAccountName, string(data)); err != nil {
		return fmt.Errorf("save OAuth token to OS credential store: %w", err)
	}
	return nil
}

func LoadOAuth() (OAuthToken, bool, error) {
	data, err := keyring.Get(serviceName, oauthAccountName)
	if err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return OAuthToken{}, false, nil
		}
		return OAuthToken{}, false, fmt.Errorf("load OAuth token from OS credential store: %w", err)
	}
	var token OAuthToken
	if err := json.Unmarshal([]byte(data), &token); err != nil {
		return OAuthToken{}, false, fmt.Errorf("decode OAuth token from OS credential store: %w", err)
	}
	return token, true, nil
}

func DeleteOAuth() error {
	if err := keyring.Delete(serviceName, oauthAccountName); err != nil && !errors.Is(err, keyring.ErrNotFound) {
		return fmt.Errorf("delete OAuth token from OS credential store: %w", err)
	}
	return nil
}
