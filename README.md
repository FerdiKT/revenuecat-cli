<p align="center">
  <img src="assets/hero-banner.svg" alt="revenuecat CLI" width="820" />
</p>

<h1 align="center">revenuecat</h1>

<p align="center">
  <strong>An agent-first CLI for RevenueCat</strong><br />
  Multi-project · JSON-first · CI-friendly · API-key contexts · OAuth-ready
</p>

<p align="center">
  <a href="#-status"><img src="https://img.shields.io/badge/status-beta-yellow?style=flat-square" alt="Beta" /></a>
  <a href="#-installation"><img src="https://img.shields.io/badge/homebrew-ready-8A6B3F?style=flat-square&logo=homebrew&logoColor=white" alt="Homebrew" /></a>
  <a href="#-quickstart"><img src="https://img.shields.io/badge/quickstart-5_min-brightgreen?style=flat-square" alt="Quickstart" /></a>
  <a href="#-why-revenuecat"><img src="https://img.shields.io/badge/agent-first-0EA5E9?style=flat-square" alt="Agent First" /></a>
  <a href="#-auth-model"><img src="https://img.shields.io/badge/auth-api_key_contexts-14B8A6?style=flat-square" alt="API Key Contexts" /></a>
  <a href="#-license"><img src="https://img.shields.io/badge/license-MIT-lightgrey?style=flat-square" alt="License" /></a>
</p>

---

## ⚠️ Status

> **This project is in public beta.** Core API-key context flows, project snapshots, metrics pulls, and resource CRUD are implemented for RevenueCat v2. OAuth login is experimental and stores tokens in the operating system credential store.

---

## ✨ Why revenuecat?

> Stop swapping tokens and MCP configs. Start **operating every RevenueCat project from one CLI.**

| | |
|---|---|
| 🧭 **Named contexts** | Keep one local registry for all project-scoped API keys |
| 🔐 **OAuth login** | Shared public client with PKCE and OS credential-store token storage |
| 🤖 **Agent-first output** | Deterministic JSON envelopes for LLMs, scripts, and CI |
| 📦 **Project snapshots** | `pull project` and `pull all` for fast planning and comparison |
| 📊 **Metrics built in** | Overview and chart endpoints without hand-rolled curl calls |
| 🛠️ **Precise CRUD** | Apps, entitlements, products, offerings, packages, customers, subscriptions, purchases |

---

## 📦 Installation

<details open>
<summary><strong>Option 1 — Homebrew</strong> (recommended)</summary>

```bash
brew tap FerdiKT/homebrew-tap
brew install revenuecat
```

</details>

<details>
<summary><strong>Option 2 — Go install</strong></summary>

```bash
go install github.com/FerdiKT/revenuecat-cli/cmd/revenuecat@latest
```

</details>

<details>
<summary><strong>Option 3 — Build from source</strong></summary>

```bash
git clone https://github.com/FerdiKT/revenuecat-cli.git
cd revenuecat-cli
make build VERSION=v0.1.0
./bin/revenuecat version
```

</details>

<details>
<summary><strong>Option 4 — Local install for testing</strong></summary>

```bash
git clone https://github.com/FerdiKT/revenuecat-cli.git
cd revenuecat-cli
make install-local PREFIX="$(pwd)/.local-dev" VERSION=local
./.local-dev/bin/revenuecat version
```

</details>

---

## 🚀 Quickstart

Get up and running in **5 minutes**.

### 1️⃣ Add your first context

```bash
revenuecat contexts add ios-prod \
  --api-key sk_your_project_secret_key \
  --project-id proj_123 \
  --project-name "Main iOS App" \
  --active
```

### 2️⃣ Inspect context state

```bash
revenuecat contexts list --format table
revenuecat auth status
```

### 3️⃣ Pull the current project snapshot

```bash
revenuecat pull project --chart trials
```

### 4️⃣ Resolve an app and inspect country metrics

```bash
revenuecat apps resolve --context ios-prod --bundle-id app.ferdi.headson
revenuecat metrics countries revenue \
  --context ios-prod \
  --app app_1 \
  --start 2026-01-01 \
  --end 2026-04-16
```

### 5️⃣ Compare every configured project

```bash
revenuecat pull all --include-customers
```

### 6️⃣ Create resources with JSON payloads

