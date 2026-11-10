# Platform API guidance for agents

Applies to the migrated Platform API CLI. Check the binary version and its help before applying these commands to an installed release. See [migration inventory and compatibility decisions](PLATFORM_API_MIGRATION.md) and [command syntax](COMMANDS.md) for details. The canonical checkout and installed binary may still be on v5 until this branch is integrated and released.

## Account and credentials

Use the existing managed credential injection route; never print or copy private keys into scripts or reports. Supply `--adAccountId <chosen-id>` after the command/action, or an explicitly configured `SEARCHADS_AD_ACCOUNT_ID` / credential `adAccountId`. Run `status` to verify the selected account before account-scoped work. Organization IDs (`orgId`, `SEARCHADS_ORG_ID`) are not account selectors. When multiple accessible accounts remain ambiguous, ask the user to choose; never choose the first account or reuse an unrelated app's account.

Record account ID, app/campaign filters, date range and timezone in analysis artifacts. Collection queries paginate automatically; explicit find limits still constrain returned results.

## Impression share / SOV

Use `sov-report --adamId <app-id> --adAccountId <chosen-id> --dateRange LAST_WEEK --reportType ALL_SLOTS --json` for a read-only Insights query. `ALL_SLOTS` and `FIRST_SLOT` measure different slot scopes; select deliberately and compare like for like. Supported windows are completed Sunday–Saturday UTC weeks: `LAST_WEEK`, `LAST_2_WEEKS`, `LAST_4_WEEKS` (default).

This is a synchronous, paginated query, not report creation. Do not create, poll or download a saved job. `reports list/get/download` are unsupported; `reportId` is null, `state: COMPLETED` means the local fetch completed, and `--name` is local metadata. CSV exports are generated locally. The old ten-report creation quota and saved-report listing limits do not apply; surface provider throttling without inventing a reset interval.

Preserve `lowImpressionShare` / `highImpressionShare` bounds and `searchPopularity1to5`. Do not convert popularity to the old 1–100 scale, apply old thresholds such as 60, collapse ranges into exact shares, or treat null metrics as zero. Insights does not supply the old performance evidence for automatic bid/budget decisions. Join campaign, keyword or search-term performance reports before recommending action; label unavailable evidence explicitly.

Search-term performance reports use account timezone (`ORTZ`); other performance reports use UTC. Account for this difference when comparing days. Missing optional daily metric objects are not evidence of a decoding failure or proof of zero activity.

## Writes and unsupported operations

The client sends each mutation once, with no automatic retry, redirect, authentication replay or v5 fallback. Agents and wrapper scripts must preserve that boundary. After a timeout, connection loss or error, read current state before deciding whether a manual retry is appropriate. For multi-step or bulk workflows, reconcile each item; an error does not prove earlier changes were rolled back. Only resume changes already authorized by the user.

- Keyword match type is immutable. Create/read back a replacement before adding isolation negatives; do not change the existing keyword's type in place.
- Campaign `--targetCpa` and `--totalBudgetAmount` are unsupported. A bid upper bound or shared budget must not silently substitute for these semantics. A strategy's target CPA can remain an analysis threshold without being sent as a campaign flag.
- `ads create` requires `--name`; `geo get` requires `--entity`.
- `budget-orders create --orgId` is unsupported; use explicit ad-account selection. Shared-budget permissions and billing fields require separate consideration; do not infer organization-wide write authorization from campaign access.
- `product-pages countries/devices` and `ad-rejections assets` have no same-semantics documented replacement. Report the unsupported operation; do not substitute Maps assets or invent endpoints.

Migration verification covered live reads and mocked writes. It did not prove live campaign or billing mutations. Publishing, Homebrew updates and live write verification remain separate approval steps for this migration.
