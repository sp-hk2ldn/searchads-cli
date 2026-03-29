---
name: apple-ads-discovery-promotion
description: Discovery -> Promotion -> Isolation operational skill for Apple Ads. Use when converting Search Match/broad search terms into exact-match keywords and maintaining negative-keyword isolation with auditable thresholds.
---

# Discovery -> Promotion -> Isolation

## Use when
- A user asks for weekly keyword optimization workflow.
- A user asks how to promote search terms from discovery campaigns.
- A user wants an automation spec for safe keyword expansion and isolation.

## Workflow
1. Pull Discovery search term and keyword/ad-group performance data for the configured window.
2. Normalize and dedupe terms; exclude terms already added as exact keywords or negatives.
3. Score terms by promotion thresholds (taps, installs, CPA, TTR).
4. Classify winners into Brand/Category/Competitor targets.
5. Promote winners as exact-match keywords in target ad groups.
6. Add promoted terms as exact-match negatives in Discovery ad groups.
7. Emit audit logs and reasons per term (why promoted/skipped).

## Default policy
- Weekly cadence (or twice weekly).
- 14-day lookback.
- Exclude most recent 24-48 hours for lag.
- Dry-run by default.
- Cap number of new keywords per run.
- No deletes or budget changes in the same run.

## Guardrails
- No destructive changes without explicit user confirmation.
- Do not remove discovery coverage before replacement exact terms are active.
- Keep Search Match isolation behavior explicit in output recommendations.
- Treat pause, negate, delete, and bulk deactivation actions as recommendation-only until the user explicitly approves execution.
- Respect Apple Ads report API limits in reporting plans:
  - Impression Share report creation (`POST /api/v5/custom-reports`): max 10 reports per 24 hours.
  - Impression Share report listing (`GET /api/v5/custom-reports`): max `limit` 50 and 150 requests per 15 minutes.
- If `429` occurs on report creation, prefer waiting for the rolling 24-hour window over aggressive retries.

## References
- Read `references/discovery-promotion-sop.md` for thresholds and automation spec fields.
