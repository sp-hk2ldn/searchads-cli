package cli

import (
	"encoding/json"
	"fmt"
	neturl "net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

var commandFailed bool

func ResetCommandFailure() {
	commandFailed = false
}

func CommandFailed() bool {
	return commandFailed
}

func markCommandFailed() {
	commandFailed = true
}

func hasFlag(args []string, flag string) bool {
	for _, arg := range args {
		if arg == flag {
			return true
		}
	}
	return false
}

func printJSON(payload any) {
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		fmt.Println(`{"ok":false,"error":"json_encode_failed"}`)
		return
	}
	fmt.Println(string(data))
}

func valueForFlag(args []string, flag string) string {
	for idx := 0; idx < len(args)-1; idx++ {
		if args[idx] == flag {
			return args[idx+1]
		}
	}
	return ""
}

func valuesForFlag(args []string, flag string) []string {
	values := make([]string, 0, 4)
	for idx := 0; idx < len(args)-1; idx++ {
		if args[idx] == flag {
			values = append(values, args[idx+1])
			idx++
		}
	}
	return values
}

func requiredIntFlag(args []string, flag string) (int, error) {
	raw := strings.TrimSpace(valueForFlag(args, flag))
	if raw == "" {
		return 0, fmt.Errorf("Missing required %s <id>", flag)
	}
	id, err := strconv.Atoi(raw)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("Invalid %s %q", flag, raw)
	}
	return id, nil
}

func parseDate(value string) (time.Time, error) {
	return time.Parse("2006-01-02", strings.TrimSpace(value))
}

func failText(format string, a ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format+"\n", a...)
}

func actionFromArgs(args []string, fallback string) string {
	if len(args) == 0 {
		return fallback
	}
	first := strings.TrimSpace(args[0])
	if first == "" || strings.HasPrefix(first, "-") {
		return fallback
	}
	return strings.ToLower(first)
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func safeDisplayURL(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return trimmed
	}
	parsed, err := neturl.Parse(trimmed)
	if err != nil || !parsed.IsAbs() {
		return trimmed
	}
	parsed.RawQuery = ""
	parsed.Fragment = ""
	return parsed.String()
}

func mergeMetricValues(dst map[string]any, src map[string]any) map[string]any {
	if len(src) == 0 {
		return dst
	}
	if dst == nil {
		dst = map[string]any{}
	}
	for key, value := range src {
		if key == "localSpend" {
			dst[key] = mergeMoneyMetric(mapFromAnyCLI(dst[key]), mapFromAnyCLI(value))
			continue
		}
		if isAdditiveMetricField(key) {
			dst[key] = intFromAnyCLI(dst[key]) + intFromAnyCLI(value)
		}
	}
	removeDerivedMetricFields(dst)
	return dst
}

func isAdditiveMetricField(key string) bool {
	switch key {
	case "impressions",
		"taps",
		"tapInstalls",
		"viewInstalls",
		"totalInstalls",
		"tapNewDownloads",
		"viewNewDownloads",
		"totalNewDownloads",
		"tapRedownloads",
		"viewRedownloads",
		"totalRedownloads",
		"tapPreOrdersPlaced",
		"viewPreOrdersPlaced",
		"totalPreOrdersPlaced":
		return true
	default:
		return false
	}
}

func removeDerivedMetricFields(metrics map[string]any) {
	for _, key := range []string{
		"avgCPM",
		"avgCPT",
		"tapInstallCPI",
		"totalAvgCPI",
		"tapInstallRate",
		"totalInstallRate",
		"conversionRate",
		"ttr",
	} {
		delete(metrics, key)
	}
}

func mergeMoneyMetric(dst map[string]any, src map[string]any) map[string]any {
	if len(src) == 0 {
		return dst
	}
	if dst == nil {
		dst = map[string]any{}
	}
	dst["amount"] = floatFromAnyCLI(dst["amount"]) + floatFromAnyCLI(src["amount"])
	if currency := strings.ToUpper(strings.TrimSpace(fmt.Sprint(src["currency"]))); currency != "" {
		dst["currency"] = currency
	}
	return dst
}

func mapFromAnyCLI(value any) map[string]any {
	if typed, ok := value.(map[string]any); ok {
		return typed
	}
	return nil
}

func intFromAnyCLI(value any) int {
	switch typed := value.(type) {
	case int:
		return typed
	case int64:
		return int(typed)
	case float64:
		return int(typed)
	case json.Number:
		v, _ := typed.Int64()
		return int(v)
	default:
		return 0
	}
}

func floatFromAnyCLI(value any) float64 {
	switch typed := value.(type) {
	case float64:
		return typed
	case float32:
		return float64(typed)
	case int:
		return float64(typed)
	case int64:
		return float64(typed)
	case json.Number:
		v, _ := typed.Float64()
		return v
	default:
		return 0
	}
}

func sortedIntFlagValues(args []string, flag string) []int {
	set := parseIntFlagSet(args, flag)
	values := make([]int, 0, len(set))
	for value := range set {
		values = append(values, value)
	}
	sort.Ints(values)
	return values
}

func valueOrZero(value *int) int {
	if value == nil {
		return 0
	}
	return *value
}
