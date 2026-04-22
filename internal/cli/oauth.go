package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/FerdiKT/revenuecat-cli/internal/config"
	"github.com/FerdiKT/revenuecat-cli/internal/credentials"
	"github.com/FerdiKT/revenuecat-cli/internal/exitcode"
	"github.com/FerdiKT/revenuecat-cli/internal/rcapi"
)

const oauthRefreshSkew = 2 * time.Minute

func (a *App) oauthClient(ctx context.Context, cfg *config.Config) (*rcapi.Client, error) {
	token, err := a.freshOAuthToken(ctx, cfg)
	if err != nil {
		return nil, err
	}
	apiBaseURL := chooseNonEmpty(cfg.OAuth.APIBaseURL, config.DefaultAPIBaseURL)
	return rcapi.NewClient(token.AccessToken, apiBaseURL), nil
}

func (a *App) freshOAuthToken(ctx context.Context, cfg *config.Config) (credentials.OAuthToken, error) {
	token, found, err := credentials.LoadOAuth()
	if err != nil {
		return credentials.OAuthToken{}, &CLIError{Code: exitcode.Internal, Message: err.Error()}
	}
	if !found || strings.TrimSpace(token.AccessToken) == "" {
		return credentials.OAuthToken{}, &CLIError{Code: exitcode.Auth, Message: "OAuth token not found; run `revenuecat auth login`"}
	}
	if !oauthTokenNeedsRefresh(token) {
		return token, nil
	}
	if strings.TrimSpace(token.RefreshToken) == "" {
		return credentials.OAuthToken{}, &CLIError{Code: exitcode.Auth, Message: "OAuth access token expired and no refresh token is available; run `revenuecat auth login`"}
	}

	response, err := exchangeOAuthRefresh(ctx, oauthRefreshOptions{
		ClientID:     chooseNonEmpty(cfg.OAuth.ClientID, defaultOAuthClientID),
		ClientSecret: os.Getenv(oauthSecretEnv),
		OAuthBaseURL: strings.TrimRight(chooseNonEmpty(cfg.OAuth.OAuthBaseURL, defaultOAuthBaseURL), "/"),
		RefreshToken: token.RefreshToken,
	})
	if err != nil {
		return credentials.OAuthToken{}, err
	}

	refreshed := credentials.OAuthToken{
		AccessToken:  response.AccessToken,
		RefreshToken: chooseNonEmpty(response.RefreshToken, token.RefreshToken),
		TokenType:    chooseNonEmpty(response.TokenType, token.TokenType, "Bearer"),
		ExpiresAt:    token.ExpiresAt,
		Scopes:       token.Scopes,
	}
	if response.ExpiresIn > 0 {
		refreshed.ExpiresAt = time.Now().Add(time.Duration(response.ExpiresIn) * time.Second).UTC().Format(time.RFC3339)
	}
	if scopes := strings.Fields(response.Scope); len(scopes) > 0 {
		refreshed.Scopes = scopes
	}
	if err := credentials.SaveOAuth(refreshed); err != nil {
		return credentials.OAuthToken{}, &CLIError{Code: exitcode.Internal, Message: err.Error()}
	}
	return refreshed, nil
}

func oauthTokenNeedsRefresh(token credentials.OAuthToken) bool {
	if strings.TrimSpace(token.ExpiresAt) == "" {
		return false
	}
	expiresAt, err := time.Parse(time.RFC3339, token.ExpiresAt)
	if err != nil {
		return true
	}
	return time.Now().Add(oauthRefreshSkew).After(expiresAt)
}

type oauthRefreshOptions struct {
	ClientID     string
	ClientSecret string
	OAuthBaseURL string
	RefreshToken string
}

func exchangeOAuthRefresh(ctx context.Context, opts oauthRefreshOptions) (*oauthTokenResponse, error) {
	if strings.TrimSpace(opts.ClientID) == "" {
		return nil, &CLIError{Code: exitcode.Auth, Message: "OAuth client ID is missing; run `revenuecat auth login`"}
	}
	if strings.TrimSpace(opts.OAuthBaseURL) == "" {
		return nil, &CLIError{Code: exitcode.Auth, Message: "OAuth base URL is missing; run `revenuecat auth login`"}
	}

	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", opts.RefreshToken)
	form.Set("client_id", opts.ClientID)
	if strings.TrimSpace(opts.ClientSecret) != "" {
		form.Set("client_secret", opts.ClientSecret)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(opts.OAuthBaseURL, "/")+"/oauth2/token", strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		var payload map[string]any
		_ = json.Unmarshal(data, &payload)
		message := fmt.Sprintf("OAuth token refresh failed with status %d", resp.StatusCode)
		if description, ok := payload["error_description"].(string); ok && description != "" {
			message = description
		} else if oauthErr, ok := payload["error"].(string); ok && oauthErr != "" {
			message = oauthErr
		}
		return nil, &CLIError{Code: exitcode.Auth, Message: message}
	}

	var token oauthTokenResponse
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, err
	}
	if token.AccessToken == "" {
		return nil, &CLIError{Code: exitcode.Auth, Message: "OAuth refresh response did not include access_token"}
	}
	if token.TokenType == "" {
		token.TokenType = "Bearer"
	}
	return &token, nil
}
