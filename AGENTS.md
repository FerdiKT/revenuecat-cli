# RevenueCat CLI Agent Guide

This repository contains an agent-first CLI for RevenueCat v2.

## Working Model

- Prefer JSON output.
- Resolve the correct context before any API call.
- Prefer `pull project` or `pull all` before planning mutations.
- Use resource commands for exact CRUD operations.
- Treat API keys as secrets. Never echo raw `sk_` values in normal output.

## Context Workflow

1. Add or inspect contexts with `revenuecat contexts ...`.
2. Use `revenuecat contexts use <alias>` or `--context <alias>` to lock the target project.
3. If a context is missing `project_id`, run `revenuecat contexts verify <alias>` or update the context manually.

## Read Before Write

- For project discovery inside a known context: `revenuecat pull project`
- For estate-wide comparison: `revenuecat pull all`
- For focused reads: `revenuecat <resource> list|get`
- For metrics: `revenuecat metrics overview|chart|options`

## Mutation Rules

- Mutations always target exactly one context.
- Use `--data '<json>'` or `--file payload.json` for create, update, archive, attach, and detach flows.
- Do not use `--all-contexts` with mutating commands.

## Auth

- V1 supports API key contexts only.
- OAuth commands are placeholders and should be treated as coming soon while waiting on RevenueCat support approval.
