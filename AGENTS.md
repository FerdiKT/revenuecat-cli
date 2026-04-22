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
2. Use `revenuecat contexts use <alias>` or `--context <alias>` to lock the target project.
3. If a context is missing `project_id`, run `revenuecat contexts verify <alias>` or update the context manually.

## Read Before Write

- For project discovery inside a known context: `revenuecat pull project`
- For estate-wide comparison: `revenuecat pull all`
- For focused reads: `revenuecat <resource> list|get`
- Resolve app ids with `revenuecat apps resolve --bundle-id ...` before app-scoped metrics queries.
- For country tables: `revenuecat metrics countries <chart_name> --app <app_id> ...`
- For raw chart payloads: `revenuecat metrics chart <chart_name> --filters-json ... --selectors-json ...`
- For quick KPI reads: `revenuecat metrics overview|options`

## Mutation Rules

- Mutations always target exactly one context.
- Use `--data '<json>'` or `--file payload.json` for create, update, archive, attach, and detach flows.
- Do not use `--all-contexts` with mutating commands.

## Auth

- API key contexts remain the stable path for project-scoped commands.
- OAuth login is available for account-level workflows and stores tokens in the OS credential store.
- Do not read, print, or copy API keys, OAuth access tokens, or OAuth refresh tokens from the OS credential store.
