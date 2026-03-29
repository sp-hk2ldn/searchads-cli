# Placement and Creative Cheatsheet (Extracted)

## Placements
- **Today tab ads**
  - App Store front page exposure.
  - Uses selected custom product page assets for animated background.
  - Tap destination is the selected custom product page.
- **Search tab ads**
  - Top of Search tab.
  - Destination can be default or custom product page.
  - Custom product page must be localized for target country/region primary language.
- **Search results ads**
  - Top of relevant search results.
  - Only one ad may appear at top.
  - Search results ad variation can use approved custom product pages.
  - Default search results ad fallback applies in unsupported/hold scenarios.
- **Product pages (while browsing)**
  - Appear in "You Might Also Like" area when browsing product pages.

## Creative basics
- Ads are generated from App Store Connect metadata.
- Core visible elements include app name, icon, subtitle.
- Search results can include screenshots/app previews.
- Up to 10 screenshots; up to 3 app previews (up to ~30 seconds each, muted by default in store contexts).
- Keep text overlays simple and benefit-focused.

## Custom product pages
- Can highlight alternate features vs default product page.
- Can include unique URL and optional deep links.
- Usable for search results variations, Search tab ads, and Today tab ads.
- Can be created without shipping a new app version.

## Quick review checklist
- Placement fit: aligns with acquisition goal and intent depth.
- Localization: CPP localized for target country/region requirements.
- Variation behavior: one active variation per ad group, with fallback awareness.
- Policy safety: privacy-safe targeting assumptions and content guideline compliance.
- Measurement ops: Impression Share reports are constrained by API limits (max 10 creates per 24 hours; report listing max `limit=50` and 150 requests per 15 minutes).

## Placement selection rubric
- Prefer **Search results ads** when:
  - the goal is high-intent acquisition against known queries
  - keyword coverage and query intent are already understood
  - the default product page or CPP is conversion-ready for search traffic
- Prefer **Search tab ads** when:
  - the goal is broader consideration before a user types a query
  - you want to test whether a CPP expands reach beyond strict keyword intent
  - you can support locale-specific CPP coverage in target markets
- Prefer **Today tab ads** when:
  - the goal is broad awareness or a major launch moment
  - the creative is strong enough for premium storefront placement
  - budget can tolerate higher CPT and weaker short-term efficiency
- Prefer **Product pages placements** when:
  - the goal is adjacent-demand capture while users browse similar apps
  - the listing is visually competitive without heavy keyword dependence

## Minimum readiness rules
- Search results:
  - require clear app name, icon, subtitle, and at least the first 3 screenshots to match the targeted intent
  - if using CPP variation, ensure the CPP is approved and maps to the intended keyword theme
- Search tab:
  - require a localized CPP when routing to CPP in that country or region
  - require a value proposition that can work with lower explicit intent than search results
- Today tab:
  - require a strong visual story, launch-worthy creative angle, and CPP assets that can carry premium placement
  - avoid recommending it when the user cannot tolerate awareness-style spend or weaker direct-response efficiency
- Product pages:
  - require competitive screenshots and app preview assets because placement relies heavily on browsing appeal

## Experiment default thresholds
- Express success relative to current baseline, not a universal number:
  - success: CPA at or below baseline, conversion rate at or above baseline, and CPT no more than 20% above baseline unless the goal is awareness
  - caution: CPT more than 20% above baseline with flat conversion rate
  - failure: CPA more than 20% above baseline after sufficient volume
- Minimum decision volume before strong conclusions:
  - at least `20 taps` for early read
  - at least `5 installs` before calling a placement clearly efficient or inefficient

## Output expectations
- For any recommendation, state:
  - chosen placement
  - why the placement fits the goal
  - readiness gaps to fix before launch
  - baseline comparison needed to judge success
