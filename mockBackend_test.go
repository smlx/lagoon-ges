package main_test

import (
	"reflect"
	"testing"

	main "github.com/smlx/lagoon-ges"
)

func TestMockBackend(t *testing.T) {
	var testCases = map[string]struct {
		input  []main.SecretStore
		expect map[string]string
	}{
		"oneMockBackend": {
			input: []main.SecretStore{
				&main.MockBackend{},
			},
			expect: map[string]string{
				"EXTERNAL_SECRET_MOCK_SECRET_FOO": "Zm9v",
				"EXTERNAL_SECRET_MOCK_SECRET_BAR": "YmFy",
			},
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(tt *testing.T) {
			fakeBuildVars := map[string]string{
				"LAGOON_EXTERNAL_SECRETS_MOCK_BACKEND_XYZ": "",
			}
			result, err := main.MergeSecrets(tc.input, fakeBuildVars)
			if err != nil {
				tt.Fatal(err)
			}
			if !reflect.DeepEqual(result, tc.expect) {
				tt.Fatalf("expected %v, got %v", tc.expect, result)
			}
		})
	}
}
