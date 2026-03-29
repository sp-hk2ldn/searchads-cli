package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

var testBinaryPath string

func TestMain(m *testing.M) {
	repoRoot := mustRepoRoot()
	binPath := filepath.Join(os.TempDir(), "searchads-test-bin")
	build := exec.Command("go", "build", "-o", binPath, "./cmd/searchads")
	build.Dir = repoRoot
	build.Env = os.Environ()
	if out, err := build.CombinedOutput(); err != nil {
		_, _ = os.Stderr.WriteString("failed to build searchads test binary: " + err.Error() + "\n" + string(out))
		os.Exit(1)
	}
	testBinaryPath = binPath
	code := m.Run()
	_ = os.Remove(binPath)
	os.Exit(code)
}

func TestGoldenJSONWithoutCredentials(t *testing.T) {
	testCases := []struct {
		name       string
		args       []string
		goldenFile string
	}{
		{
			name:       "campaigns list",
			args:       []string{"campaigns", "list", "--json"},
			goldenFile: "campaigns_list_missing_creds.json",
		},
		{
			name:       "campaigns find",
			args:       []string{"campaigns", "find", "--status", "ENABLED", "--json"},
			goldenFile: "campaigns_find_missing_creds.json",
		},
		{
			name:       "adgroups list",
			args:       []string{"adgroups", "list", "--campaignId", "1", "--json"},
			goldenFile: "adgroups_list_missing_creds.json",
		},
		{
			name:       "ads list",
			args:       []string{"ads", "list", "--campaignId", "1", "--adGroupId", "1", "--json"},
			goldenFile: "ads_list_missing_creds.json",
		},
		{
			name:       "creatives list",
			args:       []string{"creatives", "list", "--json"},
			goldenFile: "creatives_list_missing_creds.json",
		},
		{
			name:       "product-pages list",
			args:       []string{"product-pages", "list", "--adamId", "1", "--json"},
			goldenFile: "product_pages_list_missing_creds.json",
		},
		{
			name:       "apps search",
			args:       []string{"apps", "search", "--query", "meditation", "--json"},
			goldenFile: "apps_search_missing_creds.json",
		},
		{
			name:       "apps eligibility",
			args:       []string{"apps", "eligibility", "--adamId", "1", "--json"},
			goldenFile: "apps_eligibility_missing_creds.json",
		},
		{
			name:       "geo search",
			args:       []string{"geo", "search", "--query", "london", "--json"},
			goldenFile: "geo_search_missing_creds.json",
		},
		{
			name:       "ad-rejections find",
			args:       []string{"ad-rejections", "find", "--json"},
			goldenFile: "ad_rejections_find_missing_creds.json",
		},
		{
			name:       "keywords list",
			args:       []string{"keywords", "list", "--campaignId", "1", "--adGroupId", "1", "--json"},
			goldenFile: "keywords_list_missing_creds.json",
		},
		{
			name:       "keywords report",
			args:       []string{"keywords", "report", "--campaignId", "1", "--adGroupId", "1", "--startDate", "2026-02-01", "--endDate", "2026-02-07", "--json"},
			goldenFile: "keywords_report_missing_creds.json",
		},
		{
			name:       "searchterms report",
			args:       []string{"searchterms", "report", "--campaignId", "1", "--startDate", "2026-02-01", "--endDate", "2026-02-07", "--json"},
			goldenFile: "searchterms_report_missing_creds.json",
		},
		{
			name:       "negatives list",
			args:       []string{"negatives", "list", "--campaignId", "1", "--json"},
			goldenFile: "negatives_list_missing_creds.json",
		},
		{
			name:       "negatives pause",
			args:       []string{"negatives", "pause", "--campaignId", "1", "--negativeKeywordId", "99", "--json"},
			goldenFile: "negatives_pause_missing_creds.json",
		},
		{
			name:       "sov report",
			args:       []string{"sov-report", "--adamId", "123", "--json"},
			goldenFile: "sov_report_missing_creds.json",
		},
		{
			name:       "reports list",
			args:       []string{"reports", "list", "--json"},
			goldenFile: "reports_list_missing_creds.json",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := runCLIForJSON(t, tc.args...)
			expected := loadGoldenJSON(t, tc.goldenFile)
			if !reflect.DeepEqual(actual, expected) {
				actualBytes, _ := json.MarshalIndent(actual, "", "  ")
				expectedBytes, _ := json.MarshalIndent(expected, "", "  ")
				t.Fatalf("json mismatch\nexpected:\n%s\nactual:\n%s", string(expectedBytes), string(actualBytes))
			}
		})
	}
}

func runCLIForJSON(t *testing.T, args ...string) map[string]any {
	t.Helper()
	cmd := exec.Command(testBinaryPath, args...)
	cmd.Env = filteredEnvWithoutAdsCreds(os.Environ())
	out, err := cmd.CombinedOutput()
	trimmed := strings.TrimSpace(string(out))
	var payload map[string]any
	if err := json.Unmarshal([]byte(trimmed), &payload); err != nil {
		t.Fatalf("invalid JSON output: %v\nraw:\n%s", err, trimmed)
	}
	if err != nil {
		if ok, exists := payload["ok"].(bool); exists && !ok {
			return payload
		}
		t.Fatalf("command failed: %v\noutput:\n%s", err, string(out))
	}
	return payload
}

func loadGoldenJSON(t *testing.T, name string) map[string]any {
	t.Helper()
	path := filepath.Join(mustRepoRoot(), "cmd", "searchads", "testdata", "golden", name)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read golden file %s: %v", path, err)
	}
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatalf("invalid golden JSON %s: %v", path, err)
	}
	return payload
}

func filteredEnvWithoutAdsCreds(env []string) []string {
	filtered := make([]string, 0, len(env))
	for _, kv := range env {
		if strings.HasPrefix(kv, "SEARCHADS_") {
			continue
		}
		filtered = append(filtered, kv)
	}
	return filtered
}

func mustRepoRoot() string {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	current := wd
	for {
		if _, err := os.Stat(filepath.Join(current, "go.mod")); err == nil {
			return current
		}
		parent := filepath.Dir(current)
		if parent == current {
			panic("unable to locate repository root from " + wd)
		}
		current = parent
	}
}
