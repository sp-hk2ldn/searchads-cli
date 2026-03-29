package cli

import (
	"context"
	"fmt"
	"time"

	"searchads-cli/internal/appleads"
)

func RunStatus(ctx context.Context) {
	now := time.Now().UTC().Format(time.RFC3339)
	creds, err := appleads.LoadCredentials()
	if err != nil {
		fmt.Println("searchads status")
		fmt.Printf("time=%s\n", now)
		fmt.Println("credentials=invalid")
		fmt.Printf("error=%s\n", err.Error())
		return
	}

	if creds == nil || !creds.IsComplete() {
		fmt.Println("searchads status")
		fmt.Printf("time=%s\n", now)
		fmt.Println("credentials=missing")
		fmt.Println("next=set SEARCHADS_CREDENTIALS_JSON or SEARCHADS_CLIENT_ID/SEARCHADS_TEAM_ID/SEARCHADS_KEY_ID/SEARCHADS_PRIVATE_KEY")
		return
	}

	client := appleads.NewClient(nil)
	orgID, err := client.ValidateCredentials(ctx)
	if err != nil {
		fmt.Println("searchads status")
		fmt.Printf("time=%s\n", now)
		fmt.Println("credentials=present")
		fmt.Println("auth=failed")
		fmt.Printf("error=%s\n", err.Error())
		return
	}

	fmt.Println("searchads status")
	fmt.Printf("time=%s\n", now)
	fmt.Println("credentials=ok")
	fmt.Printf("orgId=%s\n", orgID)
	if creds.OrgID != "" {
		fmt.Printf("configuredOrgId=%s\n", creds.OrgID)
	}
}
