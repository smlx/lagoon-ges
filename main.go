package main

import (
	"encoding/base64"
	"encoding/json"
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

// mergeSecrets gets the secrets from all the stores and merges them into a
// single map. If two secret stores return a secret with the same key, the one
// that will end up in the merged map is undefined. So keep them unique.
func mergeSecrets(stores []SecretStore) (map[string]string, error) {
	merged := map[string]string{}
	for _, s := range stores {
		secrets, err := s.Secrets()
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

func main() {
	merged, err := mergeSecrets([]SecretStore{
		&AWSSecretsManager{},
	})
	if err != nil {
		log.Fatalf("couldn't merge secrets: %v", err)
	}
	fmt.Println(json.Marshal(merged))
}
