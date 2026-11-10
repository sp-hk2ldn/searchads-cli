package cli

import (
	"testing"

	"searchads-cli/internal/appleads"
)

func TestSelectKeywordMutationTargetPrefersExactMatchType(t *testing.T) {
	t.Parallel()

	keywords := []appleads.KeywordSummary{
		{ID: 1, Text: "life in the uk test", MatchType: "BROAD"},
		{ID: 2, Text: "life in the uk test", MatchType: "EXACT"},
		{ID: 3, Text: "other", MatchType: "EXACT"},
	}

	got := selectKeywordMutationTarget(keywords, "life in the uk test", "EXACT")
	if got == nil {
		t.Fatal("expected keyword target, got nil")
	}
	if got.ID != 2 {
		t.Fatalf("expected exact keyword id 2, got %d", got.ID)
	}
}

func TestSelectKeywordMutationTargetReturnsNilWhenExactMatchMissing(t *testing.T) {
	t.Parallel()

	keywords := []appleads.KeywordSummary{
		{ID: 1, Text: "life in the uk test", MatchType: "BROAD"},
		{ID: 2, Text: "other", MatchType: "EXACT"},
	}

	got := selectKeywordMutationTarget(keywords, "life in the uk test", "EXACT")
	if got != nil {
		t.Fatalf("expected nil when no exact match exists, got %+v", got)
	}
}

func TestKeywordReportPreviewIncludesHighSpendLargeID(t *testing.T) {
	rows := make([]map[string]any, 0, 33)
	for id := 1; id <= 31; id++ {
		rows = append(rows, map[string]any{"keywordId": id, "spend": 0.0, "taps": 0})
	}
	rows = append(rows,
		map[string]any{"keywordId": 999999, "spend": 10.0, "taps": 2},
		map[string]any{"keywordId": 999998, "spend": 10.0, "taps": 1},
	)
	sortKeywordReportRows(rows)
	// Text output renders only these first 30 rows.
	preview := rows[:30]
	if preview[0]["keywordId"] != 999999 {
		t.Fatalf("highest-spend keyword with more taps missing from preview lead: %v", preview)
	}
	if preview[1]["keywordId"] != 999998 {
		t.Fatalf("equal-spend keyword not ordered by taps: %v", preview)
	}
}
