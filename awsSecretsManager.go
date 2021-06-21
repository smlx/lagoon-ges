package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

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
const awsAccessKeyVar = "AWS_ACCESS_KEY_ID"
const awsSecretAccessKeyVar = "AWS_SECRET_ACCESS_KEY"

// AWSSecretsManager implements the SecretStore interface for AWS Secrets
// Manager.
type AWSSecretsManager struct{}

// Name returns a meaningful name for the secrets store.
func (s *AWSSecretsManager) Name() string {
	return asmName
}

// Secrets returns a map of secret key values.
func (s *AWSSecretsManager) Secrets() (map[string]string, error) {
	// find the config variable
	asmConfig, ok := os.LookupEnv(asmConfigVar)
	if !ok {
		// this isn't an error, the feature is just unused
		log.Printf("no env var %s for secret store %s", asmConfigVar, asmName)
		return nil, nil
	}
	// extract the secret store credentials
	arn, accessKey, secretAccessKey, region, err := ASMGetCreds(asmConfig)
	if err != nil {
		return nil, fmt.Errorf("couldn't parse %s value: %v", asmConfigVar, err)
	}
	// inject credentials into the environment so that the AWS SDK can find them
	if err = os.Setenv(awsAccessKeyVar, accessKey); err != nil {
		return nil, fmt.Errorf("couldn't set environment var %v", awsAccessKeyVar)
	}
	if err = os.Setenv(awsSecretAccessKeyVar, secretAccessKey); err != nil {
		return nil, fmt.Errorf("couldn't set environment var %v", awsSecretAccessKeyVar)
	}
	// load the AWS SDK configuration, including credentials
	ctx, cancel := context.WithDeadline(context.Background(),
		time.Now().Add(30*time.Second))
	defer cancel()
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithSharedConfigFiles(nil),
	)
	cfg.Region = region // SDK requires region separately
	if err != nil {
		return nil, fmt.Errorf("couldn't get env config: %v", err)
	}
	// get the secret
	client := secretsmanager.NewFromConfig(cfg)
	out, err := client.GetSecretValue(ctx,
		&secretsmanager.GetSecretValueInput{SecretId: &arn})
	if err != nil {
		return nil, fmt.Errorf("couldn't get secret value: %v", err)
	}
	// extract the secret values
	secrets := map[string]string{}
	if err = json.Unmarshal([]byte(*out.SecretString), &secrets); err != nil {
		return nil, fmt.Errorf("couldn't unmarshal secret values: %v", err)
	}
	return secrets, nil
}

// ASMGetCreds takes a #-separated string containing ARN, API_KEY and
// API_SECRET_KEY, splits them, and returns those variables. It also parses the
// region out of the ARN and returns that separately.
func ASMGetCreds(creds string) (string, string, string, string, error) {
	parts := strings.Split(creds, "#")
	if len(parts) != 3 {
		return "", "", "", "", fmt.Errorf("invalid credential format")
	}
	arnParts := strings.Split(parts[0], ":")
	if len(arnParts) < 7 {
		return "", "", "", "", fmt.Errorf("invalid credential format")
	}
	return parts[0], parts[1], parts[2], arnParts[3], nil
}
