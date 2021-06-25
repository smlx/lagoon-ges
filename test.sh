#!/usr/bin/env sh

# This test script is used for simple local smoke-testing.
#
# Usage:
#
# 	1. go build .
# 	2. ./test.sh
#
# You should see some JSON-formatted data like:
#
# $ ./test.sh
# {"EXTERNAL_SECRET_MOCK_SECRET_BAR":"YmFy","EXTERNAL_SECRET_MOCK_SECRET_FOO":"Zm9v"}


set -eu

# Mock backend

export LAGOON_PROJECT_VARIABLES='[{
	"name": "LAGOON_EXTERNAL_SECRETS_MOCK_BACKEND",
	"value": "does not matter",
	"scope": "build"
}]'

# AWS Secrets Manager backend
#
# export LAGOON_PROJECT_VARIABLES='[{
# 	"name": "LAGOON_EXTERNAL_SECRETS_AWS_SECRETS_MANAGER",
# 	"value": "<ARN>#<API_KEY>#<API_SECRET_KEY>",
# 	"scope": "build"
# }]'

./lagoon-ges
