# RevenueCat CLI Agent Guide

This repository contains an agent-first CLI for RevenueCat v2.

## Working Model

- Prefer JSON output.
- Resolve the correct context before any API call.
- Prefer `pull project` or `pull all` before planning mutations.
- Use resource commands for exact CRUD operations.
- Treat API keys and OAuth tokens as secrets. Never echo raw `sk_`, `atk_`, or `rtk_` values in normal output.
- API keys and OAuth tokens are stored in the OS credential store, not the local config file.

## Context Workflow

1. Add or inspect contexts with `revenuecat contexts ...`.
2. Use `revenuecat contexts use <alias>` or `--context <alias>` to lock the target API-key project.
3. After OAuth login, use `--project-id <project_id>` for project-scoped commands without changing contexts.
4. If a context is missing `project_id`, run `revenuecat contexts verify <alias>` or update the context manually.

## Read Before Write

- For project discovery inside a known context: `revenuecat pull project`
- For account-level project discovery after OAuth login: `revenuecat projects list`
- For account-level project creation after OAuth login: `revenuecat projects create --name "..."`
- For OAuth project-scoped reads: `revenuecat <resource> list --project-id <project_id>`
- For estate-wide comparison: `revenuecat pull all`
- For focused reads: `revenuecat <resource> list|get`
- Resolve app ids with `revenuecat apps resolve --bundle-id ...` before app-scoped metrics queries.
- Inspect app public SDK keys with `revenuecat apps public-keys <app_id>`.
- Inspect iOS StoreKit config with `revenuecat apps storekit-config <app_id>`.
- Inspect and manage paywalls with `revenuecat paywalls list|get|create|delete`.
- For country tables: `revenuecat metrics countries <chart_name> --app <app_id> ...`
- For raw chart payloads: `revenuecat metrics chart <chart_name> --filters-json ... --selectors-json ...`
- For quick KPI reads: `revenuecat metrics overview|options`

## Mutation Rules

- Mutations always target exactly one context.
- Use `--data '<json>'` or `--file payload.json` for create, update, archive, attach, detach, and paywall create flows.
- Destructive deletes require exact confirmation, e.g. `revenuecat apps delete app_123 --confirm app_123` or `revenuecat paywalls delete paywall_123 --confirm paywall_123`.
- Do not use `--all-contexts` with mutating commands.
- Do not combine `--context`, `--all-contexts`, and `--project-id`.

## Auth

- API key contexts remain the stable path for project-scoped commands.
- OAuth login is available for account-level workflows and stores tokens in the OS credential store.
- Do not read, print, or copy API keys, OAuth access tokens, or OAuth refresh tokens from the OS credential store.
