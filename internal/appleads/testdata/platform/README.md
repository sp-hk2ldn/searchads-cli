# Platform response fixtures

Captured 10 November 2026 through read-only Platform API requests for account 8023100. No credentials or request headers are included.

- `empty.json`: unchanged empty search-term response for 9 November.
- `single.json`: 3 November GB Category Exact response reduced to the named `sleep` row; pagination adjusted to one row. The provider also returns a privacy-suppressed row with a null search-term name.
- `gb-category-range.json`: unchanged 3–9 November response including that suppressed row and date-only granular metrics.
- `campaign.json`: first campaign from a recorded campaign query, reduced to one campaign.
- `page-200.json` / `page-last.json`: synthetic compact 200+1 keyword report pages; descending IDs on page one exercise deterministic sorting after pagination.

Regression tests additionally force an explicit empty terminal page after populated report fixtures, and reject missing rows without empty-page pagination metadata.
