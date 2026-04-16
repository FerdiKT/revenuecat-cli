---
name: revenuecat-cli
description: Use this skill when working with the local `revenuecat` CLI for RevenueCat v2 project operations, multi-context API key workflows, project snapshots, metrics pulls, or agent-safe mutations across apps, entitlements, products, offerings, packages, customers, subscriptions, and purchases.
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
- Use `--all-contexts` only for read commands.

If a context does not have `project_id`, run `revenuecat contexts verify <alias>`. If discovery fails, update the context manually with the correct project id.

## Read Pattern

- Start with `revenuecat pull project` for a single project snapshot.
- Use `revenuecat pull all` to compare every configured project.
- Use `revenuecat <resource> list` or `get` for narrower reads.
- Use `revenuecat apps resolve --bundle-id ...` when you need an app id for metrics filters.
- Use `revenuecat metrics countries <name>` for country breakdown tables.
- Use `revenuecat metrics chart <name>` for raw chart payloads, and prefer `--filters-json` / `--selectors-json` for complex queries.
- Use `revenuecat metrics overview` for quick KPI reads.

## Mutation Pattern

- Use `create`, `update`, `archive`, `unarchive`, `attach-products`, and `detach-products` with `--data` or `--file`.
- Keep mutations single-context.
- Prefer reading the latest snapshot immediately before changes.
- Never print raw API keys in normal output or docs.

## Auth Guardrail

V1 is API-key only. `revenuecat auth login` is a coming-soon placeholder while waiting on RevenueCat OAuth support setup.
