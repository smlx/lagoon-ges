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

# Google Secret Manager backend
#
# export LAGOON_PROJECT_VARIABLES='[{
# 	"name": "LAGOON_EXTERNAL_SECRETS_GOOGLE_SECRET_MANAGER_2",
# 	"value": "projects/1234567890/secrets/example_secret_foo/versions/1#ewogICJ0eXBlIjogInNlcnZpY2VfYWNjb3VudCIsCiAgInByb2plY3RfaWQiOiAiZXhhbXBsZS1wcm9qZWN0IiwKICAicHJpdmF0ZV9rZXlfaWQiOiAiZGEzOWEzZWU1ZTZiNGIwZDMyNTViZmVmOTU2MDE4OTBhZmQ4MDcwOSIsCiAgInByaXZhdGVfa2V5IjogIi0tLS0tQkVHSU4gRUMgUFJJVkFURSBLRVktLS0tLVxuTUhjQ0FRRUVJQnYxVDNzR3F4cGNHb1hJbmtyeFVoRy9TUUM0Z3ovbnNkd0lFdHN4Y1QwdG9Bb0dDQ3FHU000OVxuQXdFSG9VUURRZ0FFWGlEZXlOZWh5ZXdIZThDalZmTmVNMzVSSndPYzhjMWNMVXA2WE5sSE1maVVkeGlHRUVKSVxuS1RFYWpQaWNhQ09FQThGYTNBK0gzYmQwc1BWb3dyWUhoZz09XG4tLS0tLUVORCBFQyBQUklWQVRFIEtFWS0tLS0tXG4iLAogICJjbGllbnRfZW1haWwiOiAiZXhhbXBsZS1zZXJ2aWNlLWFjY291bnRAZXhhbXBsZS1wcm9qZWN0LmlhbS5nc2VydmljZWFjY291bnQuY29tIiwKICAiY2xpZW50X2lkIjogIjEyMzQ1Njc4OTAxMjM0NTY3ODkwIiwKICAiYXV0aF91cmkiOiAiaHR0cHM6Ly9hY2NvdW50cy5nb29nbGUuY29tL28vb2F1dGgyL2F1dGgiLAogICJ0b2tlbl91cmkiOiAiaHR0cHM6Ly9vYXV0aDIuZ29vZ2xlYXBpcy5jb20vdG9rZW4iLAogICJhdXRoX3Byb3ZpZGVyX3g1MDlfY2VydF91cmwiOiAiaHR0cHM6Ly93d3cuZ29vZ2xlYXBpcy5jb20vb2F1dGgyL3YxL2NlcnRzIiwKICAiY2xpZW50X3g1MDlfY2VydF91cmwiOiAiaHR0cHM6Ly93d3cuZ29vZ2xlYXBpcy5jb20vcm9ib3QvdjEvbWV0YWRhdGEveDUwOS9leGFtcGxlLXNlcnZpY2UtYWNjb3VudCU0MGV4YW1wbGUtcHJvamVjdC5pYW0uZ3NlcnZpY2VhY2NvdW50LmNvbSIKfQo=",
# 	"scope": "build"
# }]'

./lagoon-ges
