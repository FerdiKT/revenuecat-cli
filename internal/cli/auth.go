package cli

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/FerdiKT/revenuecat-cli/internal/config"
	"github.com/FerdiKT/revenuecat-cli/internal/credentials"
	"github.com/FerdiKT/revenuecat-cli/internal/exitcode"
	"github.com/FerdiKT/revenuecat-cli/internal/output"
	"github.com/spf13/cobra"
)

const (
	defaultOAuthClientID    = "UmV2ZW51ZUNhdCBDTEkgKEZlcmRpIEvEsXrEsWx0b3ByYWsp"
	defaultOAuthRedirectURI = "http://127.0.0.1:8787/oauth/callback"
	defaultOAuthBaseURL     = "https://api.revenuecat.com"
	oauthSecretEnv          = "REVENUECAT_OAUTH_CLIENT_SECRET"
)

var defaultOAuthScopes = []string{
	"project_configuration:projects:read",
	"project_configuration:projects:read_write",
	"project_configuration:apps:read",
	"project_configuration:apps:read_write",
	"project_configuration:entitlements:read",
	"project_configuration:entitlements:read_write",
	"project_configuration:offerings:read",
	"project_configuration:offerings:read_write",
	"project_configuration:packages:read",
	"project_configuration:packages:read_write",
	"project_configuration:products:read",
	"project_configuration:products:read_write",
	"charts_metrics:overview:read",
	"charts_metrics:charts:read",
}

func addAuthCommands(root *cobra.Command, app *App) {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Authentication status and OAuth login",
	}

	cmd.AddCommand(
		newAuthStatusCommand(app),
		newAuthLoginCommand(app),
		newAuthLogoutCommand(app),
	)

	root.AddCommand(cmd)
}

func newAuthStatusCommand(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show current authentication mode",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := app.loadConfig()
			if err != nil {
				return err
			}

			token, found, tokenErr := credentials.LoadOAuth()
			tokenStatus := "ok"
			if tokenErr != nil {
				tokenStatus = tokenErr.Error()
			}
			oauthConfigured := found && strings.TrimSpace(token.AccessToken) != ""
			mode := "api_key"
			if oauthConfigured {
				mode = "oauth"
			}

			return output.PrintJSON(os.Stdout, output.Success(nil, map[string]any{
				"mode":               mode,
				"active_context":     cfg.ActiveContext,
				"context_count":      len(cfg.Contexts),
				"oauth_available":    true,
				"oauth_configured":   oauthConfigured,
				"oauth_client_id":    maskClientID(cfg.OAuth.ClientID),
				"oauth_expires_at":   token.ExpiresAt,
				"oauth_scopes":       token.Scopes,
				"oauth_token_store":  chooseNonEmpty(cfg.OAuth.TokenStore, credentials.StoreName),
				"oauth_token_status": tokenStatus,
			}, output.Meta{}))
		},
	}
}

func newAuthLoginCommand(app *App) *cobra.Command {
	var clientID string
	var redirectURI string
	var scopeFlags []string
	var oauthBaseURL string
	var noOpen bool
	var timeout time.Duration

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Login with RevenueCat OAuth",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := app.loadConfig()
			if err != nil {
				return err
			}

			scopes := scopeFlags
			if len(scopes) == 0 {
				scopes = defaultOAuthScopes
			}
			token, err := runOAuthLogin(context.Background(), oauthLoginOptions{
				ClientID:     clientID,
				ClientSecret: os.Getenv(oauthSecretEnv),
				RedirectURI:  redirectURI,
				Scopes:       scopes,
				OAuthBaseURL: strings.TrimRight(oauthBaseURL, "/"),
				NoOpen:       noOpen,
				Timeout:      timeout,
				Stdout:       os.Stdout,
			})
			if err != nil {
				if cliErr, ok := err.(*CLIError); ok {
					return cliErr
				}
				return &CLIError{Code: exitcode.Internal, Message: err.Error()}
			}

			expiresAt := time.Now().Add(time.Duration(token.ExpiresIn) * time.Second).UTC().Format(time.RFC3339)
			if err := credentials.SaveOAuth(credentials.OAuthToken{
				AccessToken:  token.AccessToken,
				RefreshToken: token.RefreshToken,
				TokenType:    token.TokenType,
				ExpiresAt:    expiresAt,
				Scopes:       scopes,
			}); err != nil {
				return &CLIError{Code: exitcode.Internal, Message: err.Error()}
			}

			cfg.OAuth = config.OAuth{
				ClientID:    clientID,
				RedirectURI: redirectURI,
				Scopes:      scopes,
				TokenStore:  credentials.StoreName,
			}
			if err := app.saveConfig(cfg); err != nil {
				return err
			}

			return output.PrintJSON(os.Stdout, output.Success(nil, map[string]any{
				"mode":        "oauth",
				"expires_at":  expiresAt,
				"scopes":      scopes,
				"token_store": credentials.StoreName,
			}, output.Meta{}))
		},
	}

	cmd.Flags().StringVar(&clientID, "client-id", defaultOAuthClientID, "OAuth client ID")
	cmd.Flags().StringVar(&redirectURI, "redirect-uri", defaultOAuthRedirectURI, "OAuth redirect URI registered with RevenueCat")
	cmd.Flags().StringSliceVar(&scopeFlags, "scope", nil, "OAuth scope to request; repeat or comma-separate")
	cmd.Flags().StringVar(&oauthBaseURL, "oauth-base-url", defaultOAuthBaseURL, "RevenueCat OAuth base URL")
	cmd.Flags().BoolVar(&noOpen, "no-open", false, "Print the authorization URL instead of opening a browser")
	cmd.Flags().DurationVar(&timeout, "timeout", 5*time.Minute, "How long to wait for browser authorization")
	return cmd
}

