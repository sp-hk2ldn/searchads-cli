# searchads-cli

Lightweight Apple Search Ads CLI in Go.

It exists because clicking around the Apple Search Ads dashboard gets old quickly once you start treating acquisition as something you want to inspect, script, and hand to agents with a bit of human approval around the dangerous bits.

Use it to work with campaigns, ad groups, ads, creatives, product pages, apps, geo, keywords, search terms, negative keywords, custom reports, budget orders, and SOV reports from the terminal.

## Install

```bash
brew tap sp-hk2ldn/tap
brew install sp-hk2ldn/tap/searchads
searchads --help
```

Or run from source:

```bash
go build ./...
go run ./cmd/searchads --help
```

## Quick examples

```bash
# Check auth/org discovery
searchads status

# Find paused campaigns as JSON
searchads campaigns find --status PAUSED --json

# Pull campaign performance for a date range
searchads campaigns report \
  --startDate 2026-02-01 \
  --endDate 2026-02-07 \
  --json

# Pull keyword performance for one ad group
searchads keywords report \
  --campaignId 123 \
  --adGroupId 456 \
  --startDate 2026-02-01 \
  --endDate 2026-02-07 \
  --minTaps 5 \
  --json

# Inspect search terms that are actually spending
searchads searchterms report \
  --campaignId 123 \
  --startDate 2026-02-01 \
  --endDate 2026-02-07 \
  --minSpend 5 \
  --json

# Add obvious waste as negatives
searchads negatives add \
  --campaignId 123 \
  --text "free" \
  --text "wallpaper" \
  --matchType EXACT \
  --json

# Download a completed custom report
searchads reports download \
  --reportId 987654 \
  --out reports/custom/987654.csv \
  --json
```

IDs above are examples. Don't paste real campaign/account output into public issues or screenshots without redacting it first.

## Why this is useful with agents

The useful bit is not that this is a fancy wrapper. It is that Apple Search Ads state can become structured terminal output.

That means Codex, Claude, OpenClaw, or a normal script can do things like:

- summarize campaign/report JSON
- point out spend anomalies
- find keywords or search terms that probably need attention
- prepare negative-keyword changes for review
- generate a short daily/weekly acquisition note

The recommended pattern is **read freely, mutate carefully**:

1. Run read/report commands with `--json`.
2. Let the agent summarize or propose changes.
3. Review the exact command it wants to run.
4. Only then run budget, pause, activate, bid, or negative-keyword changes.

## Command surface

- `searchads status`
- `searchads campaigns [list|find|create|pause|activate|delete|update-budget|set-budget|set-bidding-strategy|report] [flags] [--json]`
- `searchads adgroups [list|find|create|pause|activate|delete|report] [flags] [--json]`
- `searchads ads [list|find|get|create|update|pause|activate|delete|report] [flags] [--json]`
- `searchads creatives [list|find|get|create] [flags] [--json]`
- `searchads product-pages [list|get|locales|countries|devices] [flags] [--json]`
- `searchads apps [search|get|localized-details|eligibility] [flags] [--json]`
- `searchads geo [search|get] [flags] [--json]`
- `searchads ad-rejections [find|get|assets] [flags] [--json]`
- `searchads keywords [list|find|report|add|pause|activate|remove|rebid|pause-by-text] --campaignId <id> --adGroupId <id> [flags] [--json]`
- `searchads searchterms report --campaignId <id> [--adGroupId <id>] --startDate YYYY-MM-DD --endDate YYYY-MM-DD [--minTaps N] [--minSpend X] [--json]`
- `searchads negatives [list|add|remove|pause|activate] --campaignId <id> [--adGroupId <id>] [--negativeKeywordId <id> ...] [--text <kw> ...] [--matchType EXACT|BROAD] [--json]`
- `searchads sov-report --adamId <id> [--country GB,US] [--dateRange LAST_4_WEEKS] [--out reports/sov] [--json]`
- `searchads reports [list|get|download] [--reportId <id>] [--state COMPLETED] [--nameContains text] [--limit N] [--out reports/custom/id.csv] [--json]`
- `searchads budget-orders [list|get|create|update] [--budgetOrderId <id>] [--budgetAmount N] [--budgetCurrency GBP] [--json]`

Full command and flag docs: [docs/COMMANDS.md](docs/COMMANDS.md)

## Credentials

The CLI supports either:

- `SEARCHADS_CREDENTIALS_JSON` with JSON fields:
  - required: `clientId`, `teamId`, `keyId`, `privateKey`
  - optional: `orgId`, `popularityAdamId`, `popularityAdGroupId`, `popularityWebCookie`, `popularityXsrfToken`
- split env vars:
  - `SEARCHADS_CLIENT_ID`, `SEARCHADS_TEAM_ID`, `SEARCHADS_KEY_ID`, `SEARCHADS_PRIVATE_KEY`

For local development, start from [.env.example](.env.example), copy it to `.env`, then load it into your shell before running the CLI:

```bash
cp .env.example .env
set -a
source .env
set +a
go run ./cmd/searchads status
```

Do not commit `.env`, private keys, Apple Ads account IDs, campaign IDs, or raw report output that identifies your account.

## Build / run from source

```bash
go build ./...
go run ./cmd/searchads --help
go run ./cmd/searchads status
```

## Tests

Golden command stability tests are in:

- `cmd/searchads/main_golden_test.go`
- `cmd/searchads/testdata/golden/*.json`

Run:

```bash
go test ./...
```

## Notes

- Auth flow: ES256 client-secret JWT, Apple token exchange, org discovery via `/api/v5/me`.
- Includes custom-report workflows for SOV and generic report download.
- Open source release checklist: [docs/OPEN_SOURCE_RELEASE_CHECKLIST.md](docs/OPEN_SOURCE_RELEASE_CHECKLIST.md)
- Contributor guide: [CONTRIBUTING.md](CONTRIBUTING.md)
- Security policy: [SECURITY.md](SECURITY.md)
- Code of conduct: [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md)

## Reporting API limits

- Impression Share (`sov-report`) generation is limited by Apple Ads to **10 reports per rolling 24 hours** per org.
- Custom report listing uses a maximum page size of **50** (`/custom-reports?limit=50`).
- The custom reports API is rate-limited (Apple docs indicate **150 requests per 15 minutes** for listing), so callers should use retry/backoff on `429`.
- Practical guidance: prefer `searchads reports list/get/download` for existing reports and only trigger `searchads sov-report` when needed.
