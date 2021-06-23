package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

// LAGOON_EXTERNAL_SECRETS_AWS_SECRETS_MANAGER must contain a string with this
// format:
// <ARN>#<AWS_ACCESS_KEY_ID>#<AWS_SECRET_ACCESS_KEY>
const asmConfigVar = "LAGOON_EXTERNAL_SECRETS_AWS_SECRETS_MANAGER"
const asmName = "AWS Secrets Manager"

// These environment variables are accessed by the AWS Go SDK to obtain
// credentials.
const awsAccessKeyID = "AWS_ACCESS_KEY_ID"
const awsSecretAccessKey = "AWS_SECRET_ACCESS_KEY"

// ASMParseCreds takes a #-separated string containing ARN, API_KEY and
// API_SECRET_KEY, splits them, and returns those variables. It also parses the
// region out of the ARN and returns that separately.
func ASMParseCreds(creds string) (string, string, string, string, error) {
	parts := strings.Split(creds, "#")
	if len(parts) != 3 {
		return "", "", "", "", fmt.Errorf("invalid credential format")
	}
	for _, p := range parts {
		if len(p) == 0 {
			return "", "", "", "", fmt.Errorf("invalid credential format")
		}
	}
	arnParts := strings.Split(parts[0], ":")
	if len(arnParts) != 7 {
		return "", "", "", "", fmt.Errorf("invalid ARN")
	}
	for _, p := range arnParts {
		if len(p) == 0 {
			return "", "", "", "", fmt.Errorf("invalid ARN")
		}
	}
	return parts[0], parts[1], parts[2], arnParts[3], nil
}

// AWSSecretsManager implements the SecretStore interface for AWS Secrets
// Manager.
type AWSSecretsManager struct {
	ctx context.Context
}

// Name returns a meaningful name for the secrets store.
func (s *AWSSecretsManager) Name() string {
	return asmName
}

// Secrets takes a map containing build-time variables, returns a map of secret
// key values, if any.
func (s *AWSSecretsManager) Secrets(
	buildVars map[string]string) (map[string]string, error) {
	secrets := map[string]string{}
	// look through all the build vars to find any with the ASM prefix
	for k, v := range buildVars {
		if strings.HasPrefix(k, asmConfigVar) {
			// extract the secret store credentials
			arn, accessKey, secretAccessKey, region, err := ASMParseCreds(v)
			if err != nil {
				return nil, fmt.Errorf("couldn't parse %s value: %v", asmConfigVar, err)
			}
			// get the secrets using the credentials
			err = s.getSecrets(secrets, arn, accessKey, secretAccessKey, region)
			if err != nil {
				return nil, fmt.Errorf("couldn't get secrets from %s: %v", asmName, err)
			}
		}
	}
	return secrets, nil
}

// getSecrets retrieves secret values from AWS Secrets Manager and adds them
// to the provided secrets map.
func (s *AWSSecretsManager) getSecrets(secrets map[string]string,
	arn, accessKey, secretAccessKey, region string) error {
	// inject credentials into the environment so that the AWS SDK can find them
	if err := os.Setenv(awsAccessKeyID, accessKey); err != nil {
		return fmt.Errorf("couldn't set environment var %v", awsAccessKeyID)
	}
	if err := os.Setenv(awsSecretAccessKey, secretAccessKey); err != nil {
		return fmt.Errorf("couldn't set environment var %v", awsSecretAccessKey)
	}
	// load the AWS SDK configuration, including credentials
	cfg, err := config.LoadDefaultConfig(s.ctx,
		config.WithSharedConfigFiles(nil),
	)
	cfg.Region = region // SDK requires region separately
	if err != nil {
		return fmt.Errorf("couldn't get env config: %v", err)
	}
	// get the secret
	client := secretsmanager.NewFromConfig(cfg)
	out, err := client.GetSecretValue(s.ctx,
		&secretsmanager.GetSecretValueInput{SecretId: &arn})
	if err != nil {
		return fmt.Errorf("couldn't get secret value: %v", err)
	}
	// extract the secret values
	if err = json.Unmarshal([]byte(*out.SecretString), &secrets); err != nil {
		return fmt.Errorf("couldn't unmarshal secret values: %v", err)
	}
	return nil
}
