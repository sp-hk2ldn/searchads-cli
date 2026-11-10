package cli

import (
	"context"
	"fmt"
	"time"

	"searchads-cli/internal/appleads"
)

func RunStatus(ctx context.Context, accountIDs ...string) {
	now := time.Now().UTC().Format(time.RFC3339)
	creds, err := appleads.LoadCredentials()
	if err != nil {
		fmt.Println("searchads status")
		fmt.Printf("time=%s\n", now)
		markCommandFailed()
		fmt.Println("credentials=invalid")
		fmt.Printf("error=%s\n", err.Error())
		return
	}

	if creds == nil || !creds.IsComplete() {
		fmt.Println("searchads status")
		fmt.Printf("time=%s\n", now)
		markCommandFailed()
		fmt.Println("credentials=missing")
		fmt.Println("next=set SEARCHADS_CREDENTIALS_JSON or SEARCHADS_CLIENT_ID/SEARCHADS_TEAM_ID/SEARCHADS_KEY_ID/SEARCHADS_PRIVATE_KEY")
		return
	}

	client := appleads.NewClient(nil)
	if len(accountIDs) > 0 {
		client.SetAdAccountID(accountIDs[0])
	}
	orgID, accountID, err := client.AccountDetails(ctx)
	if err != nil {
		fmt.Println("searchads status")
		fmt.Printf("time=%s\n", now)
		fmt.Println("credentials=present")
		markCommandFailed()
		fmt.Println("auth=failed")
		fmt.Printf("error=%s\n", err.Error())
		return
	}

	fmt.Println("searchads status")
	fmt.Printf("time=%s\n", now)
	fmt.Println("credentials=ok")
	fmt.Printf("orgId=%s\n", orgID)
	fmt.Printf("adAccountId=%s\n", accountID)
	if creds.OrgID != "" {
		fmt.Printf("configuredOrgId=%s\n", creds.OrgID)
	}
}
