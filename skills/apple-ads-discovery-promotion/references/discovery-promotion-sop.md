# Operational SOP: Discovery -> Promotion -> Isolation

## Cadence and window
- Cadence: weekly (or twice weekly).
- Lookback: 14 days.
- Exclude newest 24-48 hours for attribution lag.

## Inputs
- Discovery Search Terms report (Search Match + Broad ad groups).
- Keyword/ad-group metrics: taps, spend, CPT, installs.
- Attribution signals (AdServices/MMP where available).
- Config: target CPA, campaign IDs, brand/competitor lists.

## Promotion defaults
- `min_taps >= 10`
- `min_installs >= 2`
- `CPA <= target_CPA`
- `TTR >= discovery_median`

## Actions
1. Add qualifying terms to Brand/Category/Competitor as exact-match keywords.
2. Set initial bid to recommended low range (or recent CPT +10% if no recommendation).
3. Add same terms as exact-match negatives in Discovery ad groups.

## Review policy
- After 7 days, recommend pausing or negating promoted terms that fail tap/install/CPA thresholds, but require explicit user confirmation before executing any destructive action.
- Raise bids or create tailored custom product pages when terms perform strongly.

## API limit guardrails
- Impression Share report creation (`POST /api/v5/custom-reports`) is capped at 10 reports per 24 hours.
- Impression Share report listing (`GET /api/v5/custom-reports`) supports max `limit=50` and is rate-limited at 150 requests per 15 minutes.
- If creation requests return `429`, wait for the rolling 24-hour window to clear before retrying additional report creates.

## Explainability requirements
- Save per-term reason codes:
  - promoted_by_thresholds
  - skipped_low_volume
  - skipped_high_cpa
  - skipped_already_targeted
- Store snapshot reference for every decision.

## Suggested config stub
```yaml
lookback_days: 14
exclude_last_days: 2
min_taps: 10
min_installs: 2
max_cpa: target_cpa
min_ttr: discovery_median
initial_bid: recommended_low_range_or_cpt_plus_10pct
discovery_campaign_ids: []
brand_campaign_ids: []
category_campaign_ids: []
competitor_campaign_ids: []
brand_allowlist: []
competitor_allowlist: []
```