```bash
revenuecat entitlements create --data '{"lookup_key":"pro","display_name":"Pro Access"}'
revenuecat offerings create --file ./payloads/offering-create.json
```

---

## 🗺️ Command Map

<table>
  <thead>
    <tr>
      <th>Group</th>
      <th>Commands</th>
      <th>Highlights</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td><code>contexts</code></td>
      <td>add · list · use · show · remove · verify</td>
      <td>Named API key contexts, active context switching</td>
    </tr>
    <tr>
      <td><code>auth</code></td>
      <td>status · login · logout</td>
      <td>API-key contexts plus experimental OAuth login/logout</td>
    </tr>
    <tr>
      <td><code>apps</code></td>
      <td>list · get · create · update · resolve</td>
      <td>App metadata, app registration updates, and bundle-id lookup</td>
    </tr>
    <tr>
      <td><code>entitlements</code></td>
      <td>list · get · create · update · archive · unarchive · attach-products · detach-products</td>
      <td>Access model management</td>
    </tr>
    <tr>
      <td><code>products</code></td>
      <td>list · get · create · update · archive · unarchive</td>
      <td>Catalog management</td>
    </tr>
    <tr>
      <td><code>offerings</code></td>
      <td>list · get · create · update · archive · unarchive</td>
      <td>Paywall-ready offering workflows</td>
    </tr>
    <tr>
      <td><code>packages</code></td>
      <td>list · get · create · update · attach-products · detach-products</td>
      <td>Offering package management</td>
    </tr>
    <tr>
      <td><code>metrics</code></td>
      <td>overview · chart · countries · options</td>
      <td>Overview KPIs, raw chart reads, and country breakdown tables</td>
    </tr>
    <tr>
      <td><code>customers</code>, <code>subscriptions</code>, <code>purchases</code></td>
      <td>list · get</td>
      <td>Customer-side inspection for support and analytics</td>
    </tr>
    <tr>
      <td><code>pull</code></td>
      <td>project · all</td>
      <td>Normalized agent snapshots</td>
    </tr>
    <tr>
      <td><code>version</code></td>
      <td>—</td>
      <td>Build and release metadata</td>
    </tr>
  </tbody>
</table>

---

## 🤖 Agent Workflow

Use this pattern for Codex, Claude, Cursor, or internal agents:

1. Resolve the target context with `revenuecat contexts list` or `--context`.
2. Pull current state first with `revenuecat pull project` or `revenuecat pull all`.
3. Plan mutations from the snapshot instead of guessing.
4. Use resource-specific `create`, `update`, `archive`, or attach/detach commands with JSON payloads.
5. Keep `--all-contexts` read-only.

Repo-local guidance also lives in [`AGENTS.md`](AGENTS.md) and [`skills/revenuecat-cli/SKILL.md`](skills/revenuecat-cli/SKILL.md).

For metrics charts that need complex RevenueCat filters or selectors, you can pass raw JSON directly instead of shell-fragile `--param` values:

```bash
revenuecat metrics chart revenue \
  --context ios-prod \
  --filters-json '[{"name":"app_id","values":["app_1"]},{"name":"store","values":["app_store"]}]' \
  --selectors-json '{"revenue_type":"revenue"}'
```

To install the bundled Codex skill directly from the CLI:

```bash
revenuecat agent install-skill
```

For local development, if you want a symlink instead of a copied install:

```bash
revenuecat agent link-skill --source ./skills/revenuecat-cli
```

---

## 🔐 Auth Model

The stable path uses **project-scoped RevenueCat API keys** organized into named contexts.

- Active context is the default target.
- `--context <alias>` overrides the active context.
- `--all-contexts` fans out read-only commands across every configured project.
- `revenuecat auth login` uses the shared public OAuth client with PKCE.
- OAuth access and refresh tokens are stored in the OS credential store: macOS Keychain, Windows Credential Manager, or Linux Secret Service.
- The local config file stores OAuth metadata only, not OAuth tokens.
- On Linux, a Secret Service provider such as GNOME Keyring, KWallet, or KeePassXC Secret Service must be available.

---

## 🧪 Development

```bash
make test
make build VERSION=v0.1.0
make install-local PREFIX="$(pwd)/.local-dev" VERSION=local
make dist VERSION=v0.1.0
```

---

## 📄 License

MIT. See [LICENSE](LICENSE).
