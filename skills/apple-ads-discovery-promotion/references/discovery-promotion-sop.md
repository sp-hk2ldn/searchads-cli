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

## Platform API execution
- Read the [agent guidance](../../../docs/PLATFORM_API_AGENT_GUIDE.md) before executing the workflow.
- Use performance reports for promotion thresholds; Insights alone has no taps, installs, spend, or CPA evidence.
- Create and read back the destination exact keyword before adding its Discovery negative. Keyword match type cannot be changed in place.
- A failed or interrupted write is not safe to retry automatically. Read current state, reconcile each intended keyword/negative, and resume only missing authorized changes.
- Record the ad account and report timezone alongside the window. Search-term reports use account timezone (`ORTZ`); other performance reports use UTC.

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
