# Apple Ads Platform API migration

The CLI now calls `https://api.ads.apple.com/v1`. Campaign Management API v5 is scheduled to shut down on 26 January 2027. There is no runtime v5 fallback.

Sources: [Platform API](https://developer.apple.com/documentation/apple-ads-platform-api), [request guide](https://developer.apple.com/documentation/apple-ads-platform-api/calling-apple-ads-platform-api), [account access](https://developer.apple.com/documentation/apple-ads-platform-api/access-overview), [reports](https://developer.apple.com/documentation/apple-ads-platform-api/reports), [Insights](https://developer.apple.com/documentation/apple-ads-platform-api/insights-endpoints), [budget orders](https://developer.apple.com/documentation/apple-ads-platform-api/budget-orders-endpoints). Verified against documentation and read-only requests on 10 November 2026.

## Authentication and account selection

The existing managed credential route remains valid: ES256 client-secret JWT, `https://appleid.apple.com/auth/oauth2/token`, `client_credentials`, scope `searchadsorg`. No new key is required by this migration.

`GET /me` returns organization identity; `GET /acls` discovers accessible ad accounts. Account selection, in precedence order:

1. `--adAccountId <id>` after the command/action.
2. `SEARCHADS_AD_ACCOUNT_ID`.
3. `adAccountId` in `SEARCHADS_CREDENTIALS_JSON`.
4. The sole accessible account, only if there is exactly one.

Multiple accounts without selection are an error, listing candidate IDs/names without issuing an account-scoped request. An inaccessible selection is also an error. Legacy `orgId` / `SEARCHADS_ORG_ID` is not reinterpreted as an ad account ID. `status` prints both `orgId` and `adAccountId` and exits nonzero on failure. OAuth cache identity includes account selection.

Ad-account operations send `X-AP-Context: adAccountId=<id>`. Budget-order create/update are organization-scoped and deliberately omit this header; read requests retain it for account-limited callers. These organization-scoped reads may return organization-wide orders for an administrator.

## Command-to-endpoint inventory

Paths below are relative to `/v1`. Existing aliases follow the same implementation.

| CLI operations | v5 resource/pattern | Platform API implementation |
| --- | --- | --- |
| `status` | `/me`, implicit parent org | `GET /me`, `GET /acls`, explicit account resolution |
| `campaigns list/find` | `GET /campaigns` | `POST /campaigns/query`, all pages, existing CLI filters |
| `campaigns create` | `POST /campaigns` | `POST /campaigns`, promoted object, targeting, daily budget and bid strategy objects |
| `campaigns pause/activate`, budget and bidding updates | wrapped `PUT /campaigns/{id}` | flat partial `PUT /campaigns/{id}` |
| `campaigns delete` | `DELETE /campaigns/{id}` | `DELETE /campaigns/{id}` |
| `campaigns report` | aggregate nested ad-group reports | existing aggregation preserved over `/adgroups/query` and `/reports/apps/adgroups/query` |
| `adgroups list/find/create/pause/activate/delete` | nested campaign paths | `/adgroups/query`, `/adgroups`, `/adgroups/{id}`; campaign filter/body and ownership validation |
| `adgroups report` | `/reports/campaigns/{id}/adgroups` | `POST /reports/apps/adgroups/query` |
| `ads list/find/get/create/update/pause/activate/delete` | campaign/ad-group paths, `/ads/find` | `/ads/query`, `/ads`, `/ads/{id}`; parent checks before item reads/writes |
| `ads report` | nested reports path | `POST /reports/apps/ads/query` |
| `creatives list/find/get/create` | `/creatives`, `/creatives/find` | `/creatives/query`, `/creatives/{id}`, `/creatives`; destination object |
| `keywords list/find` | nested `targetingkeywords` | `POST /keywords/query` with ad-group filter |
| `keywords add` | `targetingkeywords/bulk` | `POST /keywords/bulk-create` with correlated items |
| `keywords pause/activate/rebid/pause-by-text/remove/delete` | bulk/nested keyword paths | `PUT` or `DELETE /keywords/{id}`; parent and immutable match-type checks |
| `keywords report` | nested keyword report | `POST /reports/apps/keywords/query` |
| `negatives list/add/pause/activate/remove/delete` | campaign/ad-group `negativekeywords` and fallback paths | `/negative-keywords/query`, `/negative-keywords/bulk-create`, `/negative-keywords/{id}`; campaign-only queries require null adGroupId |
| `searchterms report` | nested search-term report | `POST /reports/apps/searchterms/query`; account reporting timezone |
| `product-pages list/get/locales` | `/apps/{id}/product-pages/...` | `/product-pages/query`, `/product-pages/{id}`, `/product-pages/locale-details/query` |
| `apps search/get` | `/search/apps`, `/apps/{id}` | corresponding Platform GET endpoints |
| `apps localized-details` | GET locale details | `POST /apps/{id}/locale-details/query` |
| `apps eligibility` | `/apps/{id}/eligibilities/find` | `/eligibilities/apps/query`; local legacy filters/window; app name enriched from `/apps/{id}` |
| `geo search/get` | `/search/geo`, `/geodata` | GET or POST `/search/geo`, supply source APPSTORE; get requires entity |
| `ad-rejections find/get` | `/product-page-reasons` | `/rejection-reasons/apps/query`; get searches returned reasons by ID |
| `budget-orders list/get/create/update` | `/budgetorders` | `/shared-budgets/query`, `/shared-budgets/{id}`, `/shared-budgets`; value, dates, invoice details and ad-account assignment |
| `sov-report` | create/poll/download `/custom-reports` | synchronous `POST /insights/apps/impression-share/query` |
| `reports list/get/download` | saved custom report jobs | explicit unsupported error; use `sov-report` for current Insights data |
| `product-pages countries/devices`, `ad-rejections assets` | old country/device/App Store asset endpoints | explicit unsupported error; no same-semantics documented endpoint (Maps assets are not App Store assets) |

## Compatibility decisions

- Stable command names, flags, output envelopes and summary fields are retained where meanings match. Entity `result` payloads are decoded into the existing CLI summaries. Missing optional fields remain absent; richer system-status values follow the new API. App details now decode `id`, `artistName`, and `iconPictureUrl`; unavailable legacy genre IDs and single-country fields cannot be reconstructed from names/storefront arrays.
- `--totalBudgetAmount` on campaign creation is rejected: a shared budget needs separate creation/assignment and must not silently replace a lifetime campaign cap. `--budgetAmount` remains the existing alias for the **daily** budget, not a lifetime budget.
- `--targetCpa` is rejected: the new `bidStrategy.bid` upper bound is not the old target CPA. `MAX_CONVERSIONS` can be selected without that flag, with goal `INSTALL`; `MANUAL_CPT` uses goal `TAP`. No unsupported monetary input is silently discarded.
- Ad creation now requires `--name`, as required by Platform API. Ads read queries join creative details to preserve the `creativeType` filter.
- Keyword match type is immutable. Existing mutation code may echo the current type; an attempted change fails before mutation. `ACTIVE` inputs map to `ENABLED`; keyword filters accept both names. Returned statuses use the Platform vocabulary.
- `budget-orders create --orgId` is rejected rather than treating organization IDs as account IDs. New orders apply to the selected `--adAccountId`. Budget-order JSON adds `adAccountIds` and `systemStatus` rather than relabelling them as legacy `orgIds` or approval status. Line-of-credit permission and invoice requirements remain Apple's responsibility; no live billing writes were tested.
- `geo get` additionally requires `--entity Country|AdminArea|Locality`; existing uppercase names also work. Product-page `--expand` is accepted as redundant because the new endpoint includes expanded metadata.
- Find selectors unsupported by a new server query are applied locally to normalized records. Existing offset/limit windows are retained. Collection/report requests drain pages instead of truncating after the first response.
- Reports use `timeRange`, `filters`, `pagination.offset/pageSize`, and `result.rows` with `totalMetrics`/`granularMetrics`. One-day requests omit granularity. Search-term reports require `ORTZ`, now exposed as `timeZone` in JSON; other reports retain UTC. Empty days may lack the optional detailed `metrics` object even when top-level zero totals are present.
- Live ad-group reporting requires the entity filter `id`; the overview's `adGroupId` wording was rejected by the provider. Ads' own-ID filters similarly use `id`.
- Live eligibility pagination ends with pagination metadata (`pageSize: 0`) and no `result`. Only this explicit terminal-query form is accepted as an empty page; other missing results fail.

## SOV replacement

`--reportType ALL_SLOTS|FIRST_SLOT` selects the measure; default is `ALL_SLOTS`. `LAST_WEEK`, `LAST_2_WEEKS`, and `LAST_4_WEEKS` mean completed Sunday–Saturday weeks in UTC. The default remains four weeks. `--name` is local export metadata, not a server-side saved report.

The CSV, normalized JSON and decision-table paths remain. CSV is generated locally from paginated Insights results. `reportId` is null; `state: COMPLETED` describes the completed local fetch. Share lower/upper bounds and `searchPopularity1to5` retain their native meanings. A single `impressionShare` is only supplied when both bounds are equal. Old 1–100 popularity, spend, taps, impressions and installs are null when unavailable; no conversion or budget recommendation is inferred from missing metrics. The decision table directs the caller to review performance data alongside Insights.

There is no saved-job list/get/download replacement and no fabricated job ID. The old 10-report creation quota does not describe this synchronous query workflow; provider rate-limit errors are surfaced.

## Verification and release boundary

Baseline Go tests passed before migration. Tests now cover distinct account selection/cache scoping, pagination/non-progress, granular/single-day report decoding, parent ownership, exact request bodies, per-item bulk failure, and single mutation dispatch on connection loss and 404/429/500/redirect responses. Existing golden command tests remain in use. Five focused new test functions supplement updated existing cases; all write verification is mocked. `go test ./...`, `go build ./...`, `go vet ./...`, formatting/diff checks, and all 18 missing-credential golden parity cases passed.

Live verification used the owner's explicitly selected managed account. Both APIs returned the same 13 enabled campaign IDs and core fields. A historical 2–8 November 2026 report with nonzero spend matched campaign totals and daily top-level spend, impressions, taps and installs; optional empty-day metric objects differed as described above.

Read-only smoke checks passed for ad groups, keyword reads/reports, search-term reports, negatives, ads, creatives, app details/locales/eligibility, product-page details/locales, app search, creative details, ads reporting, geography search/lookup, rejection reasons, budget-order listing and SOV. Some collections were empty; that proves request acceptance and empty decoding, not populated coverage. App locales initially returned a provider 500 with null filters; sending an empty filter array succeeded. Live pagination was exercised on eligibility data. Neither API was used for live campaign/billing writes.

Before release: review the migration diff and compatibility changes, run the local Go tests/build/vet and parity checks, then approve publication separately. Homebrew changes and live write smoke tests are separate approvals. Repository CI currently references GitHub-hosted runners; no workflow was dispatched or PR published by this migration. Resolve CI execution policy before an external PR/release action.

## Validation corrections (10 November 2026)

- Empty report pages observed live return `result: {}` with explicit `pagination.pageSize: 0`. Accept that terminal shape without discarding previously collected rows. Missing rows without zero-page metadata still fail. Recorded fixtures cover empty, single-row and GB Category Exact daily results; synthetic 200+1 page fixtures verify full pagination and ID ordering.
- Campaign `servingStatus` and `servingStateReasons` expose native `systemStatus` and `systemStatusReasons` values. `adChannelType` maps App Store search-results placements to `SEARCH`, and search-tab/Today-tab/product-page placements to `DISPLAY`. Unknown placements remain unclassified.
- `targetCpa` is explicitly null when unavailable. The Platform API has no same-semantics equivalent: [BidStrategy.bid](https://developer.apple.com/documentation/apple-ads-platform-api/bidstrategy) is an auction monetary ceiling, exposed separately as `bidCeiling`. It must not be relabelled as a target CPA.
- Date-only ad-group metrics (empty days) expose zero installs and the campaign daily-budget currency, fetched read-only when needed. Missing campaign currency fails rather than inventing a currency. Detailed metrics remain absent on those days.
- Keyword outputs retain native `ENABLED` (v5 `ACTIVE`); both filter/input names remain accepted. Collections sort by entity ID; keyword daily reports sort by keyword ID then date. Insights CSV uses fixed decimal notation for numeric values, preserving integer app IDs.
- PR workflows require `[self-hosted, searchads-cli]`. No repository runner was registered at validation time; checks may remain queued pending runner provisioning. Local tests/build/vet/parity checks are separate evidence.
