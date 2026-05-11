# Command Reference

All commands support `--json` unless otherwise noted.

## status
- `searchads status`

## campaigns
- `searchads campaigns list`
- `searchads campaigns find [--campaignId <id> ...] [--adamId <id> ...] [--status ENABLED,PAUSED] [--nameContains text]`
- `searchads campaigns create --name <name> --dailyBudgetAmount <number> [--totalBudgetAmount <number>] [--budgetCurrency GBP] [--status ENABLED] [--adamId <id>] [--countries GB,US] [--startTime RFC3339] [--endTime RFC3339] [--supplySource APPSTORE_SEARCH_RESULTS,APPSTORE_SEARCH_TAB] [--adChannelType SEARCH] [--biddingStrategy MANUAL_CPT|MAX_CONVERSIONS] [--targetCpa <number>] [--targetCpaCurrency GBP]`
- `searchads campaigns pause --campaignId <id>`
- `searchads campaigns activate --campaignId <id>`
- `searchads campaigns delete --campaignId <id>`
- `searchads campaigns update-budget --campaignId <id> --dailyBudgetAmount <number> [--budgetCurrency GBP]`
- `searchads campaigns set-budget --campaignId <id> --dailyBudgetAmount <number> [--budgetCurrency GBP]`
- `searchads campaigns set-bidding-strategy --campaignId <id> --biddingStrategy MANUAL_CPT|MAX_CONVERSIONS [--targetCpa <number>] [--targetCpaCurrency GBP]`
- `searchads campaigns report --startDate YYYY-MM-DD --endDate YYYY-MM-DD [--nameIncludes text] [--nameExcludes text] [--includePaused]`

Notes:
- `--budgetAmount` is still accepted as a backwards-compatible alias for `--dailyBudgetAmount` on create/update budget commands.
- Apple Ads API 5.2.1 supports create-only campaign total budgets through `--totalBudgetAmount`.
- `MAX_CONVERSIONS` requires `--targetCpa` and only supports `APPSTORE_SEARCH_RESULTS`.

## adgroups
- `searchads adgroups list --campaignId <id>`
- `searchads adgroups find --campaignId <id> [--adGroupId <id> ...] [--status ENABLED,PAUSED] [--nameContains text]`
- `searchads adgroups create --campaignId <id> --name <name> [--defaultBid <number>] [--currency GBP] [--status ENABLED] [--automatedKeywordsOptIn] [--automatedKeywordsRequired]`
- `searchads adgroups pause --campaignId <id> --adGroupId <id>`
- `searchads adgroups activate --campaignId <id> --adGroupId <id>`
- `searchads adgroups delete --campaignId <id> --adGroupId <id>`
- `searchads adgroups report --campaignId <id> --startDate YYYY-MM-DD --endDate YYYY-MM-DD [--adGroupId <id>]`

## ads
- `searchads ads list --campaignId <id> --adGroupId <id>`
- `searchads ads find [--campaignId <id>] [--adGroupId <id>] [--status ENABLED,PAUSED] [--creativeType CUSTOM_PRODUCT_PAGE,DEFAULT_PRODUCT_PAGE] [--nameContains text] [--offset N] [--limit N]`
- `searchads ads get --campaignId <id> --adGroupId <id> --adId <id>`
- `searchads ads create --campaignId <id> --adGroupId <id> --creativeId <id> [--name text] [--status ENABLED|PAUSED]`
- `searchads ads update --campaignId <id> --adGroupId <id> --adId <id> [--name text] [--status ENABLED|PAUSED]`
- `searchads ads pause --campaignId <id> --adGroupId <id> --adId <id>`
- `searchads ads activate --campaignId <id> --adGroupId <id> --adId <id>`
- `searchads ads delete --campaignId <id> --adGroupId <id> --adId <id>`
- `searchads ads report --campaignId <id> --startDate YYYY-MM-DD --endDate YYYY-MM-DD [--adId <id> ...] [--adGroupId <id> ...]`

## creatives
- `searchads creatives list`
- `searchads creatives find [--nameContains text] [--type CUSTOM_PRODUCT_PAGE,DEFAULT_PRODUCT_PAGE,CREATIVE_SET] [--state VALID,INVALID] [--adamId <id> ...] [--offset N] [--limit N]`
- `searchads creatives get --creativeId <id>`
- `searchads creatives create --adamId <id> --name <creative name> [--type CUSTOM_PRODUCT_PAGE|DEFAULT_PRODUCT_PAGE] [--productPageId <uuid>]`

Notes:
- `--productPageId` is required for `CUSTOM_PRODUCT_PAGE` creates.

## product-pages
- `searchads product-pages list --adamId <id> [--state VISIBLE,HIDDEN] [--nameContains text]`
- `searchads product-pages get --adamId <id> --productPageId <id>`
- `searchads product-pages locales --adamId <id> --productPageId <id> [--expand]`
- `searchads product-pages countries [--code GB,US] [--nameContains text]`
- `searchads product-pages devices [--deviceClass IPHONE,IPAD] [--nameContains text]`

## apps
- `searchads apps search --query <text> [--returnOwnedApps] [--limit N] [--offset N]`
- `searchads apps get --adamId <id>`
- `searchads apps localized-details --adamId <id>`
- `searchads apps eligibility [--adamId <id> ...] [--countryOrRegion GB,US] [--supplySource APPSTORE_SEARCH_RESULTS] [--state ELIGIBLE,INELIGIBLE] [--eligible true|false] [--appNameContains text] [--offset N] [--limit N]`

