# Provider NodeRings

[![CI](https://github.com/noderings/provider-upjet-noderings/actions/workflows/ci.yml/badge.svg)](https://github.com/noderings/provider-upjet-noderings/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/noderings/provider-upjet-noderings?sort=semver)](https://github.com/noderings/provider-upjet-noderings/releases)
[![Package](https://img.shields.io/badge/xpkg-ghcr.io%2Fnoderings-2496ED.svg?logo=crossplane)](https://github.com/noderings/provider-upjet-noderings/pkgs/container/provider-upjet-noderings)
[![License](https://img.shields.io/badge/license-Apache--2.0-blue)](LICENSE)

`provider-upjet-noderings` is a [Crossplane](https://crossplane.io) provider for [NodeRings](https://noderings.com). NodeRings is an infrastructure marketplace: independent hosting providers sell compute, GPU, storage, network and more, and you buy and manage them through one account and one API.

With it, you manage NodeRings from Kubernetes: apply a `VirtualMachine`, `Zone`, `Record` or `Ruleset`, and the provider creates the real resource, keeps it in the state you declared, and deletes it when you delete the object.

It is built with [Upjet](https://github.com/crossplane/upjet) from the [NodeRings Terraform provider](https://github.com/noderings/terraform-provider-noderings), which is compiled into the provider. No Terraform binary runs in your cluster.

New to it? The step-by-step guide [How to use Crossplane on NodeRings to order a VPS](https://noderings.com/blog/crossplane-vps) walks through a first run.

## Install

You need Crossplane v2 or newer.

<!-- x-release-please-start-version -->
```yaml
apiVersion: pkg.crossplane.io/v1
kind: Provider
metadata:
  name: provider-upjet-noderings
spec:
  package: ghcr.io/noderings/provider-upjet-noderings:v0.2.0
```
<!-- x-release-please-end -->

```bash
kubectl apply -f examples/install.yaml
kubectl get providers.pkg.crossplane.io provider-upjet-noderings
```

Or with the Crossplane CLI:

<!-- x-release-please-start-version -->
```bash
crossplane xpkg install provider ghcr.io/noderings/provider-upjet-noderings:v0.2.0
```
<!-- x-release-please-end -->

Every release is built for `linux/amd64` and `linux/arm64`.

## Configure credentials

1. In the NodeRings console, open **Organization**, then **Access Control**, then **Service accounts**. Create a service account, give it a role, and generate a token.

2. Store the token in a Secret. The `credentials` key holds JSON:

   ```bash
   kubectl -n crossplane-system create secret generic noderings-creds \
     --from-literal=credentials='{"token":"<your token>"}'
   ```

   Optional keys: `organization_id` (act in another organization), `endpoint`, `max_retries`, `requests_per_second`, `http_timeout`.

3. Point a `ClusterProviderConfig` at it. Namespaced resources use this by default:

   ```yaml
   apiVersion: m.noderings.com/v1beta1
   kind: ClusterProviderConfig
   metadata:
     name: default
   spec:
     credentials:
       source: Secret
       secretRef:
         name: noderings-creds
         namespace: crossplane-system
         key: credentials
   ```

   To give one namespace its own credentials, create a namespaced `ProviderConfig` (`m.noderings.com/v1beta1`) in that namespace and reference it with `providerConfigRef: {kind: ProviderConfig, name: ...}`.

## Create resources

A DNS zone and a record that points at it. Both are free:

```yaml
apiVersion: dns.m.noderings.com/v1alpha1
kind: Zone
metadata:
  name: example
  namespace: default
spec:
  forProvider:
    name: example.com
---
apiVersion: dns.m.noderings.com/v1alpha1
kind: Record
metadata:
  name: www
  namespace: default
spec:
  forProvider:
    zoneIdRef:
      name: example
    name: www
    type: A
    ttl: 300
    records:
      - 192.0.2.10
```

```bash
kubectl apply -f zone.yaml
kubectl get zones.dns.m.noderings.com,records.dns.m.noderings.com -n default
kubectl get zone example -n default -o jsonpath='{.status.atProvider.nameservers}'
```

`zoneIdRef` points the record at the `Zone` object by name. Every ID field has a `...Ref` and a `...Selector` form, so you never copy IDs between objects.

More examples, all tested against the live API:

| Example | Creates | Cost |
|---|---|---|
| [`dns/v1alpha1/zone.yaml`](examples/namespaced/dns/v1alpha1/zone.yaml) | Zone, A record, GeoDNS record, DNSSEC | Free |
| [`firewall/v1alpha1/ruleset.yaml`](examples/namespaced/firewall/v1alpha1/ruleset.yaml) | Address group, ruleset, ordered rules | Free |
| [`iam/v1alpha1/role.yaml`](examples/namespaced/iam/v1alpha1/role.yaml) | Custom role, group, role assignment | Free |
| [`iam/v1alpha1/secret.yaml`](examples/namespaced/iam/v1alpha1/secret.yaml) | Organization secret from a Kubernetes Secret, environment variable | Free |
| [`compute/v1alpha1/virtualmachine.yaml`](examples/namespaced/compute/v1alpha1/virtualmachine.yaml) | SSH key, hourly VM, firewall attachment | Billed by the hour |

Every resource also has a generated example in [`examples-generated/`](examples-generated).

### Ordering a virtual machine

A `VirtualMachine` takes a `planId` and a `regionId`, which together pick one offer in the marketplace: a size, sold by one hosting provider, in one city. Find them in the console's order page, with the API's `ListMarketplaceOfferings`, or with the Terraform `noderings_marketplace_offer` data source. Hourly (usage-billed) plans draw from prepaid credit and stop billing when the `VirtualMachine` is deleted.

Changing `hostname`, `distro` or `sshKeyIds` replaces the server (new IP, empty disk) unless `reinstallOnChange: true`, which reinstalls the same server and keeps its addresses. Set `deletionProtection: true` on servers billed monthly, since deleting one cancels the subscription without a refund.

## Resources

Every resource exists twice: namespaced (`<group>.m.noderings.com`, recommended for Crossplane v2) and cluster-scoped (`<group>.noderings.com`).

| Group | Kinds |
|---|---|
| `compute` | `VirtualMachine`, `VirtualMachineBackup`, `VirtualMachineSnapshot`, `VirtualMachineAdditionalIP` |
| `dns` | `Zone`, `Record`, `GeoRecord`, `ZoneDNSSEC` |
| `firewall` | `Ruleset`, `RulesetRules`, `RulesetAttachment`, `AddressGroup` |
| `iam` | `SSHKey`, `Role`, `RoleAssignment`, `Group`, `GroupMembership`, `Secret`, `EnvironmentVariable` |

Every field is documented in the CRDs in [`package/crds`](package/crds), which `kubectl explain` also reads.

### Adopt existing resources

Set the `crossplane.io/external-name` annotation to the resource's ID and the provider takes it over instead of creating a new one:

| Kind | External name |
|---|---|
| `Record`, `GeoRecord` | `<zone id>/<name>/<type>`, for example `<zone id>/www/A` |
| `VirtualMachineSnapshot` | `<vm id>/<snapshot id>` |
| `VirtualMachineAdditionalIP` | `<vm id>/<ip address>` |
| `RoleAssignment` | `<user, group or service_account>/<principal id>` |
| everything else | the NodeRings ID |

Use `managementPolicies: ["Observe"]` to read a resource without ever changing it.

## Developing

You need Go (see `go.mod`), Docker, and `goimports` on the `PATH`.

```bash
git clone --recurse-submodules https://github.com/noderings/provider-upjet-noderings.git
cd provider-upjet-noderings
make generate        # regenerate APIs, CRDs and controllers from the Terraform provider
make test            # unit tests
make reviewable      # generate, lint and test, as CI does
make local-deploy    # build the package and run it in a local kind cluster
```

To move to a new Terraform provider release:

```bash
go get github.com/noderings/terraform-provider-noderings@vX.Y.Z
make generate
```

The schema and docs used for generation come from that same Go module, so the generated APIs always match the code that runs.

### End-to-end tests

The examples above are tested with [uptest](https://github.com/crossplane/uptest) against the real NodeRings API. Each run creates resources with random names and deletes them afterwards.

```bash
export UPTEST_CLOUD_CREDENTIALS='{"token":"<test service account token>"}'
export UPTEST_DATASOURCE_PATH=.work/uptest-datasource.yaml   # planId, regionId, sshPublicKey for the VM example
export UPTEST_EXAMPLE_LIST=examples/namespaced/dns/v1alpha1/zone.yaml,examples/namespaced/firewall/v1alpha1/ruleset.yaml
make e2e
```

On a pull request, a maintainer comments `/test-examples="examples/namespaced/dns/v1alpha1/zone.yaml"` to run the same tests in CI.

## Report a bug

Open an [issue](https://github.com/noderings/provider-upjet-noderings/issues). For security issues, see [SECURITY.md](SECURITY.md).

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). Every commit needs a [DCO](https://developercertificate.org/) sign-off (`git commit -s`).

## License

Apache License 2.0. See [LICENSE](LICENSE).
