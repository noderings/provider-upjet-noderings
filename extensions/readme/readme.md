# Provider NodeRings


`provider-upjet-noderings` is a [Crossplane](https://crossplane.io) provider for [NodeRings](https://noderings.com), a cloud marketplace where independent hosting providers sell virtual machines and game servers through one account and one API.

With it, you manage NodeRings from Kubernetes: apply a `VirtualMachine`, `Zone`, `Record` or `Ruleset`, and the provider creates the real resource, keeps it in the state you declared, and deletes it when you delete the object.

It is built with [Upjet](https://github.com/crossplane/upjet) from the [NodeRings Terraform provider](https://github.com/noderings/terraform-provider-noderings), which is compiled into the provider. No Terraform binary runs in your cluster.

## Install

You need Crossplane v2 or newer.

```yaml
apiVersion: pkg.crossplane.io/v1
kind: Provider
metadata:
  name: provider-upjet-noderings
spec:
  package: xpkg.upbound.io/noderings/provider-upjet-noderings:v0.1.0
```

```bash
kubectl apply -f examples/install.yaml
kubectl get providers.pkg.crossplane.io provider-upjet-noderings
```

Or with the Crossplane CLI:

```bash
crossplane xpkg install provider xpkg.upbound.io/noderings/provider-upjet-noderings:v0.1.0
```

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
   apiVersion: noderings.m.crossplane.io/v1beta1
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

   To give one namespace its own credentials, create a namespaced `ProviderConfig` (`noderings.m.crossplane.io/v1beta1`) in that namespace and reference it with `providerConfigRef: {kind: ProviderConfig, name: ...}`.

## Create resources

A DNS zone and a record that points at it. Both are free:

```yaml
apiVersion: dns.noderings.m.crossplane.io/v1alpha1
kind: Zone
metadata:
  name: example
  namespace: default
spec:
  forProvider:
    name: example.com
---
apiVersion: dns.noderings.m.crossplane.io/v1alpha1
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
kubectl get zones.dns.noderings.m.crossplane.io,records.dns.noderings.m.crossplane.io -n default
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

Every resource exists twice: namespaced (`<group>.noderings.m.crossplane.io`, recommended for Crossplane v2) and cluster-scoped (`<group>.noderings.crossplane.io`).

| Group | Kinds |
|---|---|
| `compute` | `VirtualMachine`, `VirtualMachineBackup`, `VirtualMachineSnapshot`, `VirtualMachineAdditionalIP` |
| `dns` | `Zone`, `Record`, `GeoRecord`, `ZoneDNSSEC` |
| `firewall` | `Ruleset`, `RulesetRules`, `RulesetAttachment`, `AddressGroup` |
| `gameserver` | `GameServer`, `Allocation`, `Backup`, `Database`, `Schedule` |
| `iam` | `SSHKey`, `Role`, `RoleAssignment`, `Group`, `GroupMembership`, `Secret`, `EnvironmentVariable` |

The full API reference, with every field documented, is on the [Upbound Marketplace](https://marketplace.upbound.io/providers/noderings/provider-upjet-noderings/) and in [`package/crds`](package/crds).

### Adopt existing resources

Set the `crossplane.io/external-name` annotation to the resource's ID and the provider takes it over instead of creating a new one:

| Kind | External name |
|---|---|
| `Record`, `GeoRecord` | `<zone id>/<name>/<type>`, for example `<zone id>/www/A` |
| `VirtualMachineSnapshot` | `<vm id>/<snapshot id>` |
| `VirtualMachineAdditionalIP` | `<vm id>/<ip address>` |
| `Allocation`, `Backup`, `Database`, `Schedule` | `<game server id>/<id>` |
| `RoleAssignment` | `<user, group or service_account>/<principal id>` |
| everything else | the NodeRings ID |

Use `managementPolicies: ["Observe"]` to read a resource without ever changing it.
