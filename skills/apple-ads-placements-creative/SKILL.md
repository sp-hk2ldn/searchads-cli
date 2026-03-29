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
- If recommending Impression Share report-based analysis, include Apple Ads API limits:
  - `POST /api/v5/custom-reports`: up to 10 reports per 24 hours.
  - `GET /api/v5/custom-reports`: max `limit` 50 and 150 requests per 15 minutes.

## References
- Read `references/placements-creative-cheatsheet.md` for extracted observed placement findings, readiness rules, and experiment rubric.