func newAuthLogoutCommand(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Remove locally stored OAuth tokens",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := app.loadConfig()
			if err != nil {
				return err
			}
			cfg.OAuth = config.OAuth{}
			if err := credentials.DeleteOAuth(); err != nil {
				return &CLIError{Code: exitcode.Internal, Message: err.Error()}
			}
			if err := app.saveConfig(cfg); err != nil {
				return err
			}
			_, err = fmt.Fprintln(os.Stdout, "OAuth tokens removed from OS credential store")
			return err
		},
	}
}

type oauthLoginOptions struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
	Scopes       []string
	OAuthBaseURL string
	NoOpen       bool
	Timeout      time.Duration
	Stdout       io.Writer
}

type oauthTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
}

func runOAuthLogin(ctx context.Context, opts oauthLoginOptions) (*oauthTokenResponse, error) {
	if strings.TrimSpace(opts.ClientID) == "" {
		return nil, &CLIError{Code: exitcode.Usage, Message: "--client-id is required"}
	}
	if strings.TrimSpace(opts.RedirectURI) == "" {
		return nil, &CLIError{Code: exitcode.Usage, Message: "--redirect-uri is required"}
	}
	if strings.TrimSpace(opts.OAuthBaseURL) == "" {
		return nil, &CLIError{Code: exitcode.Usage, Message: "--oauth-base-url is required"}
	}
	if len(opts.Scopes) == 0 {
		return nil, &CLIError{Code: exitcode.Usage, Message: "at least one OAuth scope is required"}
	}
	if opts.Timeout <= 0 {
		opts.Timeout = 5 * time.Minute
	}
	if opts.Stdout == nil {
		opts.Stdout = io.Discard
	}

	redirectURL, err := url.Parse(opts.RedirectURI)
	if err != nil {
		return nil, &CLIError{Code: exitcode.Usage, Message: fmt.Sprintf("invalid --redirect-uri: %v", err)}
	}
	if redirectURL.Scheme != "http" || redirectURL.Hostname() != "127.0.0.1" {
		return nil, &CLIError{Code: exitcode.Usage, Message: "--redirect-uri must be a localhost loopback URL, e.g. http://127.0.0.1:8787/oauth/callback"}
	}

	verifier, challenge, err := generatePKCEPair()
	if err != nil {
		return nil, err
	}
	state, err := randomURLString(24)
	if err != nil {
		return nil, err
	}

	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)
	server, err := startOAuthCallbackServer(redirectURL, state, codeCh, errCh)
	if err != nil {
		return nil, err
	}
	defer server.Close()

	authorizeURL, err := buildAuthorizeURL(opts, state, challenge)
	if err != nil {
		return nil, err
	}
	fmt.Fprintf(opts.Stdout, "Open this URL to authorize RevenueCat CLI:\n%s\n", authorizeURL)
	if !opts.NoOpen {
		if err := openBrowser(authorizeURL); err != nil {
			fmt.Fprintf(opts.Stdout, "Could not open browser automatically: %v\n", err)
		}
	}

	waitCtx, cancel := context.WithTimeout(ctx, opts.Timeout)
	defer cancel()

	select {
	case code := <-codeCh:
		return exchangeOAuthCode(waitCtx, opts, code, verifier)
	case err := <-errCh:
		return nil, err
	case <-waitCtx.Done():
		return nil, &CLIError{Code: exitcode.Usage, Message: "timed out waiting for OAuth callback"}
	}
}

