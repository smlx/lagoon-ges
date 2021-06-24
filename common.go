package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

const (
	// LagoonProjectVars defines the build environment variable containing Lagoon
	// project variables.
	LagoonProjectVars = "LAGOON_PROJECT_VARIABLES"
	// LagoonEnvVars defines the build environment variable containing Lagoon
	// environment variables.
	LagoonEnvVars = "LAGOON_ENVIRONMENT_VARIABLES"
)

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

// lagoonVar represents a Lagoon environment variable as found in the build
// variables LAGOON_PROJECT_VARIABLES and LAGOON_ENVIRONMENT_VARIABLES.
type lagoonVar struct {
	Name  string
	Value string
	Scope string
}

// NormalizeShellVar replaces characters not valid for shell variable names with
// underscores.
func NormalizeShellVar(i string) string {
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

// MergeSecrets gets the secrets from all the stores and merges them into a
// single map. If two secret stores return a secret with the same key, the one
// that will end up in the merged map is undefined. So keep them unique.
func MergeSecrets(stores []SecretStore,
	buildVars map[string]string) (map[string]string, error) {
	merged := map[string]string{}
	for _, s := range stores {
		secrets, err := s.Secrets(buildVars)
		if err != nil {
			return nil, fmt.Errorf("couldn't retrieve secrets from %s: %v", s.Name(),
				err)
		}
		for k, v := range secrets {
			merged[fmt.Sprintf("EXTERNAL_SECRET_%s", NormalizeShellVar(k))] =
				base64.StdEncoding.EncodeToString([]byte(v))
		}
	}
	return merged, nil
}

// LagoonBuildVars returns a map of build (and global) scoped lagoon variables
// extracted from the process environment.
func LagoonBuildVars() (map[string]string, error) {
	buildVars := map[string]string{}
	for _, envVar := range []string{LagoonProjectVars, LagoonEnvVars} {
		// read variables from environment
		varsJSON, ok := os.LookupEnv(envVar)
		if !ok {
			// variable is not set in process environment.
			// this is OK because these variables are optional. see:
			// https://github.com/amazeeio/lagoon-kbd/blob/main/
			// 	controllers/lagoonbuild_controller.go
			continue
		}
		// unmarshal JSON
		var vars []lagoonVar
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
