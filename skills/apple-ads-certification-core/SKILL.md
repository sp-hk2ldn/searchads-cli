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
- Before running CLI commands or interpreting Insights, read the [Platform API agent guidance](../../docs/PLATFORM_API_AGENT_GUIDE.md). It covers explicit account selection, SOV metric limits, unsupported operations, and write recovery.

## References
- Read `references/certification-findings.md` for extracted observed findings, review thresholds, and inferred policy notes.
