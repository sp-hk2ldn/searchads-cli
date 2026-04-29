#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

SWIFT_CMD_DEFAULT=(swift run --package-path ../searchads-cli searchads)
GO_CMD_DEFAULT=(go run ./cmd/searchads)

SWIFT_CMD=(${SWIFT_CMD_OVERRIDE:-${SWIFT_CMD_DEFAULT[*]}})
GO_CMD=(${GO_CMD_OVERRIDE:-${GO_CMD_DEFAULT[*]}})

has_jq() {
  command -v jq >/dev/null 2>&1
}

normalize_json() {
  local input="$1"
  if has_jq; then
    printf '%s' "$input" | jq -S .
  else
    python3 - <<'PY' "$input"
import json,sys
print(json.dumps(json.loads(sys.argv[1]), sort_keys=True, indent=2))
PY
  fi
}

run_json() {
  local tool="$1"; shift
  local -a cmd
  if [[ "$tool" == "swift" ]]; then
    cmd=("${SWIFT_CMD[@]}" "$@")
  else
    cmd=("${GO_CMD[@]}" "$@")
  fi
  "${cmd[@]}"
}

compare_case() {
  local name="$1"; shift
  local -a args=("$@")

  printf '\n== %s ==\n' "$name"
  echo "args: ${args[*]}"

  local swift_out go_out swift_norm go_norm
  local swift_stdout swift_stderr go_stdout go_stderr
  swift_stdout="$(mktemp)"
  swift_stderr="$(mktemp)"
  go_stdout="$(mktemp)"
  go_stderr="$(mktemp)"
  set +e
  run_json swift "${args[@]}" >"$swift_stdout" 2>"$swift_stderr"
  local swift_ec=$?
  run_json go "${args[@]}" >"$go_stdout" 2>"$go_stderr"
  local go_ec=$?
  set -e
  swift_out="$(cat "$swift_stdout")"
  go_out="$(cat "$go_stdout")"

  if [[ $swift_ec -ne 0 ]]; then
    echo "Swift command exited $swift_ec"
    cat "$swift_stderr"
    cat "$swift_stdout"
    rm -f "$swift_stdout" "$swift_stderr" "$go_stdout" "$go_stderr"
    return 1
  fi
  if [[ $go_ec -ne 0 ]]; then
    echo "Go command exited $go_ec"
    cat "$go_stderr"
    cat "$go_stdout"
    rm -f "$swift_stdout" "$swift_stderr" "$go_stdout" "$go_stderr"
    return 1
  fi

  swift_norm="$(normalize_json "$swift_out")"
  go_norm="$(normalize_json "$go_out")"

  if [[ "$swift_norm" == "$go_norm" ]]; then
    echo "PASS"
  else
    echo "FAIL"
    echo "--- Swift"
    echo "$swift_norm"
    echo "--- Go"
    echo "$go_norm"
    rm -f "$swift_stdout" "$swift_stderr" "$go_stdout" "$go_stderr"
    return 1
  fi
  rm -f "$swift_stdout" "$swift_stderr" "$go_stdout" "$go_stderr"
}

missing_creds=true
if [[ -n "${SEARCHADS_CREDENTIALS_JSON:-}" || (-n "${SEARCHADS_CLIENT_ID:-}" && -n "${SEARCHADS_TEAM_ID:-}" && -n "${SEARCHADS_KEY_ID:-}" && -n "${SEARCHADS_PRIVATE_KEY:-}") ]]; then
  missing_creds=false
fi

echo "Running parity checks"

if [[ "$missing_creds" == "true" ]]; then
  echo "Detected missing credentials; running error-path parity checks only."
  compare_case "campaigns/list" campaigns list --json
  compare_case "adgroups/list" adgroups list --campaignId 1 --json
  compare_case "keywords/list" keywords list --campaignId 1 --adGroupId 1 --json
  compare_case "searchterms/report" searchterms report --campaignId 1 --startDate 2026-02-01 --endDate 2026-02-07 --json
  compare_case "negatives/list" negatives list --campaignId 1 --json
  compare_case "sov-report" sov-report --adamId 123 --json
  echo "All missing-credentials parity checks passed."
  exit 0
fi

compare_case "campaigns/list" campaigns list --json

if [[ -n "${SEARCHADS_PARITY_CAMPAIGN_ID:-}" ]]; then
  compare_case "adgroups/list" adgroups list --campaignId "$SEARCHADS_PARITY_CAMPAIGN_ID" --json
  compare_case "negatives/list" negatives list --campaignId "$SEARCHADS_PARITY_CAMPAIGN_ID" --json
fi

if [[ -n "${SEARCHADS_PARITY_CAMPAIGN_ID:-}" && -n "${SEARCHADS_PARITY_ADGROUP_ID:-}" ]]; then
  compare_case "keywords/list" keywords list --campaignId "$SEARCHADS_PARITY_CAMPAIGN_ID" --adGroupId "$SEARCHADS_PARITY_ADGROUP_ID" --json
  compare_case "searchterms/report" searchterms report --campaignId "$SEARCHADS_PARITY_CAMPAIGN_ID" --adGroupId "$SEARCHADS_PARITY_ADGROUP_ID" --startDate 2026-02-01 --endDate 2026-02-07 --json
fi

echo "Base live parity checks completed."
echo "Tip: set SEARCHADS_PARITY_CAMPAIGN_ID and SEARCHADS_PARITY_ADGROUP_ID for deeper checks."