func startOAuthCallbackServer(redirectURL *url.URL, state string, codeCh chan<- string, errCh chan<- error) (*http.Server, error) {
	listener, err := net.Listen("tcp", redirectURL.Host)
	if err != nil {
		return nil, &CLIError{Code: exitcode.Usage, Message: fmt.Sprintf("listen on %s: %v", redirectURL.Host, err)}
	}

	mux := http.NewServeMux()
	mux.HandleFunc(redirectURL.Path, func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		if got := query.Get("state"); got != state {
			http.Error(w, "invalid state", http.StatusBadRequest)
			errCh <- &CLIError{Code: exitcode.Auth, Message: "OAuth state mismatch"}
			return
		}
		if oauthErr := query.Get("error"); oauthErr != "" {
			description := query.Get("error_description")
			if description == "" {
				description = oauthErr
			}
			http.Error(w, description, http.StatusBadRequest)
			errCh <- &CLIError{Code: exitcode.Auth, Message: description}
			return
		}
		code := query.Get("code")
		if code == "" {
			http.Error(w, "missing code", http.StatusBadRequest)
			errCh <- &CLIError{Code: exitcode.Auth, Message: "OAuth callback did not include code"}
			return
		}
		_, _ = fmt.Fprintln(w, "RevenueCat CLI authorization complete. You can close this tab.")
		codeCh <- code
	})

	server := &http.Server{Handler: mux}
	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()
	return server, nil
}

func buildAuthorizeURL(opts oauthLoginOptions, state, challenge string) (string, error) {
	baseURL, err := url.Parse(strings.TrimRight(opts.OAuthBaseURL, "/") + "/oauth2/authorize")
	if err != nil {
		return "", err
	}
	query := baseURL.Query()
	query.Set("client_id", opts.ClientID)
	query.Set("response_type", "code")
	query.Set("redirect_uri", opts.RedirectURI)
	query.Set("scope", strings.Join(opts.Scopes, " "))
	query.Set("state", state)
	query.Set("code_challenge", challenge)
	query.Set("code_challenge_method", "S256")
	baseURL.RawQuery = query.Encode()
	return baseURL.String(), nil
}

func exchangeOAuthCode(ctx context.Context, opts oauthLoginOptions, code, verifier string) (*oauthTokenResponse, error) {
	tokenURL := strings.TrimRight(opts.OAuthBaseURL, "/") + "/oauth2/token"
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	form.Set("redirect_uri", opts.RedirectURI)
	form.Set("client_id", opts.ClientID)
	form.Set("code_verifier", verifier)
	if strings.TrimSpace(opts.ClientSecret) != "" {
		form.Set("client_secret", opts.ClientSecret)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(form.Encode()))
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
		message := fmt.Sprintf("OAuth token exchange failed with status %d", resp.StatusCode)
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
		return nil, &CLIError{Code: exitcode.Auth, Message: "OAuth token response did not include access_token"}
	}
	if token.TokenType == "" {
		token.TokenType = "Bearer"
	}
	return &token, nil
}

func generatePKCEPair() (string, string, error) {
	verifier, err := randomURLString(32)
	if err != nil {
		return "", "", err
	}
	sum := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(sum[:])
	return verifier, challenge, nil
}

func randomURLString(size int) (string, error) {
	data := make([]byte, size)
	if _, err := rand.Read(data); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(data), nil
}

func openBrowser(target string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", target).Start()
	case "linux":
		return exec.Command("xdg-open", target).Start()
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", target).Start()
	default:
		return fmt.Errorf("unsupported platform %s", runtime.GOOS)
	}
}

func maskClientID(value string) string {
	if value == "" {
		return ""
	}
	if len(value) <= 12 {
		return "****"
	}
	return value[:6] + strings.Repeat("*", len(value)-12) + value[len(value)-6:]
}
