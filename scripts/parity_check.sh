#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

GOLDEN_DIR="cmd/searchads/testdata/golden"
TMP_BIN="$(mktemp -t searchads-parity-bin)"
trap 'rm -f "$TMP_BIN"' EXIT

go build -o "$TMP_BIN" ./cmd/searchads

has_jq() {
  command -v jq >/dev/null 2>&1
}

normalize_json_file() {
  local input="$1"
  if has_jq; then
    jq -S . "$input"
  else
    python3 - <<'PY' "$input"
import json
import sys

with open(sys.argv[1], "r", encoding="utf-8") as f:
    print(json.dumps(json.load(f), sort_keys=True, indent=2))
PY
  fi
}

run_without_credentials() {
  env \
    -u SEARCHADS_CREDENTIALS_JSON \
    -u SEARCHADS_CLIENT_ID \
    -u SEARCHADS_TEAM_ID \
    -u SEARCHADS_KEY_ID \
    -u SEARCHADS_PRIVATE_KEY \
    "$TMP_BIN" "$@"
}

compare_golden() {
  local name="$1"
  local golden="$2"
  shift 2
  local -a args=("$@")

  printf '\n== %s ==\n' "$name"
  echo "args: ${args[*]}"

  local stdout stderr actual_norm expected_norm
  stdout="$(mktemp)"
  stderr="$(mktemp)"
  set +e
  run_without_credentials "${args[@]}" >"$stdout" 2>"$stderr"
  local ec=$?
  set -e

  if [[ ! -s "$stdout" ]]; then
    echo "FAIL: command produced no stdout; exit=$ec"
    cat "$stderr"
    rm -f "$stdout" "$stderr"
    return 1
  fi

  actual_norm="$(normalize_json_file "$stdout")"
  expected_norm="$(normalize_json_file "$GOLDEN_DIR/$golden")"

  if [[ "$actual_norm" == "$expected_norm" ]]; then
    echo "PASS"
  else
    echo "FAIL"
    echo "--- expected"
    echo "$expected_norm"
    echo "--- actual"
    echo "$actual_norm"
    echo "--- stderr"
    cat "$stderr"
    rm -f "$stdout" "$stderr"
    return 1
  fi
  rm -f "$stdout" "$stderr"
}

run_live_smoke() {
  local name="$1"
  shift
  local -a args=("$@")

  printf '\n== live: %s ==\n' "$name"
  echo "args: ${args[*]}"

  local stdout stderr
  stdout="$(mktemp)"
  stderr="$(mktemp)"
  set +e
  "$TMP_BIN" "${args[@]}" >"$stdout" 2>"$stderr"
  local ec=$?
  set -e

  if [[ $ec -ne 0 ]]; then
    echo "FAIL: command exited $ec"
    cat "$stderr"
    cat "$stdout"
    rm -f "$stdout" "$stderr"
    return 1
  fi
  normalize_json_file "$stdout" >/dev/null
  echo "PASS"
  rm -f "$stdout" "$stderr"
}

has_credentials=false
if [[ -n "${SEARCHADS_CREDENTIALS_JSON:-}" || (-n "${SEARCHADS_CLIENT_ID:-}" && -n "${SEARCHADS_TEAM_ID:-}" && -n "${SEARCHADS_KEY_ID:-}" && -n "${SEARCHADS_PRIVATE_KEY:-}") ]]; then
  has_credentials=true
fi

echo "Running Go CLI parity checks"
echo "No Swift package detected; using golden JSON and read-only live smoke checks."

compare_golden "campaigns/list" campaigns_list_missing_creds.json campaigns list --json
compare_golden "campaigns/find" campaigns_find_missing_creds.json campaigns find --status ENABLED --json
compare_golden "adgroups/list" adgroups_list_missing_creds.json adgroups list --campaignId 1 --json
compare_golden "ads/list" ads_list_missing_creds.json ads list --campaignId 1 --adGroupId 1 --json
compare_golden "creatives/list" creatives_list_missing_creds.json creatives list --json
compare_golden "product-pages/list" product_pages_list_missing_creds.json product-pages list --adamId 1 --json
compare_golden "apps/search" apps_search_missing_creds.json apps search --query meditation --json
compare_golden "apps/eligibility" apps_eligibility_missing_creds.json apps eligibility --adamId 1 --json
compare_golden "geo/search" geo_search_missing_creds.json geo search --query london --json
compare_golden "ad-rejections/find" ad_rejections_find_missing_creds.json ad-rejections find --json
compare_golden "keywords/list" keywords_list_missing_creds.json keywords list --campaignId 1 --adGroupId 1 --json
compare_golden "keywords/report" keywords_report_missing_creds.json keywords report --campaignId 1 --adGroupId 1 --startDate 2026-02-01 --endDate 2026-02-07 --json
compare_golden "searchterms/report" searchterms_report_missing_creds.json searchterms report --campaignId 1 --startDate 2026-02-01 --endDate 2026-02-07 --json
compare_golden "negatives/list" negatives_list_missing_creds.json negatives list --campaignId 1 --json
compare_golden "negatives/pause" negatives_pause_missing_creds.json negatives pause --campaignId 1 --negativeKeywordId 99 --json
compare_golden "sov-report" sov_report_missing_creds.json sov-report --adamId 123 --json
compare_golden "reports/list" reports_list_missing_creds.json reports list --json
compare_golden "budget-orders/list" budget_orders_list_missing_creds.json budget-orders list --json

echo
echo "All missing-credential golden parity checks passed."

if [[ "$has_credentials" != "true" ]]; then
  echo "No Apple Ads credentials detected; skipping live read-only smoke checks."
  exit 0
fi

echo "Apple Ads credentials detected; running read-only live smoke checks."
run_live_smoke "campaigns/list" campaigns list --json
run_live_smoke "reports/list" reports list --json
run_live_smoke "budget-orders/list" budget-orders list --json

if [[ -n "${SEARCHADS_PARITY_CAMPAIGN_ID:-}" ]]; then
  run_live_smoke "adgroups/list" adgroups list --campaignId "$SEARCHADS_PARITY_CAMPAIGN_ID" --json
  run_live_smoke "negatives/list" negatives list --campaignId "$SEARCHADS_PARITY_CAMPAIGN_ID" --json
fi

if [[ -n "${SEARCHADS_PARITY_CAMPAIGN_ID:-}" && -n "${SEARCHADS_PARITY_ADGROUP_ID:-}" ]]; then
  run_live_smoke "ads/list" ads list --campaignId "$SEARCHADS_PARITY_CAMPAIGN_ID" --adGroupId "$SEARCHADS_PARITY_ADGROUP_ID" --json
  run_live_smoke "keywords/list" keywords list --campaignId "$SEARCHADS_PARITY_CAMPAIGN_ID" --adGroupId "$SEARCHADS_PARITY_ADGROUP_ID" --json
  run_live_smoke "adgroups/report" adgroups report --campaignId "$SEARCHADS_PARITY_CAMPAIGN_ID" --adGroupId "$SEARCHADS_PARITY_ADGROUP_ID" --startDate 2026-02-01 --endDate 2026-02-07 --json
  run_live_smoke "keywords/report" keywords report --campaignId "$SEARCHADS_PARITY_CAMPAIGN_ID" --adGroupId "$SEARCHADS_PARITY_ADGROUP_ID" --startDate 2026-02-01 --endDate 2026-02-07 --json
fi

echo
echo "Live read-only parity checks completed."
