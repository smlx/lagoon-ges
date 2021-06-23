package main_test

import (
	"reflect"
	"testing"

	main "github.com/smlx/lagoon-ges"
)

func TestASMParseCreds(t *testing.T) {
	var testCases = map[string]struct {
		input     string
		expect    []string
		expectErr bool
	}{
		"badVar0": {input: "", expect: nil, expectErr: true},
		"badVar1": {
			input:     "arn:aws::secretsmanager:eu-central-1:123456789:secret:LagoonTestSecret-XXXXXX#APIKEY#APISECRETKEY",
			expect:    nil,
			expectErr: true},
		"badVar2": {
			input:     "arn:aws:uhoh:secretsmanager:eu-central-1:123456789:secret:LagoonTestSecret-XXXXXX#APIKEY#APISECRETKEY",
			expect:    nil,
			expectErr: true},
		"badVar3": {
			input:     "arn:aws:secretsmanager:eu-central-1:123456789:secret:LagoonTestSecret-XXXXXX#APIKEY#APISECRETKEY#",
			expect:    nil,
			expectErr: true},
		"badVar4": {
			input:     "arn:aws:secretsmanager:eu-central-1:123456789:secret:LagoonTestSecret-XXXXXX#APIKEY#",
			expect:    nil,
			expectErr: true},
		"badVar5": {
			input:     "arn:aws:secretsmanager:eu-central-1:123456789:secret:LagoonTestSecret-XXXXXX##",
			expect:    nil,
			expectErr: true},
		"badVar6": {
			input:     "arn:aws:secretsmanager:eu-central-1:123456789:secret:LagoonTestSecret-XXXXXX#APIKEY",
			expect:    nil,
			expectErr: true},
		"goodVar0": {
			input: "arn:aws:secretsmanager:eu-central-1:123456789:secret:LagoonTestSecret-XXXXXX#APIKEY#APISECRETKEY",
			expect: []string{
				"arn:aws:secretsmanager:eu-central-1:123456789:secret:LagoonTestSecret-XXXXXX",
				"APIKEY",
				"APISECRETKEY",
				"eu-central-1",
			},
		},
		"goodVar1": {
			input: "arn:aws:secretsmanager:eu-central-2:123456789:secret:LagoonTestSecret-YYYYYY#TESTAPIKEY#TESTAPISECRETKEY",
			expect: []string{
				"arn:aws:secretsmanager:eu-central-2:123456789:secret:LagoonTestSecret-YYYYYY",
				"TESTAPIKEY",
				"TESTAPISECRETKEY",
				"eu-central-2",
			},
			expectErr: false},
	}
	for name, tc := range testCases {
		t.Run(name, func(tt *testing.T) {
			arn, accessKey, secretAccessKey, region, err :=
				main.ASMParseCreds(tc.input)
			if tc.expectErr {
				if err == nil {
					tt.Fatalf("no error")
				}
			} else {
				if err != nil {
					tt.Fatal(err)
				}
				result := []string{arn, accessKey, secretAccessKey, region}
				if !reflect.DeepEqual(result, tc.expect) {
					tt.Fatalf("expected %v, got %v", tc.expect, result)
				}
			}
		})
	}
}
