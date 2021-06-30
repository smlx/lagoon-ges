package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"google.golang.org/api/option"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
)

// LAGOON_EXTERNAL_SECRETS_GOOGLE_SECRET_MANAGER must contain the RESOURCE_ID
// and API_KEY_JSON (base64 encoded) in this format (# separated):
// <RESOURCE_ID>#<API_KEY_JSON (base64 encoded)>
const gsmConfigVar = "LAGOON_EXTERNAL_SECRETS_GOOGLE_SECRET_MANAGER"
const gsmName = "Google Secret Manager"

// GSMParseCreds takes a #-separated string containing RESOURCE_ID and
// API_KEY_JSON (base64 encoded), splits them, decodes the API_KEY_JSON and
// returns those variables.
func GSMParseCreds(creds string) (string, []byte, error) {
	parts := strings.Split(creds, "#")
	if len(parts) != 2 {
		return "", nil, fmt.Errorf("invalid credential format")
	}
	for _, p := range parts {
		if len(p) == 0 {
			return "", nil, fmt.Errorf("invalid credential format")
		}
	}
	data, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return "", nil, fmt.Errorf("couldn't decode API_KEY_JSON")
	}

	return parts[0], data, nil
}

// GoogleSecretManager implements the SecretStore interface for Google Secret
// Manager.
type GoogleSecretManager struct {
	ctx context.Context
}

// Name returns a meaningful name for the secrets store.
func (s *GoogleSecretManager) Name() string {
	return gsmName
}

// Secrets takes a map containing build-time variables, returns a map of secret
// key values, if any.
func (s *GoogleSecretManager) Secrets(
	buildVars map[string]string) (map[string]string, error) {
	secrets := map[string]string{}
	// look through all the build vars to find any with the GSM prefix
	for k, v := range buildVars {
		if strings.HasPrefix(k, gsmConfigVar) {
			// extract the secret store credentials
			resourceID, apiKey, err := GSMParseCreds(v)
			if err != nil {
				return nil, fmt.Errorf("couldn't parse %s value: %v", k, err)
			}
			// get the secrets using the credentials
			err = s.getSecret(secrets, resourceID, apiKey)
			if err != nil {
				return nil, fmt.Errorf("couldn't get secret from %s for %s: %v", gsmName, k, err)
			}
		}
	}
	return secrets, nil
}

// getSecret retrieves secret values from Google Secret Manager and adds them
// to the provided secrets map.
func (s *GoogleSecretManager) getSecret(secrets map[string]string,
	resourceID string, apiKey []byte) error {
	c, err := secretmanager.NewClient(s.ctx, option.WithCredentialsJSON(apiKey))
	if err != nil {
		return fmt.Errorf("couldn't construct secretmanager client: %v", err)
	}
	defer c.Close()
	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: resourceID,
	}
	resp, err := c.AccessSecretVersion(s.ctx, req)
	if err != nil {
		return fmt.Errorf("couldn't access secret: %v", err)
	}
	secrets[strings.Split(resp.Name, "/")[3]] = string(resp.Payload.Data)
	return nil
}
