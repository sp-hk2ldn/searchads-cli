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
