package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"strings"
)

// SecretStore represents an external secret store from which secrets can be
// obtained for use in a Lagoon build.
type SecretStore interface {
	// Name returns a meaningful name for the secrets store.
	Name() string
	// Secrets returns the secrets available in this secret store as a map of
	// key-value pairs.
	// It may return a nil map and nil error if this secret store is not
	// configured via the associated Lagoon environment variable.
	Secrets() (map[string]string, error)
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

func main() {
	stores := []SecretStore{
		&AWSSecretsManager{},
	}

	var normalKey string
	for _, s := range stores {
		secrets, err := s.Secrets()
		if err != nil {
			log.Fatalf("couldn't retrieve secrets from %s: %v", s.Name(), err)
		}
		// TODO split this out for testability
		for k, v := range secrets {
			normalKey = normalizeShellVar(k)
			fmt.Printf("EXTERNAL_SECRET_%s=%s; export EXTERNAL_SECRET_%s;\n",
				normalKey, base64.StdEncoding.EncodeToString([]byte(v)), normalKey)
		}
	}
}
