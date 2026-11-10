---
name: apple-ads-placements-creative
description: Apple Ads certification guidance for ad placements and creative setup, including Today tab, Search tab, Search results, Product pages, custom product pages, and ad variation behavior. Use for placement decisions and creative readiness checks.
---

# Apple Ads Placements and Creative

## Use when
- A user asks where ads appear and how each placement behaves.
- A user asks how custom product pages and ad variations are used per placement.
- A user asks for a certification-aligned creative readiness checklist.

## Workflow
1. Identify placement(s): Today tab, Search tab, Search results, Product pages.
2. Validate placement-specific constraints using the rubric in `references/placements-creative-cheatsheet.md`:
   - destination page behavior
   - localization requirements
   - creative asset source (default vs custom product page)
3. Validate creative readiness against the minimum checklist:
   - icon/subtitle clarity
   - screenshot/app preview suitability
   - policy-safe and age-appropriate expectations
4. Recommend one placement experiment plan with:
   - why that placement fits the acquisition goal
   - the minimum readiness gaps to fix before launch, if any
   - success and failure thresholds for CPT, conversion rate, and CPA relative to the current baseline

## Guardrails
- Do not claim multiple search-result ads can show simultaneously (only one may appear at top).
- Clearly state when default ad fallback behavior applies for search results variations.
- Keep localization requirements explicit when recommending custom product pages.
- Before running CLI commands or interpreting Insights, read the [Platform API agent guidance](../../docs/PLATFORM_API_AGENT_GUIDE.md). It covers explicit account selection, SOV metric limits, unsupported operations, and write recovery.

## References
- Read `references/placements-creative-cheatsheet.md` for extracted observed placement findings, readiness rules, and experiment rubric.
