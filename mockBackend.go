package main

import (
	"context"
	"strings"
)

const mockConfigVar = "LAGOON_EXTERNAL_SECRETS_MOCK_BACKEND"
const mockName = "Mock Secrets Storage"

// MockBackend implements the SecretStore interface for CI testing.
type MockBackend struct {
	ctx context.Context
}

// Name returns a meaningful name for the secrets store.
func (s *MockBackend) Name() string {
	return mockName
}

// Secrets takes a map containing build-time variables, returns a map of secret
// key values, if any.
func (s *MockBackend) Secrets(
	buildVars map[string]string) (map[string]string, error) {
	// look through all the build vars to find any with the mock config prefix
	for k := range buildVars {
		if strings.HasPrefix(k, mockConfigVar) {
			return map[string]string{
				"MOCK_SECRET_FOO": "foo",
				"MOCK_SECRET_BAR": "bar",
			}, nil
		}
	}
	return nil, nil
}
