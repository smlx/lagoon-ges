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

const awsSecretsManagerVar = "LAGOON_EXTERNAL_SECRETS_AWS_SECRETS_MANAGER"
const accessKeyVar = "AWS_ACCESS_KEY_ID"
const secretAccessKeyVar = "AWS_SECRET_ACCESS_KEY"

// ExtractCredentials takes a #-separated string containing ARN, API_KEY and
// API_SECRET_KEY, splits them, and returns those variables.
// It also parses the region out of the ARN and returns that separately.
func ExtractCredentials(creds string) (string, string, string, string, error) {
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

// Encode base64-encodes the given string.
func Encode(i string) (string, error) {
	b, err := json.Marshal(i)
	if err != nil {
		return "", fmt.Errorf(`couldn't marshal "%s": %v`, i, err)
	}
	return string(b), nil
}

// Normalize replaces characters not valid for shell variable names with
// underscores.
func Normalize(i string) string {
	return strings.Map(func(r rune) rune {
		switch {
		case r >= 'A' && r <= 'Z':
			return r
		case r >= 'a' && r <= 'z':
			return r
		case r >= '0' && r <= '9':
			return r
		default:
			return '_'
		}
	}, i)
}

func main() {
	// LAGOON_EXTERNAL_SECRETS_AWS_SECRETS_MANAGER must contain a string:
	// <ARN>#<AWS_ACCESS_KEY_ID>#<AWS_SECRET_ACCESS_KEY>
	creds, ok := os.LookupEnv(awsSecretsManagerVar)
	if !ok {
		log.Fatalf("missing env var %s", awsSecretsManagerVar)
	}
	arn, accessKey, secretAccessKey, region, err := ExtractCredentials(creds)
	if err != nil {
		log.Fatalf("couldn't parse %s value: %v", awsSecretsManagerVar, err)
	}
	// inject credentials into the environment so that the AWS SDK can find them
	if err = os.Setenv(accessKeyVar, accessKey); err != nil {
		log.Fatalf("couldn't set environment var %v", accessKeyVar)
	}
	if err = os.Setenv(secretAccessKeyVar, secretAccessKey); err != nil {
		log.Fatalf("couldn't set environment var %v", accessKeyVar)
	}
	// load the configuration, including credentials
	ctx, cancel := context.WithDeadline(context.Background(),
		time.Now().Add(30*time.Second))
	defer cancel()
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithSharedConfigFiles(nil),
	)
	cfg.Region = region // region is required
	if err != nil {
		log.Fatalf("couldn't get env config: %v", err)
	}
	// get the secret
	client := secretsmanager.NewFromConfig(cfg)
	out, err := client.GetSecretValue(ctx,
		&secretsmanager.GetSecretValueInput{SecretId: &arn})
	if err != nil {
		log.Fatalf("couldn't get secret value: %v", err)
	}
	secrets := map[string]string{}
	if err = json.Unmarshal([]byte(*out.SecretString), &secrets); err != nil {
		log.Fatalf("couldn't unmarshal secret values: %v", err)
	}
	// print some shell code
	for k, v := range secrets {
		cleanVal, err := EscapeQuote(v)
		if err != nil {
			log.Fatalf(`couldn't escape and quote "%v": %v`, cleanVal, err)
		}
		normalKey := Normalize(k)
		fmt.Printf("EXTERNAL_SECRET_%s=%s; export EXTERNAL_SECRET_%s;\n", normalKey,
			cleanVal, normalKey)
	}
}
