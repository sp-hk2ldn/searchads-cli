# searchads-cli (Go)

Lightweight Apple Search Ads CLI in Go.

## Command Surface
- `searchads status`
- `searchads campaigns [list|find|create|pause|activate|delete|update-budget|set-budget|report] [flags] [--json]`
- `searchads adgroups [list|find|create|pause|activate|delete|report] [flags] [--json]`
- `searchads ads [list|find|get|create|update|pause|activate|delete] [flags] [--json]`
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

Full command and flag docs: [docs/COMMANDS.md](docs/COMMANDS.md)
Open source release checklist: [docs/OPEN_SOURCE_RELEASE_CHECKLIST.md](docs/OPEN_SOURCE_RELEASE_CHECKLIST.md)
Contributor guide: [CONTRIBUTING.md](CONTRIBUTING.md)
Security policy: [SECURITY.md](SECURITY.md)
Code of conduct: [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md)

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

## Install (Homebrew)
```bash
brew tap sp-hk2ldn/tap
brew install sp-hk2ldn/tap/searchads
searchads --help
```

## Build / Run From Source
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

## Reporting API Limits
- Impression Share (`sov-report`) generation is limited by Apple Ads to **10 reports per rolling 24 hours** per org.
- Custom report listing uses a maximum page size of **50** (`/custom-reports?limit=50`).
- The custom reports API is rate-limited (Apple docs indicate **150 requests per 15 minutes** for listing), so callers should use retry/backoff on `429`.
- Practical guidance: prefer `searchads reports list/get/download` for existing reports and only trigger `searchads sov-report` when needed.
