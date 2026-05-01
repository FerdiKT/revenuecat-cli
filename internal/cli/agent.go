package cli

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/FerdiKT/revenuecat-cli/internal/exitcode"
	"github.com/spf13/cobra"
)

const (
	bundledSkillName = "revenuecat-cli"
	bundledSkillDoc  = `---
name: revenuecat-cli
description: Use this skill when working with the local ` + "`revenuecat`" + ` CLI for RevenueCat v2 OAuth-first project operations, optional multi-context API key workflows, project snapshots, metrics pulls, or agent-safe mutations across apps, entitlements, products, offerings, packages, paywalls, customers, subscriptions, and purchases.
---

# RevenueCat CLI

Use this skill for repository-local RevenueCat CLI work.

## Workflow

1. Resolve the target project first.
2. Pull current state before planning mutations.
3. Prefer JSON output for agent workflows.
4. Use precise resource commands for changes.

## Context Resolution

- After OAuth login, use ` + "`--project-id <project_id>`" + ` for project-scoped commands.
- Inspect contexts with ` + "`revenuecat contexts list`" + ` or ` + "`revenuecat contexts show`" + ` when you need fixed API-key aliases.
- Select a default with ` + "`revenuecat contexts use <alias>`" + `.
- Override a context per call with ` + "`--context <alias>`" + `.
- Use ` + "`--all-contexts`" + ` only for read commands.

If an API-key context does not have ` + "`project_id`" + `, run ` + "`revenuecat contexts verify <alias>`" + `. If discovery fails, update the context manually with the correct project id.

## Read Pattern

- After OAuth login, use ` + "`revenuecat projects list`" + ` for account-level project discovery.
- After OAuth login, use ` + "`revenuecat projects create --name \"...\"`" + ` for account-level project creation.
- Start with ` + "`revenuecat pull project --project-id <project_id>`" + ` for a single project snapshot.
- For OAuth project-scoped reads, pass ` + "`--project-id <project_id>`" + ` to the resource command.
- Use ` + "`revenuecat pull all`" + ` to compare every configured project.
- Use ` + "`revenuecat <resource> list`" + ` or ` + "`get`" + ` for narrower reads.
- Use ` + "`revenuecat apps public-keys <app_id>`" + ` to inspect app public SDK keys.
- Use ` + "`revenuecat apps storekit-config <app_id>`" + ` to inspect iOS StoreKit configuration.
- Use ` + "`revenuecat paywalls list|get|create|delete`" + ` for paywall configuration.
- Use ` + "`revenuecat metrics overview`" + ` or ` + "`revenuecat metrics chart <name>`" + ` for KPI and chart data.

## Mutation Pattern

- Use ` + "`create`" + `, ` + "`update`" + `, ` + "`archive`" + `, ` + "`unarchive`" + `, ` + "`attach-products`" + `, and ` + "`detach-products`" + ` with ` + "`--data`" + ` or ` + "`--file`" + `.
- Keep mutations single-project.
- Prefer reading the latest snapshot immediately before changes.
- Destructive deletes require exact confirmation, e.g. ` + "`revenuecat apps delete app_123 --confirm app_123`" + ` or ` + "`revenuecat paywalls delete paywall_123 --confirm paywall_123`" + `.
- Never print raw API keys or OAuth tokens in normal output or docs.

## Auth Guardrail

OAuth is the preferred path for account-level and project-scoped commands. API-key contexts remain useful for named aliases and ` + "`pull all`" + `. API keys and OAuth tokens are stored in the OS credential store.
`
	bundledAgentYAML = `display_name: RevenueCat CLI
short_description: Work with the local RevenueCat agent-first CLI using context-first, pull-first workflows.
default_prompt: Use the local revenuecat CLI. Resolve the target project first, prefer JSON output, pull current state before planning mutations, and treat API keys and OAuth tokens as secrets.
`
)

func addAgentCommands(root *cobra.Command) {
	cmd := &cobra.Command{
		Use:   "agent",
		Short: "Install or link agent assets such as the bundled Codex skill",
	}

	cmd.AddCommand(
		newAgentInstallSkillCommand(),
		newAgentLinkSkillCommand(),
		newAgentShowSkillPathCommand(),
	)

	root.AddCommand(cmd)
}

func newAgentInstallSkillCommand() *cobra.Command {
	var codexHome string
	var force bool

	cmd := &cobra.Command{
		Use:   "install-skill",
		Short: "Install the bundled revenuecat Codex skill into CODEX_HOME",
		RunE: func(cmd *cobra.Command, args []string) error {
			targetDir, err := resolveSkillInstallDir(codexHome)
			if err != nil {
				return err
			}
			if err := ensureReplaceableTarget(targetDir, force); err != nil {
				return err
			}
			if err := writeBundledSkill(targetDir); err != nil {
				return &CLIError{Code: exitcode.Internal, Message: err.Error()}
			}
			_, err = fmt.Fprintf(os.Stdout, "skill installed: %s\n", targetDir)
			return err
		},
	}

	cmd.Flags().StringVar(&codexHome, "codex-home", "", "Override CODEX_HOME for the destination")
	cmd.Flags().BoolVar(&force, "force", false, "Replace an existing installed skill")
	return cmd
}

