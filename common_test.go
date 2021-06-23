package main_test

import (
	"os"
	"reflect"
	"testing"

	main "github.com/smlx/lagoon-ges"
)

func TestNormalizeShellVar(t *testing.T) {
	var testCases = map[string]struct {
		input  string
		expect string
	}{
		"evilChars0": {input: `VAR_,./;'[]\-=`, expect: `VAR___________`},
		"evilChars1": {input: `VAR_<>?:"{}|_+`, expect: `VAR___________`},
		"evilChars2": {input: "!@#$%^&*()`~_VAR", expect: "_____________VAR"},
		"evilChars3": {input: "BAD!@#$%^&*()`~_VAR", expect: "BAD_____________VAR"},
		"niceChars0": {input: "NICE_VAR", expect: "NICE_VAR"},
	}
	for name, tc := range testCases {
		t.Run(name, func(tt *testing.T) {
			result := main.NormalizeShellVar(tc.input)
			if tc.expect != main.NormalizeShellVar(tc.input) {
				tt.Fatalf("expected %s, got %s", tc.expect, result)
			}
		})
	}
}

type MockSecret struct {
	secrets map[string]string
}

func (s *MockSecret) Name() string {
	return "Mock Secret"
}

func (s *MockSecret) Secrets(_ map[string]string) (map[string]string, error) {
	return s.secrets, nil
}

func TestMergeSecrets(t *testing.T) {
	var testCases = map[string]struct {
		input  []main.SecretStore
		expect map[string]string
	}{
		"twoSecrets": {
			input: []main.SecretStore{
				&MockSecret{
					secrets: map[string]string{"FOO": "foo", "BAR": "bar"},
				},
				&MockSecret{
					secrets: map[string]string{"BAZ": "baz", "QUUX": "quux"},
				},
			},
			expect: map[string]string{
				"EXTERNAL_SECRET_FOO":  "Zm9v",
				"EXTERNAL_SECRET_BAR":  "YmFy",
				"EXTERNAL_SECRET_BAZ":  "YmF6",
				"EXTERNAL_SECRET_QUUX": "cXV1eA==",
			},
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(tt *testing.T) {
			var fakeBuildVars map[string]string
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

func TestLagoonBuildVars(t *testing.T) {
	var testCases = map[string]struct {
		input  map[string]string
		expect map[string]string
	}{
		"envVarsOnly": {input: map[string]string{
			main.LagoonProjectVars: `[]`,
			main.LagoonEnvVars: `[
			{"name":"FOO", "value":"foo", "scope":"runtime"},
			{"name":"BAR", "value":"bar", "scope":"build"},
			{"name":"BAZ", "value":"baz", "scope":"global"}]`,
		}, expect: map[string]string{
			"BAR": "bar",
			"BAZ": "baz",
		}},
		"projectVarsOnly": {input: map[string]string{
			main.LagoonProjectVars: `[
			{"name":"FOO", "value":"projectfoo", "scope":"build"},
			{"name":"BAR", "value":"projectbar", "scope":"global"},
			{"name":"BAZ", "value":"projectbaz", "scope":"runtime"}]`,
			main.LagoonEnvVars: `[]`,
		}, expect: map[string]string{
			"FOO": "projectfoo",
			"BAR": "projectbar",
		}},
		"bothVars": {input: map[string]string{
			main.LagoonProjectVars: `[
			{"name":"FOOPROJECT", "value":"projectfoo", "scope":"global"},
			{"name":"BARPROJECT", "value":"projectbar", "scope":"runtime"},
			{"name":"BAZPROJECT", "value":"projectbaz", "scope":"build"}]`,
			main.LagoonEnvVars: `[
			{"name":"FOOENV", "value":"envfoo", "scope":"build"},
			{"name":"BARENV", "value":"envbar", "scope":"runtime"},
			{"name":"BAZENV", "value":"envbaz", "scope":"global"}]`,
		}, expect: map[string]string{
			"FOOPROJECT": "projectfoo",
			"BAZPROJECT": "projectbaz",
			"FOOENV":     "envfoo",
			"BAZENV":     "envbaz",
		}},
		"noVars": {input: map[string]string{
			main.LagoonProjectVars: `[]`,
			main.LagoonEnvVars:     `[]`,
		}, expect: map[string]string{}},
	}
	for name, tc := range testCases {
		t.Run(name, func(tt *testing.T) {
			// set env vars
			for k, v := range tc.input {
				if err := os.Setenv(k, v); err != nil {
					tt.Fatal(err)
				}
			}
			result, err := main.LagoonBuildVars()
			if err != nil {
				tt.Fatal(err)
			}
			if !reflect.DeepEqual(result, tc.expect) {
				tt.Fatalf("expected %v, got %v", tc.expect, result)
			}
		})
	}
}
