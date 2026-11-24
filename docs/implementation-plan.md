# provider-upjet-noderings: implementation plan

This document is the plan for building the NodeRings Crossplane provider and the checklist for donating it to `crossplane-contrib`, the way DigitalOcean, UpCloud, Cloudflare and Harbor did, and listing it on the Upbound Marketplace.

> **Decision:** the provider will be donated to `crossplane-contrib`. It is developed under `noderings` until the donation is approved, and the first release (v0.1.0) is made from `crossplane-contrib` after the transfer, so users never see a package address or Go module path that changes.

It is based on:

- the official template [`crossplane/upjet-provider-template`](https://github.com/crossplane/upjet-provider-template)
- [`crossplane-contrib/provider-upjet-digitalocean`](https://github.com/crossplane-contrib/provider-upjet-digitalocean), a contrib Upjet provider
- [`vultr/crossplane-provider-vultr`](https://github.com/vultr/crossplane-provider-vultr), a vendor-owned Upjet provider
- [`crossplane-contrib/provider-redpanda`](https://github.com/crossplane-contrib/provider-redpanda) and [`lansweeper-oss/provider-clickhouse`](https://github.com/lansweeper-oss/provider-clickhouse), providers built on terraform-plugin-framework, like ours
- Crossplane [GOVERNANCE.md](https://github.com/crossplane/crossplane/blob/main/GOVERNANCE.md#community-extension-projects), the [xpkg specification](https://github.com/crossplane/crossplane/blob/main/contributing/specifications/xpkg.md), the [Upjet v2 upgrade guide](https://github.com/crossplane/upjet/blob/main/docs/upjet-v2-upgrade.md), and Upbound's [Marketplace publishing docs](https://docs.upbound.io/manuals/marketplace/repositories/publish-packages/)

---

## 1. What we are building

A Crossplane provider is a controller that runs in a Kubernetes cluster. It turns NodeRings resources into Kubernetes objects: a user applies a `VirtualMachine` or `DNSRecord` manifest, and the provider creates, updates and deletes the real resource through the NodeRings API, then keeps correcting drift.

It is generated with [Upjet](https://github.com/crossplane/upjet) from the NodeRings Terraform provider ([`noderings/terraform-provider-noderings`](https://github.com/noderings/terraform-provider-noderings)). The API logic stays in one place, the same way the Pulumi provider works.

| | Decision |
|---|---|
| Repository | `github.com/noderings/provider-upjet-noderings` |
| Package | After donation: `xpkg.crossplane.io/crossplane-contrib/provider-upjet-noderings` (served from `ghcr.io/crossplane-contrib`), mirrored to `xpkg.upbound.io/crossplane-contrib` |
| Upjet | `github.com/crossplane/upjet/v2` v2.5.1 (latest at time of writing) |
| Crossplane | v2 (`spec.crossplane.version: ">=v2.0.0-0"`), cluster-scoped and namespaced resources |
| Terraform provider integration | terraform-plugin-framework, in-process ("no-fork"). No Terraform CLI in the runtime image |
| License | Apache-2.0 (required by CNCF and crossplane-contrib) |
| Commits | DCO sign-off on every commit (`git commit -s`) from the first commit, so donation needs no history rewrite |

## 2. How it works inside

```
 kubectl apply VirtualMachine ──► provider pod (single Go binary)
                                    ├─ generated controller per kind
                                    ├─ Upjet "Terraform Plugin Framework" connector
                                    └─ NodeRings Terraform provider, compiled in
                                         (github.com/noderings/terraform-provider-noderings/shim)
                                              └─ HTTPS ──► NodeRings API
```

- **In-process provider.** The Terraform provider is imported as a Go library through the exported `shim` package, which already exists for Pulumi. Upjet calls it over the Terraform plugin protocol in-process: no `terraform` binary, no workspace on disk, no provider download at runtime. The image can be `gcr.io/distroless/static`, as provider-clickhouse does.
- **Credentials.** A Kubernetes Secret holds JSON. The `ProviderConfig` points at it, and the setup function turns it into the Terraform provider's configuration.
- **User-Agent.** The provider identifies itself as `crossplane-provider-noderings/<version>` through `shim.NewProviderWithUserAgent`, so NodeRings can tell Crossplane traffic apart from Terraform and Pulumi.

### Wiring (from the template and provider-redpanda, verified against Upjet v2.5.1)

`config/provider.go`, used once for cluster-scoped and once for namespaced resources:

```go
pc := ujconfig.NewProvider([]byte(providerSchema), "noderings", modulePath, []byte(providerMetadata),
    ujconfig.WithRootGroup(rootGroup),                       // see section 3
    ujconfig.WithIncludeList(nil),                           // no CLI-mode resources
    ujconfig.WithTerraformPluginFrameworkIncludeList(ExternalNameConfigured()),
    ujconfig.WithTerraformPluginFrameworkProvider(shim.NewProvider(version.Version)),
    ujconfig.WithReferenceInjectors([]ujconfig.ReferenceInjector{reference.NewInjector(modulePath)}),
    ujconfig.WithFeaturesPackage("internal/features"),
    ujconfig.WithDefaultResourceOptions(ExternalNameConfigurations()),
)
```

`internal/clients/noderings.go`, the setup function. The template's `resolveProviderConfig` logic (ProviderConfig, ClusterProviderConfig, usage tracking) is kept unchanged:

```go
creds := map[string]string{}
_ = json.Unmarshal(secretData, &creds)
ps.Configuration = map[string]any{
    "token":           creds["token"],
    "organization_id": creds["organization_id"], // optional
    "endpoint":        creds["endpoint"],        // optional
}
ps.FrameworkProvider = shim.NewProviderWithUserAgent(version.Version,
    "crossplane-provider-noderings/"+version.Version)
```

The keys in `Configuration` must match the Terraform provider's own schema attribute names: `token`, `organization_id`, `endpoint`, `max_retries`, `requests_per_second`, `http_timeout`. Following provider-clickhouse, the configured provider is cached per ProviderConfig, so it is not reconfigured on every reconcile.

## 3. API design

### API groups

| Scope | Root group | Example |
|---|---|---|
| Cluster-scoped | `noderings.crossplane.io` | `compute.noderings.crossplane.io/v1alpha1`, kind `VirtualMachine` |
| Namespaced (Crossplane v2) | `noderings.m.crossplane.io` | `compute.noderings.m.crossplane.io/v1alpha1`, kind `VirtualMachine` |

Crossplane governance requires `*.crossplane.io` for projects that live in crossplane-contrib, and DigitalOcean's donated provider uses the same pattern (`digitalocean.crossplane.io`, `digitalocean.m.crossplane.io`). Because the provider is being donated and nothing is released yet, it uses this group from the start. The API group cannot change after the first release without breaking every user.

### Kinds

All 24 Terraform resources, in five groups. Every API starts at `v1alpha1` and is promoted to `v1beta1` once it has passed e2e for a release cycle ([API promotion guide](https://github.com/crossplane/crossplane/blob/main/contributing/guide-api-promotion.md)).

| Group | Kind | Terraform resource | External name (import ID) |
|---|---|---|---|
| `compute` | `VirtualMachine` | `noderings_virtual_machine` | VM ID |
| `compute` | `VirtualMachineBackup` | `noderings_virtual_machine_backup` | VM ID |
| `compute` | `VirtualMachineSnapshot` | `noderings_virtual_machine_snapshot` | `<vm id>/<snapshot id>` |
| `compute` | `VirtualMachineAdditionalIP` | `noderings_virtual_machine_additional_ip` | `<vm id>/<ip>` |
| `dns` | `Zone` | `noderings_dns_zone` | zone ID |
| `dns` | `Record` | `noderings_dns_record` | `<zone id>/<name>/<type>` |
| `dns` | `GeoRecord` | `noderings_dns_geo_record` | `<zone id>/<name>/<type>` |
| `dns` | `ZoneDNSSEC` | `noderings_dns_zone_dnssec` | zone ID |
| `firewall` | `Ruleset` | `noderings_firewall_ruleset` | ruleset ID |
| `firewall` | `RulesetRules` | `noderings_firewall_ruleset_rules` | ruleset ID |
| `firewall` | `RulesetAttachment` | `noderings_firewall_ruleset_attachment` | VM ID |
| `firewall` | `AddressGroup` | `noderings_firewall_address_group` | group ID |
| `gameserver` | `GameServer` | `noderings_game_server` | server ID |
| `gameserver` | `Allocation` | `noderings_game_server_allocation` | `<server id>/<port>` |
| `gameserver` | `Backup` | `noderings_game_server_backup` | `<server id>/<backup id>` |
| `gameserver` | `Database` | `noderings_game_server_database` | `<server id>/<database id>` |
| `gameserver` | `Schedule` | `noderings_game_server_schedule` | `<server id>/<schedule id>` |
| `iam` | `SSHKey` | `noderings_ssh_key` | key ID |
| `iam` | `Role` | `noderings_role` | role ID |
| `iam` | `RoleAssignment` | `noderings_role_assignment` | `<kind>/<principal id>` |
| `iam` | `Group` | `noderings_group` | group ID |
| `iam` | `GroupMembership` | `noderings_group_membership` | group ID |
| `iam` | `Secret` | `noderings_secret` | secret ID |
| `iam` | `EnvironmentVariable` | `noderings_environment_variable` | variable ID |

Every external name is the Terraform resource's own ID, so all 24 use `config.IdentifierFromProvider`. A unit test fails if a Terraform resource has no external-name entry, so nothing new upstream can be silently left out.

### References between resources

Cross-resource references let users write `zoneIdRef: {name: my-zone}` instead of copying IDs:

| Field | References |
|---|---|
| `Record.zoneId`, `GeoRecord.zoneId`, `ZoneDNSSEC.zoneId` | `dns.Zone` |
| `VirtualMachine.sshKeyIds` | `iam.SSHKey` |
| `RulesetRules.rulesetId`, `RulesetAttachment.rulesetId` | `firewall.Ruleset` |
| `RulesetAttachment.virtualMachineId`, `VirtualMachineBackup/Snapshot/AdditionalIP.virtualMachineId` | `compute.VirtualMachine` |
| `Allocation/Backup/Database/Schedule.gameServerId` | `gameserver.GameServer` |
| `RoleAssignment.roleId`, `RoleAssignment.groupId`, `GroupMembership.groupId` | `iam.Role`, `iam.Group` |
| `Record.records` | (planned) `compute.VirtualMachine` status `ipv4`, which needs a custom extractor |

### Secrets and connection details

Sensitive Terraform attributes become Kubernetes secret references (inputs) or connection details (outputs). Examples: `VirtualMachine.password` (input), `Secret.value` (input), `Database.password` (output, published to the resource's connection secret).

### Known gap: choosing a plan without data sources

Upjet generates **resources only**, not Terraform data sources. Terraform and Pulumi users find a plan and region with the `marketplace_offer` lookup. Crossplane users have to pass `planId` and `regionId` directly.

- **Phase 1:** document how to get them with the `nr` CLI, the API, or the console's order page.
- **Phase 2:** add a small hand-written controller with an observe-only `MarketplaceOffer` kind (plan code + city → plan ID and region ID) that a `VirtualMachine` can reference. This is the same pattern the AWS provider uses for AMI lookups through Compositions.

## 4. Repository layout

The layout follows the template, with the NodeRings names filled in by `hack/prepare.sh`:

```
.github/
  workflows/ci.yml                     lint, generated-code drift, unit tests, local deploy
  workflows/e2e.yaml                   uptest on PR comment /test-examples=...
  workflows/publish-provider-package.yml   build and push xpkg (ghcr + Upbound)
  workflows/tag.yaml                   tag a release
  renovate.json5                       dependency updates
apis/            generated CRD types (cluster/ and namespaced/)
cluster/images/provider-upjet-noderings/Dockerfile   distroless, provider binary only
cluster/test/setup.sh                 uptest: credentials secret + default ProviderConfigs
cmd/generator/main.go                 code generator
cmd/provider/main.go                  provider entrypoint
config/
  provider.go                         wiring (section 2)
  external_name.go                    the external-name table (section 3)
  groups.go                           Terraform resource → group + kind
  cluster/<group>/config.go           references, sensitive fields, late init, per resource
  namespaced/<group>/config.go
  schema.json                         generated from the Terraform provider
  provider-metadata.yaml              generated from the Terraform docs
examples/
  install.yaml
  cluster/providerconfig/  namespaced/providerconfig/   (secret.yaml.tmpl, providerconfig.yaml)
  cluster/<group>/v1alpha1/<kind>.yaml     hand-written, tested by uptest
  namespaced/<group>/v1alpha1/<kind>.yaml
examples-generated/        generated from the Terraform docs
extensions/icons/icon.svg  extensions/readme/readme.md   Marketplace extras
internal/clients/noderings.go          credentials → Terraform provider config
internal/controller/      generated controllers
package/crossplane.yaml   package metadata (section 7)
build/                    git submodule: crossplane/build makelib
LICENSE  CODE_OF_CONDUCT.md  CONTRIBUTING.md  SECURITY.md  OWNERS.md  CODEOWNERS
README.md  CHANGELOG.md  Makefile  go.mod
```

### Makefile settings

```make
PROJECT_NAME ?= provider-upjet-noderings
PROJECT_REPO ?= github.com/noderings/$(PROJECT_NAME)
export TERRAFORM_PROVIDER_SOURCE  ?= noderings/noderings
export TERRAFORM_PROVIDER_REPO    ?= https://github.com/noderings/terraform-provider-noderings
export TERRAFORM_PROVIDER_VERSION ?= 0.1.0
export TERRAFORM_DOCS_PATH        ?= docs/resources
export TERRAFORM_VERSION          ?= 1.5.7     # build-time only, for `terraform providers schema`
REGISTRY_ORGS ?= ghcr.io/noderings xpkg.upbound.io/noderings
XPKG_REG_ORGS ?= xpkg.upbound.io/noderings ghcr.io/noderings
```

The Terraform CLI is needed only at **build** time, to export `config/schema.json`. Until `noderings/noderings` is on the Terraform Registry, the Makefile downloads the provider from its GitHub release and uses a `dev_overrides` CLI config, as provider-clickhouse does, so building does not depend on the Registry.

The Go module version of the Terraform provider (in `go.mod`) and `TERRAFORM_PROVIDER_VERSION` (for schema and docs) must always point at the same release. CI checks this.

## 5. Tests

| Layer | What | Runs | Needs |
|---|---|---|---|
| Unit | Every Terraform resource has an external name and a group; the credentials JSON maps to the right Terraform attributes; config resolution for ProviderConfig and ClusterProviderConfig | every PR (`make test`) | nothing |
| Generated-code drift | `make generate` produces no diff | every PR | nothing |
| CRD compatibility | `crddiff` reports breaking schema changes | every PR | nothing |
| Local deploy | Provider installs in a kind cluster and becomes Healthy | every PR (`make local-deploy`) | nothing |
| End to end | uptest applies each example, waits for `Ready` and `Synced`, checks import, updates, deletes | PR comment `/test-examples=...`, and nightly on main | `UPTEST_CLOUD_CREDENTIALS` secret |

End-to-end examples, one file per kind. Each has `meta.upbound.io/example-id` and an `uptest.upbound.io/timeout` annotation:

| Suite | Examples | Cost |
|---|---|---|
| Free (default) | `dns/Zone` + `Record` + `GeoRecord` + `ZoneDNSSEC`, `iam/SSHKey`, `Role`, `Group`, `GroupMembership`, `RoleAssignment`, `Secret`, `EnvironmentVariable`, `firewall/AddressGroup`, `Ruleset`, `RulesetRules` | free |
| Billed (opt-in) | `compute/VirtualMachine` + `RulesetAttachment` + `VirtualMachineBackup` + `AdditionalIP` + a `Record` pointing at the VM | cents per run |
| Game servers (opt-in) | `gameserver/*` with a small game plan | cents per run |

Zone names and hostnames use random suffixes from the uptest datasource (`${data.suffix}`), so parallel runs never collide. Everything an example creates is deleted at the end, also on failure, which is how uptest works.

`UPTEST_CLOUD_CREDENTIALS` is the same JSON as the ProviderConfig secret: `{"token": "<service account token>"}`. Use a dedicated service account in a test organization with a credit limit.

## 6. Files required for the Marketplace and crossplane-contrib

| File | Why | Content |
|---|---|---|
| `LICENSE` | CNCF policy for contrib (hard requirement) | Apache-2.0 |
| `CODE_OF_CONDUCT.md` | CNCF policy | Points to the CNCF Code of Conduct |
| `OWNERS.md` | Contrib requires maintainers listed here | NodeRings maintainers, with GitHub handles |
| `CODEOWNERS` | Review routing | `* @noderings/crossplane-maintainers` |
| `CONTRIBUTING.md` | Contributor guide, DCO | Build, generate, test, sign-off |
| `SECURITY.md` | Recommended (Harbor has it) | Report to security@noderings.com |
| `README.md` | Rendered on GitHub and the Marketplace | Install, configure, first resource, docs link, report bugs |
| `CHANGELOG.md` | Release notes | Per version |
| `package/crossplane.yaml` | Marketplace metadata | See below |
| `extensions/icons/icon.svg` | Marketplace listing icon | NodeRings mark (a vector version is still needed) |
| `extensions/readme/readme.md` | Marketplace readme tab | Short version of README.md |

### package/crossplane.yaml

```yaml
apiVersion: meta.pkg.crossplane.io/v1
kind: Provider
metadata:
  name: provider-upjet-noderings
  annotations:
    meta.crossplane.io/maintainer: NodeRings <partners@noderings.com>
    meta.crossplane.io/source: github.com/noderings/provider-upjet-noderings
    meta.crossplane.io/license: Apache-2.0
    meta.crossplane.io/description: |
      Crossplane provider for NodeRings: order virtual machines and manage DNS,
      firewalls, game servers and access control on the NodeRings cloud marketplace.
    meta.crossplane.io/readme: |
      See https://github.com/noderings/provider-upjet-noderings for documentation.
    friendly-name.meta.crossplane.io: Provider NodeRings
spec:
  capabilities:
    - SafeStart
  crossplane:
    version: ">=v2.0.0-0"
```

Every generated field carries the Terraform docs as its GoDoc comment. The Marketplace builds its API reference from the CRDs, so this is what users read.

## 7. Publishing

### Upbound Marketplace (community listing)

One-time setup, done by NodeRings:

1. Create an Upbound account and an organization named `noderings` at <https://accounts.upbound.io>.
2. Create the repository: `up repository create provider-upjet-noderings`. It is public by default.
3. Create a **robot** (Organization, then Robots), add it to a team with write access to the repository, and create a robot token. Personal tokens return 401 on push.
4. Store the robot's access ID and token in GitHub as `XPKG_MIRROR_ACCESS_ID` and `XPKG_MIRROR_TOKEN`.
5. After the first push, publish the listing: `up repository update provider-upjet-noderings --publish` (up CLI v0.39.0 or newer). It appears at `https://marketplace.upbound.io/providers/noderings/provider-upjet-noderings/`.

> Older Upbound docs said new repositories need a one-time manual approval, requested in `#upbound` on the Crossplane Slack with the Git repo, the Upbound org and the repository name. The current docs no longer mention it. If the listing stays hidden after `--publish`, ask there.

Each release then runs the **Publish Provider Package** workflow, using `crossplane-contrib/provider-workflows` `publish-provider.yml` with `registry_org: noderings`. It pushes to `ghcr.io/noderings` and mirrors to `xpkg.upbound.io/noderings`. Only semver tags (`v0.1.0`) are listed.

Listings are Official (made by Upbound), Partner (built with Upbound under an agreement) or Community. A self-published provider is Community. Partner is a business conversation with Upbound.

### Donating to crossplane-contrib (later)

This is how DigitalOcean, UpCloud and Harbor got there:

1. Open an issue in [`crossplane/org`](https://github.com/crossplane/org/issues) giving the provider name, its purpose and scope, the maintainer team (GitHub handles), and a statement of long-term maintenance commitment. Precedents: [UpCloud #129](https://github.com/crossplane/org/issues/129), [Harbor #126](https://github.com/crossplane/org/issues/126) (accepted with one maintainer), [Cloudflare #113](https://github.com/crossplane/org/issues/113).
2. The steering committee approves it, and the repository is transferred into `crossplane-contrib`.
3. After that:
   - packages must also be published to `ghcr.io/crossplane-contrib`, which serves `xpkg.crossplane.io`
   - maintainers go in `OWNERS.md` and Crossplane's `maintainers.yaml`
   - the project is added to the community extension projects page
   - CNCF project health guidelines apply

There is no formal minimum number of maintainers. Two or more, ideally from more than one organization, looks stronger.

Because the repo starts with Apache-2.0, DCO sign-offs, OWNERS.md and the contrib workflows, a donation needs no rework.

## 8. Milestones

| # | Milestone | Done when |
|---|---|---|
| 0 | Prerequisites | The Terraform provider has a release tag that contains `shim.NewProviderWithUserAgent` (v0.1.1). Upbound org created |
| 1 | Scaffold | Template applied with NodeRings names; Apache-2.0, CoC, OWNERS, CONTRIBUTING, SECURITY in place; CI green on an empty provider |
| 2 | Generate | All 24 kinds generated for both scopes; external names, groups, references, sensitive fields configured; `make generate` clean; unit tests pass |
| 3 | Runs locally | `make local-deploy` installs in kind; a `Zone` and `Record` reach Ready against the live API |
| 4 | End to end | Free suite green in CI; billed and game-server suites green on demand |
| 5 | Docs | README, examples per kind, credential setup, plan and region lookup documented |
| 6 | Release v0.1.0 | Package on `xpkg.upbound.io/noderings` and `ghcr.io/noderings`; Marketplace listing published |
| 7 | Phase 2 | `MarketplaceOffer` lookup kind; promote stable kinds to `v1beta1` |
| 8 | Optional | Donation issue in `crossplane/org` |

## 9. What NodeRings needs to do (cannot be automated)

| When | Action |
|---|---|
| Before milestone 3 | Create a service account in a test organization with a credit limit, and store its token as the repository secret `UPTEST_CLOUD_CREDENTIALS` in the form `{"token":"..."}` |
| Before milestone 6 | Upbound account, `noderings` org, robot and token (section 7); store `XPKG_MIRROR_ACCESS_ID` and `XPKG_MIRROR_TOKEN` |
| Before milestone 6 | A vector SVG of the NodeRings mark for `extensions/icons/icon.svg` |
| Before milestone 6 | Name the maintainers for `OWNERS.md` (GitHub handles) |
| Milestone 8 | Decide on donation, and commit maintainers to the CNCF health guidelines |

## 10. Open questions

1. ~~API group~~ Decided: `noderings.crossplane.io` / `noderings.m.crossplane.io`, because the provider will be donated.
2. **Game servers in v0.1.0,** or in a later release after their e2e tests are stable?
3. **Maintainers.** Who besides Shaygun (`@shaygun`)? A second maintainer helps a donation.

## 11. Implementation notes

Findings from building and testing the provider:

- **Empty ID on first read.** Upjet reads a resource before it exists, with an empty `id`. The NodeRings Terraform provider used to send that empty ID to the API, which rejected it with HTTP 400, so nothing could be created. Every resource's `Read` now reports a resource without an ID as missing (`framework.GoneIfNoID`, terraform-provider-noderings commit 66d0d55). Terraform never reads without an ID, so Terraform users see no change.
- **Schema and docs from the Go module.** `make generate` builds the Terraform provider from the exact Go module version in `go.mod` and reads its schema through a `dev_overrides` CLI config, then copies `docs/resources` from the same module. The registry is not involved, and the generated APIs cannot drift from the code that runs.
- **Nested objects.** Single nested attributes in the Terraform schema, such as a firewall rule's `source`, become objects in the CRD, not lists.
- **No organization check per reconcile.** The setup function sets `skip_organization_check`, so the provider does not spend an API call per resource per poll validating the token. A bad token still fails on the first real call.
- **Terraform input validation does not run.** Upjet's in-process mode skips the Terraform provider's config validators (regex and length checks). An invalid value reaches the API, which may normalize it (an environment variable key `a-b` is stored as `ab`), and the next reconcile then wants to replace the resource. The examples use valid values; adding CEL validation rules to the generated CRDs for the important patterns is a follow-up.