func newAgentLinkSkillCommand() *cobra.Command {
	var codexHome string
	var source string
	var force bool

	cmd := &cobra.Command{
		Use:   "link-skill",
		Short: "Symlink the local revenuecat skill into CODEX_HOME",
		RunE: func(cmd *cobra.Command, args []string) error {
			targetDir, err := resolveSkillInstallDir(codexHome)
			if err != nil {
				return err
			}
			if source == "" {
				source = filepath.Join(".", "skills", bundledSkillName)
			}
			source, err = filepath.Abs(source)
			if err != nil {
				return &CLIError{Code: exitcode.Usage, Message: fmt.Sprintf("resolve source path: %v", err)}
			}
			if err := validateSkillSource(source); err != nil {
				return err
			}
			if err := ensureReplaceableTarget(targetDir, force); err != nil {
				return err
			}
			if err := os.MkdirAll(filepath.Dir(targetDir), 0o755); err != nil {
				return &CLIError{Code: exitcode.Internal, Message: fmt.Sprintf("create destination parent: %v", err)}
			}
			if err := os.Symlink(source, targetDir); err != nil {
				return &CLIError{Code: exitcode.Internal, Message: fmt.Sprintf("create symlink: %v", err)}
			}
			_, err = fmt.Fprintf(os.Stdout, "skill linked: %s -> %s\n", targetDir, source)
			return err
		},
	}

	cmd.Flags().StringVar(&codexHome, "codex-home", "", "Override CODEX_HOME for the destination")
	cmd.Flags().StringVar(&source, "source", "", "Source skill directory to symlink; defaults to ./skills/revenuecat-cli")
	cmd.Flags().BoolVar(&force, "force", false, "Replace an existing installed skill or symlink")
	return cmd
}

func newAgentShowSkillPathCommand() *cobra.Command {
	var codexHome string

	cmd := &cobra.Command{
		Use:   "show-skill-path",
		Short: "Print the destination path for the revenuecat Codex skill",
		RunE: func(cmd *cobra.Command, args []string) error {
			targetDir, err := resolveSkillInstallDir(codexHome)
			if err != nil {
				return err
			}
			_, err = fmt.Fprintln(os.Stdout, targetDir)
			return err
		},
	}

	cmd.Flags().StringVar(&codexHome, "codex-home", "", "Override CODEX_HOME for the destination")
	return cmd
}

func resolveSkillInstallDir(codexHome string) (string, error) {
	home := codexHome
	if home == "" {
		home = os.Getenv("CODEX_HOME")
	}
	if home == "" {
		userHome, err := os.UserHomeDir()
		if err != nil {
			return "", &CLIError{Code: exitcode.Internal, Message: fmt.Sprintf("resolve user home: %v", err)}
		}
		home = filepath.Join(userHome, ".codex")
	}
	return filepath.Join(home, "skills", bundledSkillName), nil
}

func ensureReplaceableTarget(target string, force bool) error {
	info, err := os.Lstat(target)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return &CLIError{Code: exitcode.Internal, Message: fmt.Sprintf("inspect target: %v", err)}
	}
	if !force {
		return &CLIError{Code: exitcode.Conflict, Message: fmt.Sprintf("target already exists: %s (use --force to replace)", target)}
	}
	if info.Mode()&os.ModeSymlink != 0 || info.IsDir() {
		if err := os.RemoveAll(target); err != nil {
			return &CLIError{Code: exitcode.Internal, Message: fmt.Sprintf("remove existing target: %v", err)}
		}
		return nil
	}
	if err := os.Remove(target); err != nil {
		return &CLIError{Code: exitcode.Internal, Message: fmt.Sprintf("remove existing target: %v", err)}
	}
	return nil
}

func writeBundledSkill(target string) error {
	if err := os.MkdirAll(filepath.Join(target, "agents"), 0o755); err != nil {
		return fmt.Errorf("create skill directory: %w", err)
	}
	if err := os.WriteFile(filepath.Join(target, "SKILL.md"), []byte(bundledSkillDoc), 0o644); err != nil {
		return fmt.Errorf("write SKILL.md: %w", err)
	}
	if err := os.WriteFile(filepath.Join(target, "agents", "openai.yaml"), []byte(bundledAgentYAML), 0o644); err != nil {
		return fmt.Errorf("write agents/openai.yaml: %w", err)
	}
	return nil
}

func validateSkillSource(source string) error {
	info, err := os.Stat(source)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return &CLIError{Code: exitcode.Usage, Message: fmt.Sprintf("skill source does not exist: %s", source)}
		}
		return &CLIError{Code: exitcode.Internal, Message: fmt.Sprintf("inspect source: %v", err)}
	}
	if !info.IsDir() {
		return &CLIError{Code: exitcode.Usage, Message: fmt.Sprintf("skill source is not a directory: %s", source)}
	}
	if _, err := os.Stat(filepath.Join(source, "SKILL.md")); err != nil {
		return &CLIError{Code: exitcode.Usage, Message: fmt.Sprintf("skill source is missing SKILL.md: %s", source)}
	}
	return nil
}
