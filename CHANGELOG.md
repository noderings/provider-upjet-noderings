# Changelog

## v0.1.0 (unreleased)

First release, generated from the NodeRings Terraform provider.

- 24 managed resources in five groups (`compute`, `dns`, `firewall`, `gameserver`, `iam`), each available namespaced (`*.noderings.m.crossplane.io`) and cluster-scoped (`*.noderings.crossplane.io`).
- Cross-resource references for zone, ruleset, virtual machine, game server, role, group and SSH key IDs.
- The Terraform provider runs in-process (terraform-plugin-framework); the image contains only the provider binary.
- Requests identify themselves as `crossplane-provider-noderings/<version>`.
