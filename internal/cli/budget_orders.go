package cli

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"searchads-cli/internal/appleads"
)

func RunBudgetOrders(ctx context.Context, client *appleads.Client, args []string, jsonOut bool) {
	if err := ensureCredentialsPresent(); err != nil {
		respondCommandError("budget-orders", jsonOut, err)
		return
	}

	action := actionFromArgs(args, "list")
	switch action {
	case "list":
		runBudgetOrdersList(ctx, client, jsonOut)
	case "get", "show":
		runBudgetOrdersGet(ctx, client, args, jsonOut)
	case "create":
		runBudgetOrdersCreate(ctx, client, args, jsonOut)
	case "update":
		runBudgetOrdersUpdate(ctx, client, args, jsonOut)
	default:
		respondCommandError("budget-orders", jsonOut, fmt.Errorf("Unsupported budget-orders action: %s. Use: list|get|create|update", action))
	}
}

func runBudgetOrdersList(ctx context.Context, client *appleads.Client, jsonOut bool) {
	orders, err := client.FetchBudgetOrders(ctx)
	if err != nil {
		respondCommandError("budget-orders", jsonOut, err)
		return
	}
	sort.Slice(orders, func(i, j int) bool { return orders[i].ID < orders[j].ID })
	respondBudgetOrders(jsonOut, orders)
}

func runBudgetOrdersGet(ctx context.Context, client *appleads.Client, args []string, jsonOut bool) {
	id, err := requiredIntFlag(args, "--budgetOrderId")
	if err != nil {
		respondCommandError("budget-orders", jsonOut, err)
		return
	}
	order, err := client.FetchBudgetOrder(ctx, id)
	if err != nil {
		respondCommandError("budget-orders", jsonOut, err)
		return
	}
	if jsonOut {
		printJSON(order)
		return
	}
	printBudgetOrder(*order)
}

func runBudgetOrdersCreate(ctx context.Context, client *appleads.Client, args []string, jsonOut bool) {
	order, err := parseBudgetOrderInput(args, true)
	if err != nil {
		respondCommandError("budget-orders", jsonOut, err)
		return
	}
	created, err := client.CreateBudgetOrder(ctx, order, sortedIntFlagValues(args, "--orgId"))
	if err != nil {
		respondCommandError("budget-orders", jsonOut, err)
		return
	}
	if jsonOut {
		printJSON(map[string]any{"ok": true, "action": "create", "budgetOrder": created})
		return
	}
	fmt.Printf("ok action=create budgetOrderId=%d status=%s name=%s\n", created.ID, created.Status, created.Name)
}

func runBudgetOrdersUpdate(ctx context.Context, client *appleads.Client, args []string, jsonOut bool) {
	id, err := requiredIntFlag(args, "--budgetOrderId")
	if err != nil {
		respondCommandError("budget-orders", jsonOut, err)
		return
	}
	order, err := parseBudgetOrderInput(args, false)
	if err != nil {
		respondCommandError("budget-orders", jsonOut, err)
		return
	}
	updated, err := client.UpdateBudgetOrder(ctx, id, order)
	if err != nil {
		respondCommandError("budget-orders", jsonOut, err)
		return
	}
	if jsonOut {
		printJSON(map[string]any{"ok": true, "action": "update", "budgetOrder": updated})
		return
	}
	fmt.Printf("ok action=update budgetOrderId=%d status=%s name=%s\n", updated.ID, updated.Status, updated.Name)
}

func parseBudgetOrderInput(args []string, requireBasics bool) (appleads.BudgetOrderSummary, error) {
	name := strings.TrimSpace(valueForFlag(args, "--name"))
	if requireBasics && name == "" {
		return appleads.BudgetOrderSummary{}, fmt.Errorf("Missing required --name <budget order name>")
	}
	order := appleads.BudgetOrderSummary{Name: name}
	order.StartDate = stringPtrFromFlag(args, "--startDate")
	order.EndDate = stringPtrFromFlag(args, "--endDate")
	order.OrderNumber = stringPtrFromFlag(args, "--orderNumber")
	order.ClientName = stringPtrFromFlag(args, "--clientName")
	order.PrimaryBuyerName = stringPtrFromFlag(args, "--primaryBuyerName")
	order.PrimaryBuyerEmail = stringPtrFromFlag(args, "--primaryBuyerEmail")
	order.BillingEmail = stringPtrFromFlag(args, "--billingEmail")

	if raw := strings.TrimSpace(valueForFlag(args, "--budgetAmount")); raw != "" {
		amount := 0.0
		if _, err := fmt.Sscanf(raw, "%f", &amount); err != nil || amount <= 0 {
			return appleads.BudgetOrderSummary{}, fmt.Errorf("Invalid --budgetAmount %q", raw)
		}
		order.Budget = &appleads.MoneyAmount{
			Amount:   amount,
			Currency: strings.ToUpper(firstNonEmptyString(valueForFlag(args, "--budgetCurrency"), "GBP")),
		}
	} else if requireBasics {
		return appleads.BudgetOrderSummary{}, fmt.Errorf("Missing required --budgetAmount <number>")
	}
	return order, nil
}

func stringPtrFromFlag(args []string, flag string) *string {
	value := strings.TrimSpace(valueForFlag(args, flag))
	if value == "" {
		return nil
	}
	return &value
}

func respondBudgetOrders(jsonOut bool, orders []appleads.BudgetOrderSummary) {
	if jsonOut {
		printJSON(orders)
		return
	}
	fmt.Printf("budgetOrderCount=%d\n", len(orders))
	for _, order := range orders {
		printBudgetOrder(order)
	}
}

func printBudgetOrder(order appleads.BudgetOrderSummary) {
	budget := "-"
	if order.Budget != nil {
		budget = fmt.Sprintf("%.4f %s", order.Budget.Amount, order.Budget.Currency)
	}
	fmt.Printf("%d\t%s\t%s\t%s\n", order.ID, order.Status, budget, order.Name)
}
