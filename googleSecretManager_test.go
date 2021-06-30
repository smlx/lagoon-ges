package main_test

import (
	"reflect"
	"testing"

	main "github.com/smlx/lagoon-ges"
)

func TestGSMParseCreds(t *testing.T) {
	var testCases = map[string]struct {
		input            string
		expectResourceID string
		expectAPIKey     []byte
		expectErr        bool
	}{
		"badVar0": {
			input:            "",
			expectResourceID: "",
			expectAPIKey:     nil,
			expectErr:        true},
		"badVar1": {
			input:            "#e30=",
			expectResourceID: "",
			expectAPIKey:     nil,
			expectErr:        true},
		"badVar2": {
			input:            "foo#",
			expectResourceID: "",
			expectAPIKey:     nil,
			expectErr:        true},
		"goodVar": {
			input:            "foo#e30=",
			expectResourceID: "foo",
			expectAPIKey:     []byte("{}"),
			expectErr:        false},
	}
	for name, tc := range testCases {
		t.Run(name, func(tt *testing.T) {
			resourceID, apiKey, err := main.GSMParseCreds(tc.input)
			if tc.expectErr {
				if err == nil {
					tt.Fatalf("no error")
				}
			} else {
				if err != nil {
					tt.Fatal(err)
				}
				if !reflect.DeepEqual(resourceID, tc.expectResourceID) {
					tt.Fatalf("expected resourceID %v, got %v", tc.expectResourceID,
						resourceID)
				}
				if !reflect.DeepEqual(apiKey, tc.expectAPIKey) {
					tt.Fatalf("expected apiKey %v, got %v", tc.expectAPIKey, apiKey)
				}
			}
		})
	}
}
