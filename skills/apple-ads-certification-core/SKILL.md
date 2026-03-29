---
name: apple-ads-certification-core
description: Core Apple Ads certification guidance for keyword strategy, campaign structure (brand/category/competitor/discovery), Search Match usage, and performance interpretation. Use when planning or reviewing search-results campaign strategy against certification-aligned best practices.
---

# Apple Ads Certification Core

## Use when
- A user asks for Apple Ads strategy aligned with certification guidance.
- A user wants campaign/account structure recommendations for search results campaigns.
- A user wants to understand how to use Search Match, exact/broad/negative keywords, and keyword suggestions in a certification-consistent way.

## Workflow
1. Confirm goal and campaign type (new setup, restructure, or optimization review).
2. Use certification-aligned structure:
   - Separate Brand, Category, Competitor, and Discovery campaigns.
   - Keep Search Match focused on Discovery ad groups.
3. Validate keyword/match-type setup against the checklist and default thresholds in `references/certification-findings.md`.
4. Recommend concrete next actions with measurable thresholds:
   - Observation only when data volume is below review minimums.
   - Bid or budget adjustment only when efficiency is acceptable and volume is proven.
   - Restructure when query intent, match-type hygiene, or campaign separation is breaking attribution clarity.
5. Clearly label observed guidance vs working-policy inference.

## Guardrails
- Do not claim unsupported Apple Ads API endpoints.
- Prefer observed certification guidance; mark inferred automation policy as inferred.
- Keep recommendations region-sensitive for Search Match and localization constraints.
- Apply documented Apple Ads report API limits in all advice:
  - Impression Share report creation (`POST /api/v5/custom-reports`): up to 10 reports per 24 hours.
  - Impression Share report listing (`GET /api/v5/custom-reports`): max `limit` is 50; endpoint rate limit is 150 requests per 15 minutes.
- When users hit `429` on report creation, advise waiting for the 24-hour window to roll instead of repeatedly retrying.

## References
- Read `references/certification-findings.md` for extracted observed findings, review thresholds, and inferred policy notes.
