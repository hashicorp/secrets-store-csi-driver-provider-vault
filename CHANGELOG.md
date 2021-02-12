## Unreleased

CHANGES:

* All secret engines are now supported [[GH-63](https://github.com/hashicorp/secrets-store-csi-driver-provider-vault/pull/63)]
  * **This makes several breaking changes to the configuration of the SecretProviderClass' `objects` entry**
  * There is no top-level `array` entry under `objects`
  * `objectVersion` is now ignored
  * `objectPath` is renamed to `secretPath`
  * `secretKey`, `secretArgs` and `method` are newly available options
  * `objectName` no longer determines which key is read from the secret's data
  * If `secretKey` is set, that is the key from the secret's data that will be written
  * If `secretKey` is not set, the whole JSON response from Vault will be written
  * `vaultSkipTLSVerify` is no longer required to be set to `"true"` if the `vaultAddress` scheme is not `https`
* The provider will now authenticate to Vault as the requesting pod's service account [[GH-64](https://github.com/hashicorp/secrets-store-csi-driver-provider-vault/pull/64)]
  * **This is likely a breaking change for existing deployments being upgraded**
  * secrets-store-csi-driver-provider-vault service account now requires cluster-wide permission to create service account tokens
  * auth/kubernetes mounts in Vault will now need to bind ACL policies to the requesting pods'
    service accounts instead of the provider's service account.
  * `spec.parameters.kubernetesServiceAccountPath` is now ignored and will log a warning if set
  * Min k8s version
  * API to require enabled

IMPROVEMENTS

* The provider now uses the `hashicorp/vault/api` package to communicate with Vault [[GH-61](https://github.com/hashicorp/secrets-store-csi-driver-provider-vault/pull/61)]
* `--version` flag will now print the version of Go used to build the provider [[GH-62](https://github.com/hashicorp/secrets-store-csi-driver-provider-vault/pull/62)]
* CircleCI linting, tests and integration tests added [[GH-60](https://github.com/hashicorp/secrets-store-csi-driver-provider-vault/pull/60)]

## 0.0.7 (January 20th, 2021)

CHANGES:

* Switch provider to gRPC. [[GH-54](https://github.com/hashicorp/secrets-store-csi-driver-provider-vault/pull/54)]
  * Note this requires at least v0.0.14 of the driver, and the driver should have 'vault' included in `--grpcSupportedProviders`.
