package credentials

import "errors"

func connectorStoreError(action, platform string) error {
	message := "The CLI cannot " + action + " the connector registration in the system credential store. Make sure that you can access the store."
	if platform != "windows" {
		message += " You can also set STUFF_STASH_CLI_CONNECTOR_CREDENTIAL_FILE to a private file path before you pair the connector."
	}
	return errors.New(message)
}
