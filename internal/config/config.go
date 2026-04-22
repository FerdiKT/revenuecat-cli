package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

const (
	DefaultAPIBaseURL = "https://api.revenuecat.com/v2"
	configDirName     = "revenuecat"
	configFileName    = "config.json"
)

type Context struct {
	Alias          string         `json:"alias"`
	APIKey         string         `json:"api_key"`
	ProjectID      string         `json:"project_id,omitempty"`
	ProjectName    string         `json:"project_name,omitempty"`
	APIBaseURL     string         `json:"api_base_url,omitempty"`
	CachedMetadata map[string]any `json:"cached_metadata,omitempty"`
}

type Config struct {
	ActiveContext string    `json:"active_context,omitempty"`
	OutputFormat  string    `json:"output_format,omitempty"`
	Contexts      []Context `json:"contexts,omitempty"`
	OAuth         OAuth     `json:"oauth,omitempty"`
}

type OAuth struct {
	ClientID    string   `json:"client_id,omitempty"`
	RedirectURI string   `json:"redirect_uri,omitempty"`
	Scopes      []string `json:"scopes,omitempty"`
	TokenStore  string   `json:"token_store,omitempty"`
}

type Store struct {
	path string
}

func NewStore(path string) (*Store, error) {
	if path == "" {
		defaultPath, err := DefaultPath()
		if err != nil {
			return nil, err
		}
		path = defaultPath
	}

	return &Store{path: path}, nil
}

func DefaultPath() (string, error) {
	if configHome := os.Getenv("XDG_CONFIG_HOME"); configHome != "" {
		return filepath.Join(configHome, configDirName, configFileName), nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}

	return filepath.Join(homeDir, ".config", configDirName, configFileName), nil
}

func (s *Store) Path() string {
	return s.path
}

func (s *Store) Load() (*Config, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &Config{}, nil
		}
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("decode config: %w", err)
	}

	for i := range cfg.Contexts {
		cfg.Contexts[i].normalize()
	}

	return &cfg, nil
}

func (s *Store) Save(cfg *Config) error {
	if err := cfg.Validate(); err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("encode config: %w", err)
	}
	data = append(data, '\n')

	if err := os.WriteFile(s.path, data, 0o600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	return nil
}

func (c *Config) Validate() error {
	seen := map[string]struct{}{}
	for i := range c.Contexts {
		if err := c.Contexts[i].Validate(); err != nil {
			return err
		}
		key := strings.ToLower(c.Contexts[i].Alias)
		if _, ok := seen[key]; ok {
			return fmt.Errorf("duplicate context alias: %s", c.Contexts[i].Alias)
		}
		seen[key] = struct{}{}
	}

	if c.ActiveContext != "" {
		if _, ok := c.FindContext(c.ActiveContext); !ok {
			return fmt.Errorf("active context %q does not exist", c.ActiveContext)
		}
	}

	return nil
}

func (c *Config) FindContext(alias string) (*Context, bool) {
	for i := range c.Contexts {
		if strings.EqualFold(c.Contexts[i].Alias, alias) {
			return &c.Contexts[i], true
		}
	}
	return nil, false
}

func (c *Config) UpsertContext(ctx Context) {
	ctx.normalize()
	for i := range c.Contexts {
		if strings.EqualFold(c.Contexts[i].Alias, ctx.Alias) {
			c.Contexts[i] = ctx
			return
		}
	}

	c.Contexts = append(c.Contexts, ctx)
	slices.SortFunc(c.Contexts, func(a, b Context) int {
		return strings.Compare(strings.ToLower(a.Alias), strings.ToLower(b.Alias))
	})
}

func (c *Config) RemoveContext(alias string) bool {
	for i := range c.Contexts {
		if strings.EqualFold(c.Contexts[i].Alias, alias) {
			c.Contexts = append(c.Contexts[:i], c.Contexts[i+1:]...)
			if strings.EqualFold(c.ActiveContext, alias) {
				c.ActiveContext = ""
			}
			return true
		}
	}
	return false
}

func (c *Config) normalize() {
	for i := range c.Contexts {
		c.Contexts[i].normalize()
	}
}

func (c *Context) normalize() {
	c.Alias = strings.TrimSpace(c.Alias)
	c.APIKey = strings.TrimSpace(c.APIKey)
	c.ProjectID = strings.TrimSpace(c.ProjectID)
	c.ProjectName = strings.TrimSpace(c.ProjectName)
	c.APIBaseURL = strings.TrimSpace(c.APIBaseURL)
	if c.APIBaseURL == "" {
		c.APIBaseURL = DefaultAPIBaseURL
	}
}

func (c Context) Validate() error {
	if c.Alias == "" {
		return errors.New("context alias is required")
	}
	if c.APIKey == "" {
		return fmt.Errorf("context %q is missing api_key", c.Alias)
	}
	return nil
}

func MaskAPIKey(value string) string {
	if value == "" {
		return ""
	}
	if len(value) <= 8 {
		return "****"
	}
	return value[:4] + strings.Repeat("*", len(value)-8) + value[len(value)-4:]
}
