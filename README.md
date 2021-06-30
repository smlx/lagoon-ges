# Lagoon GES (Get External Secrets)

![Tag and Release](https://github.com/smlx/lagoon-ges/workflows/Tag%20and%20Release/badge.svg)
[![Coverage Status](https://coveralls.io/repos/github/smlx/lagoon-ges/badge.svg?branch=main)](https://coveralls.io/github/smlx/lagoon-ges?branch=main)

**WARNING: this tool is not ready for production use.**

## Overview

`lagoon-ges` is an _experimental_ tool for injecting secrets from an external secret storage service into a [Lagoon](https://github.com/uselagoon/lagoon) deployment.

## Backend support status

* [x] Mock (used in CI testing)
* [x] [AWS Secrets Manager](https://aws.amazon.com/secrets-manager)
* [x] [Google Secret Manager](https://cloud.google.com/secret-manager)
* [ ] [Azure Key Vault](https://azure.microsoft.com/en-au/services/key-vault/)

## How it works

`lagoon-ges` is executed during the Lagoon build-deploy process.
It looks for secret store credentials in the build environment and if it finds any it attempts to connect to the secret storage backend to retrieve the secrets.
These secrets are then injected into the runtime environment of the pod running on the Lagoon platform.

Each backend has an associated build variable prefix set out in the table below.
If a build variable with this prefix is defined, `lagoon-ges` will interpret it as credentials for the given secret storage backend.
Multiple variables for one or more secret storage backends may be defined.

| Secret Storage backend | Lagoon build variable prefix                    | Value format                                    |
| ---                    | ---                                             | ---                                             |
| Mock (testing only)    | `LAGOON_EXTERNAL_SECRETS_MOCK_BACKEND`          | n/a (value is ignored)                          |
| AWS Secrets Manager    | `LAGOON_EXTERNAL_SECRETS_AWS_SECRETS_MANAGER`   | `<ARN>#<API_KEY>#<API_SECRET_KEY>`              |
| Google Secret Manager  | `LAGOON_EXTERNAL_SECRETS_GOOGLE_SECRET_MANAGER` | `<RESOURCE_ID>#<API_KEY_JSON (base64 encoded)>` |

## How to use it

### AWS Secrets Manager

1. Create an API access key for your secret object. **Ensure this is tightly scoped. See below for some suggestions.**
2. Add the appropriate build variable(s) to your Lagoon project/environment. e.g. `LAGOON_EXTERNAL_SECRETS_AWS_SECRETS_MANAGER_0`
3. Deploy environment in Lagoon and see that the secret values are injected into the runtime environment. This can be confirmed e.g. by SSH.

#### Restricting secret storage access

It is critical that the service account whose API key you add to Lagoon is tightly scoped:

* It should only be able to access one Lagoon-specific secret object.
* It should have read-only access to that object.
* The secret object should not store any value which is not required by the Lagoon project.
* The service account should also be IP restricted. See [here](https://aws.amazon.com/premiumsupport/knowledge-center/iam-restrict-calls-ip-addresses/) for instructions. Your Lagoon administrator will be able to provide outbound IP addresses for your cluster.

### Google Secret Manager

1. [Create an API access key](https://cloud.google.com/secret-manager/docs/reference/libraries#setting_up_authentication) for your secret object. **Ensure this is tightly scoped. See below for some suggestions.**
2. Add the appropriate build variable(s) to your Lagoon project/environment. e.g. `LAGOON_EXTERNAL_SECRETS_GOOGLE_SECRET_MANAGER_9`
3. Deploy environment in Lagoon and see that the secret values are injected into the runtime environment. This can be confirmed e.g. by SSH.

#### Restricting secret storage access

It is critical that the service account whose API key you add to Lagoon is tightly scoped:

* It should only be able to access one Lagoon-specific secret object.
* It should have read-only access to that object.
* The secret object should not store any value which is not required by the Lagoon project.
* Unfortunately it currently [doesn't seem possible](https://stackoverflow.com/questions/51535493) to configure IP restrictions on service accounts.
