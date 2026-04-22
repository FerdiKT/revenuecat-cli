---
name: revenuecat-cli
description: Use this skill when working with the local `revenuecat` CLI for RevenueCat v2 project operations, multi-context API key workflows, project snapshots, metrics pulls, or agent-safe mutations across apps, entitlements, products, offerings, packages, paywalls, customers, subscriptions, and purchases.
---

# RevenueCat CLI

Use this skill for repository-local RevenueCat CLI work.

## Workflow

1. Resolve context first.
2. Pull current state before planning mutations.
3. Prefer JSON output for agent workflows.
4. Use precise resource commands for changes.

## Context Resolution

- Inspect contexts with `revenuecat contexts list` or `revenuecat contexts show`.
- Select a default with `revenuecat contexts use <alias>`.
- Override per call with `--context <alias>`.
- After OAuth login, use `--project-id <project_id>` for project-scoped commands without changing API-key contexts.
- Use `--all-contexts` only for read commands.

If a context does not have `project_id`, run `revenuecat contexts verify <alias>`. If discovery fails, update the context manually with the correct project id.

## Read Pattern

- Start with `revenuecat pull project` for a single project snapshot.
- After OAuth login, use `revenuecat projects list` for account-level project discovery.
- After OAuth login, use `revenuecat projects create --name "..."` for account-level project creation.
- For OAuth project-scoped reads, pass `--project-id <project_id>` to the resource command.
- Use `revenuecat pull all` to compare every configured project.
- Use `revenuecat <resource> list` or `get` for narrower reads.
- Use `revenuecat apps resolve --bundle-id ...` when you need an app id for metrics filters.
- Use `revenuecat apps public-keys <app_id>` to inspect app public SDK keys.
- Use `revenuecat apps storekit-config <app_id>` to inspect iOS StoreKit configuration.
- Use `revenuecat paywalls list|get|create|delete` for paywall configuration.
- Use `revenuecat metrics countries <name>` for country breakdown tables.
- Use `revenuecat metrics chart <name>` for raw chart payloads, and prefer `--filters-json` / `--selectors-json` for complex queries.
- Use `revenuecat metrics overview` for quick KPI reads.

## Mutation Pattern

- Use `create`, `update`, `archive`, `unarchive`, `attach-products`, and `detach-products` with `--data` or `--file`.
- Keep mutations single-context.
- Prefer reading the latest snapshot immediately before changes.
- Destructive deletes require exact confirmation, e.g. `revenuecat apps delete app_123 --confirm app_123` or `revenuecat paywalls delete paywall_123 --confirm paywall_123`.
- Never print raw API keys in normal output or docs.

## Auth Guardrail

API key contexts remain the stable path for project-scoped commands. OAuth-backed `projects list|get` is available for account-level project discovery, and `--project-id` lets project-scoped commands use OAuth. API keys and OAuth tokens are stored in the OS credential store. Never print raw API keys, OAuth access tokens, or OAuth refresh tokens.
