package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"
)

const (
	// define the build environment variables containing Lagoon project and
	// environment variables
	lagoonProjectVars = "LAGOON_PROJECT_VARIABLES"
	lagoonEnvVars     = "LAGOON_ENVIRONMENT_VARIABLES"
	// exit code on interrupt
	exitCodeInterrupt = 2
)

// default global timeout value
var timeout = flag.Uint("timeout", 120, "timeout in seconds")

// SecretStore represents an external secret store from which secrets can be
// obtained for use in a Lagoon build.
type SecretStore interface {
	// Name returns a meaningful name for the secrets store.
	Name() string
	// Secrets takes a map of build-time environment variables as an argument,
	// and returns the secrets available in this secret store as a map of
	// key-value pairs.
	// It may return a nil map and nil error if this secret store is not
	// configured via the associated Lagoon environment variable.
	Secrets(map[string]string) (map[string]string, error)
}

// LagoonVar represents a Lagoon environment variable as found in the build
// variables LAGOON_PROJECT_VARIABLES and LAGOON_ENVIRONMENT_VARIABLES.
type LagoonVar struct {
	Name  string
	Value string
	Scope string
}

// normalizeShellVar replaces characters not valid for shell variable names with
// underscores.
func normalizeShellVar(i string) string {
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

// mergeSecrets gets the secrets from all the stores and merges them into a
// single map. If two secret stores return a secret with the same key, the one
// that will end up in the merged map is undefined. So keep them unique.
func mergeSecrets(stores []SecretStore, buildVars map[string]string) (map[string]string, error) {
	merged := map[string]string{}
	for _, s := range stores {
		secrets, err := s.Secrets(buildVars)
		if err != nil {
			return nil, fmt.Errorf("couldn't retrieve secrets from %s: %v", s.Name(),
				err)
		}
		for k, v := range secrets {
			merged[normalizeShellVar(k)] =
				base64.StdEncoding.EncodeToString([]byte(v))
		}
	}
	return merged, nil
}

// lagoonBuildVars returns a map of build (and global) scoped lagoon variables
// extracted from the process environment.
func lagoonBuildVars() (map[string]string, error) {
	buildVars := map[string]string{}
	for _, envVar := range []string{lagoonProjectVars, lagoonEnvVars} {
		// read variables from environment
		varsJSON, ok := os.LookupEnv(envVar)
		if !ok {
			return nil, fmt.Errorf("couldn't find %s in environment", envVar)
		}
		// unmarshal JSON
		var vars []LagoonVar
		if err := json.Unmarshal([]byte(varsJSON), &vars); err != nil {
			return nil, fmt.Errorf("couldn't unmarshal %s variables: %v", envVar, err)
		}
		// extract and store the build scoped variables
		for _, v := range vars {
			if v.Scope == "build" || v.Scope == "global" {
				buildVars[v.Name] = v.Value
			}
		}
	}
	return buildVars, nil
}

func main() {
	flag.Parse()
	// handle ^C interrupt signals gracefully
	ctx, cancel := context.WithDeadline(context.Background(),
		time.Now().Add(time.Duration(*timeout)*time.Second))
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	defer func() {
		signal.Stop(signalChan)
		cancel()
	}()
	go func() {
		select {
		case <-signalChan:
			log.Println("exiting. ^C again to force.")
			cancel()
		case <-ctx.Done():
		}
		<-signalChan
		os.Exit(exitCodeInterrupt)
	}()
	// get Lagoon project/environment build variables from process environment
	buildVars, err := lagoonBuildVars()
	if err != nil {
		log.Fatalf("couldn't get Lagoon build variables: %v", err)
	}
	// get variables from any configured secret stores
	merged, err := mergeSecrets([]SecretStore{
		&AWSSecretsManager{ctx: ctx},
	}, buildVars)
	if err != nil {
		log.Fatalf("couldn't merge secrets: %v", err)
	}
	// print JSON-formatted variables to stdout
	fmt.Println(json.Marshal(merged))
}
