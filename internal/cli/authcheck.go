package cli

import (
	"errors"

	"searchads-cli/internal/appleads"
)

const missingCredsMessage = "Missing Apple Ads credentials in env. Set SEARCHADS_CREDENTIALS_JSON or SEARCHADS_CLIENT_ID/SEARCHADS_TEAM_ID/SEARCHADS_KEY_ID/SEARCHADS_PRIVATE_KEY"

func ensureCredentialsPresent() error {
	creds, err := appleads.LoadCredentials()
	if err != nil {
		return err
	}
	if creds == nil || !creds.IsComplete() {
		return errors.New(missingCredsMessage)
	}
	return nil
}
