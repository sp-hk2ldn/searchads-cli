# Apple Ads Certification Findings (Extracted)

Source extraction: local research notes, section "Apple Ads Certification Findings (Observed 2026-01-31)".

## Observed guidance
- Use keyword themes across Brand, Category, Competitor, and Discovery.
- Use exact match in Brand/Category/Competitor ad groups.
- Use broad + Search Match for Discovery-oriented keyword mining.
- Keep Search Match configuration at ad-group level and use it intentionally for discovery.
- Use keyword suggestions as a seed source, then promote terms based on performance.
- Add exact-match negatives in discovery ad groups for promoted terms to reduce overlap.
- For search results ad variations, approved custom product pages can be selected; one active variation per ad group.
- Default search results ad behavior applies when variation is unavailable (for example, unsupported iOS or variation on hold).
- Reporting and optimization rely on campaign/ad-group/keyword/ad dashboards, custom reports, and impression share/rank recommendations.
- Impression Share report API constraints:
  - `POST /api/v5/custom-reports`: up to 10 report generations within 24 hours.
  - `GET /api/v5/custom-reports`: max `limit` 50; rate limit 150 requests within 15 minutes.

## Performance and attribution
- Analyze taps, spend, CPT, installs, conversion rate, and CPA.
- Use tap-through and view-through installs correctly when interpreting impact.
- Use Attribution API/MMP signals where available to connect performance to value.

## Default review thresholds
- Treat any recommendation below these minimums as provisional:
  - keyword or search-term review minimum: `taps >= 10`
  - efficiency minimum for stronger action: `installs >= 2`
  - conversion-rate context: compare against campaign or ad-group median rather than a universal fixed target
  - tap-through-rate context: compare against campaign or ad-group median rather than a universal fixed target
- Default action rules:
  - Observe only when a term or keyword has fewer than 10 taps, or fewer than 2 installs, unless spend is materially above target.
  - Recommend bid increase only when CPA is at or below target, conversion rate is at or above local median, and volume is proven.
  - Recommend bid reduction or tighter matching when CPA is above target and volume is sufficient to judge.
  - Recommend restructure when Brand, Category, Competitor, and Discovery intent are mixed in the same ad group or when Search Match is used outside Discovery without a clear exception.
  - Recommend adding negatives when promoted exact terms are still competing with Discovery traffic.

## Output expectations
- When giving advice, include:
  - the metric basis used
  - whether the guidance is observed certification guidance or inferred working policy
  - the next action category: observe, refine bids, add negatives, or restructure

## Inferred working policy
- Discovery is a term-mining pipeline.
- Promotion should be threshold-based (quality + efficiency).
- Isolation should be enforced through exact-match negatives in discovery.