## geo
- `searchads geo search --query <text> [--countryCode GB] [--entity COUNTRY|ADMIN_AREA|LOCALITY] [--limit N]`
- `searchads geo get --geoId <id>`

## ad-rejections
- `searchads ad-rejections find [--adamId <id> ...] [--productPageId <id> ...] [--reasonType <value> ...] [--reasonLevel <value> ...] [--reasonCode <value> ...] [--countryOrRegion GB,US] [--languageCode en-GB] [--supplySource APPSTORE_SEARCH_RESULTS] [--commentContains text] [--offset N] [--limit N]`
- `searchads ad-rejections get --reasonId <id>`
- `searchads ad-rejections assets --adamId <id> [--assetType APP_PREVIEW,SCREENSHOT] [--orientation LANDSCAPE,PORTRAIT] [--appPreviewDevice <value> ...] [--assetGenId <value> ...] [--includeDeleted] [--offset N] [--limit N]`

## keywords
- `searchads keywords list --campaignId <id> --adGroupId <id>`
- `searchads keywords find --campaignId <id> --adGroupId <id> [--keywordId <id> ...] [--text <exactText> ...] [--textContains partial] [--status ACTIVE,PAUSED] [--matchType BROAD,EXACT]`
- `searchads keywords report --campaignId <id> --adGroupId <id> --startDate YYYY-MM-DD --endDate YYYY-MM-DD [--minTaps N] [--minSpend X] [--keywordId <id> ...] [--text <exactText> ...] [--textContains partial] [--status ACTIVE,PAUSED] [--matchType BROAD,EXACT]`
- `searchads keywords add --campaignId <id> --adGroupId <id> --text <keyword> ... [--matchType BROAD|EXACT] [--status ACTIVE|PAUSED] [--bidAmount N] [--currency GBP]`
- `searchads keywords add --campaignId <id> --adGroupId <id> --file <csvOrJsonFile> [--matchType BROAD|EXACT] [--status ACTIVE|PAUSED] [--currency GBP]`
- `searchads keywords pause --campaignId <id> --adGroupId <id> (--keywordId <id> ... | --text <exactText> ...)`
- `searchads keywords activate --campaignId <id> --adGroupId <id> (--keywordId <id> ... | --text <exactText> ...)`
- `searchads keywords remove --campaignId <id> --adGroupId <id> (--keywordId <id> ... | --text <exactText> ...)`
- `searchads keywords rebid --campaignId <id> --adGroupId <id> --bidAmount <number> [--currency GBP] (--keywordId <id> ... | --text <exactText> ...)`
- `searchads keywords pause-by-text --campaignId <id> --adGroupId <id> --text <exactText> ...`

## searchterms
- `searchads searchterms report --campaignId <id> --startDate YYYY-MM-DD --endDate YYYY-MM-DD [--adGroupId <id>] [--minTaps N] [--minSpend X]`

## negatives
- `searchads negatives list --campaignId <id> [--adGroupId <id>]`
- `searchads negatives add --campaignId <id> [--adGroupId <id>] --text <keyword> ... [--matchType EXACT|BROAD]`
- `searchads negatives remove --campaignId <id> [--adGroupId <id>] (--negativeKeywordId <id> ... | --text <exactText> ...)`
- `searchads negatives pause --campaignId <id> [--adGroupId <id>] (--negativeKeywordId <id> ... | --text <exactText> ...)`
- `searchads negatives activate --campaignId <id> [--adGroupId <id>] (--negativeKeywordId <id> ... | --text <exactText> ...)`

## sov-report
- `searchads sov-report --adamId <id> [--country GB,US] [--dateRange LAST_4_WEEKS] [--name report_name] [--out reports/sov]`
- `--appId` is accepted as an alias for `--adamId`.

Outputs:
- raw CSV
- normalized JSON
- decision table JSON

## reports (Custom Reports)
- `searchads reports list [--state COMPLETED,FAILED] [--nameContains text] [--limit N]`
- `searchads reports get --reportId <id>`
- `searchads reports download --reportId <id> [--out reports/custom/<id>.csv]`

## budget-orders
- `searchads budget-orders list`
- `searchads budget-orders get --budgetOrderId <id>`
- `searchads budget-orders create --name <name> --budgetAmount <number> [--budgetCurrency GBP] [--orgId <id> ...] [--startDate RFC3339] [--endDate RFC3339] [--orderNumber text] [--clientName text] [--primaryBuyerName text] [--primaryBuyerEmail email] [--billingEmail email]`
- `searchads budget-orders update --budgetOrderId <id> [--name <name>] [--budgetAmount <number>] [--budgetCurrency GBP] [--startDate RFC3339] [--endDate RFC3339] [--orderNumber text] [--clientName text] [--primaryBuyerName text] [--primaryBuyerEmail email] [--billingEmail email]`

## Useful examples
```bash
# Find paused campaigns
searchads campaigns find --status PAUSED --json

# Daily keyword report for one ad group
searchads keywords report \
  --campaignId 123 --adGroupId 456 \
  --startDate 2026-02-01 --endDate 2026-02-07 \
  --minTaps 5 --json

# Pause negatives by text at campaign level
searchads negatives pause --campaignId 123 --text "free" --text "cheap" --json

# Download a completed custom report
searchads reports download --reportId 987654 --out reports/custom/987654.csv --json
```
