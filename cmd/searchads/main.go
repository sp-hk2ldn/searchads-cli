package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"searchads-cli/internal/appleads"
	"searchads-cli/internal/cli"
)

func main() {
	args := os.Args
	if len(args) < 2 {
		printHelp()
		os.Exit(0)
	}
	if hasFlag(args, "--help") || hasFlag(args, "-h") {
		printHelp()
		os.Exit(0)
	}

	ctx := context.Background()
	cli.ResetCommandFailure()
	command := strings.ToLower(args[1])
	switch command {
	case "status":
		cli.RunStatus(ctx)
	case "campaigns":
		commandArgs := []string{}
		if len(args) > 2 {
			commandArgs = args[2:]
		}
		c := appleads.NewClient(nil)
		cli.RunCampaigns(ctx, c, commandArgs, hasFlag(args, "--json"))
	case "adgroups":
		commandArgs := []string{}
		if len(args) > 2 {
			commandArgs = args[2:]
		}
		c := appleads.NewClient(nil)
		cli.RunAdGroups(ctx, c, commandArgs, hasFlag(args, "--json"))
	case "ads":
		commandArgs := []string{}
		if len(args) > 2 {
			commandArgs = args[2:]
		}
		c := appleads.NewClient(nil)
		cli.RunAds(ctx, c, commandArgs, hasFlag(args, "--json"))
	case "creatives":
		commandArgs := []string{}
		if len(args) > 2 {
			commandArgs = args[2:]
		}
		c := appleads.NewClient(nil)
		cli.RunCreatives(ctx, c, commandArgs, hasFlag(args, "--json"))
	case "product-pages":
		commandArgs := []string{}
		if len(args) > 2 {
			commandArgs = args[2:]
		}
		c := appleads.NewClient(nil)
		cli.RunProductPages(ctx, c, commandArgs, hasFlag(args, "--json"))
	case "apps":
		commandArgs := []string{}
		if len(args) > 2 {
			commandArgs = args[2:]
		}
		c := appleads.NewClient(nil)
		cli.RunApps(ctx, c, commandArgs, hasFlag(args, "--json"))
	case "geo":
		commandArgs := []string{}
		if len(args) > 2 {
			commandArgs = args[2:]
		}
		c := appleads.NewClient(nil)
		cli.RunGeo(ctx, c, commandArgs, hasFlag(args, "--json"))
	case "ad-rejections":
		commandArgs := []string{}
		if len(args) > 2 {
			commandArgs = args[2:]
		}
		c := appleads.NewClient(nil)
		cli.RunAdRejections(ctx, c, commandArgs, hasFlag(args, "--json"))
	case "keywords":
		commandArgs := []string{}
		if len(args) > 2 {
			commandArgs = args[2:]
		}
		c := appleads.NewClient(nil)
		cli.RunKeywords(ctx, c, commandArgs, hasFlag(args, "--json"))
	case "searchterms":
		commandArgs := []string{}
		if len(args) > 2 {
			commandArgs = args[2:]
		}
		c := appleads.NewClient(nil)
		cli.RunSearchTerms(ctx, c, commandArgs, hasFlag(args, "--json"))
	case "negatives":
		commandArgs := []string{}
		if len(args) > 2 {
			commandArgs = args[2:]
		}
		c := appleads.NewClient(nil)
		cli.RunNegatives(ctx, c, commandArgs, hasFlag(args, "--json"))
	case "sov-report":
		commandArgs := []string{}
		if len(args) > 2 {
			commandArgs = args[2:]
		}
		c := appleads.NewClient(nil)
		cli.RunSovReport(ctx, c, commandArgs)
	case "reports":
		commandArgs := []string{}
		if len(args) > 2 {
			commandArgs = args[2:]
		}
		c := appleads.NewClient(nil)
		cli.RunReports(ctx, c, commandArgs, hasFlag(args, "--json"))
	case "budget-orders":
		commandArgs := []string{}
		if len(args) > 2 {
			commandArgs = args[2:]
		}
		c := appleads.NewClient(nil)
		cli.RunBudgetOrders(ctx, c, commandArgs, hasFlag(args, "--json"))
	case "help", "-h", "--help":
		printHelp()
	default:
		printHelp()
		os.Exit(1)
	}
	if cli.CommandFailed() {
		os.Exit(1)
	}
}

func hasFlag(args []string, flag string) bool {
	for _, arg := range args {
		if arg == flag {
			return true
		}
	}
	return false
}

func printHelp() {
	fmt.Println(`searchads

Commands:
  searchads status
  searchads campaigns [list|find|create|pause|activate|delete|update-budget|set-budget|set-bidding-strategy|report] [flags] [--json]
  searchads adgroups [list|find|create|pause|activate|delete|report] [flags] [--json]
  searchads ads [list|find|get|create|update|pause|activate|delete|report] [flags] [--json]
  searchads creatives [list|find|get|create] [flags] [--json]
  searchads product-pages [list|get|locales|countries|devices] [flags] [--json]
  searchads apps [search|get|localized-details|eligibility] [flags] [--json]
  searchads geo [search|get] [flags] [--json]
  searchads ad-rejections [find|get|assets] [flags] [--json]
  searchads keywords [list|find|report|add|pause|activate|remove|rebid|pause-by-text] --campaignId <id> --adGroupId <id> [flags] [--json]
  searchads searchterms report --campaignId <id> [--adGroupId <id>] --startDate YYYY-MM-DD --endDate YYYY-MM-DD [--minTaps N] [--minSpend X] [--json]
  searchads negatives [list|add|remove|pause|activate] --campaignId <id> [--adGroupId <id>] [--negativeKeywordId <id> ...] [--text <kw> ...] [--matchType EXACT|BROAD] [--json]
  searchads sov-report --adamId <id> [--country GB,US] [--dateRange LAST_4_WEEKS] [--out reports/sov] [--json]
  searchads reports [list|get|download] [--reportId <id>] [--state COMPLETED] [--nameContains text] [--limit N] [--out reports/custom/id.csv] [--json]
  searchads budget-orders [list|get|create|update] [flags] [--json]`)
}
