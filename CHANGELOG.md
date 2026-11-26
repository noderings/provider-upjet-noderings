# Changelog

## [0.2.0](https://github.com/noderings/provider-upjet-noderings/compare/v0.1.0...v0.2.0) (2026-11-26)


### Features

* add Artifact Hub OCI annotations to published images ([#8](https://github.com/noderings/provider-upjet-noderings/issues/8)) ([85608b6](https://github.com/noderings/provider-upjet-noderings/commit/85608b671c93ec619368fd622f3d49e14cb09a74))

## 0.1.0 (2026-11-26)

First release, generated with Upjet from the NodeRings Terraform provider v0.1.1.

### Features

* first release of the NodeRings Crossplane provider ([#1](https://github.com/noderings/provider-upjet-noderings/issues/1)) ([487c230](https://github.com/noderings/provider-upjet-noderings/commit/487c2308bfa21679ff1bb6438044916d214fa3f4))
* 24 managed resources, each namespaced (`<group>.m.noderings.com`) and cluster-scoped (`<group>.noderings.com`), including `VirtualMachine`, `Zone`, `Record`, `Ruleset`, `SSHKey` and `Role`
* cross-resource references (`...Ref` and `...Selector`) for every ID field
* the Terraform provider runs in-process, so no Terraform binary runs in the cluster
* package for `linux/amd64` and `linux/arm64` at `ghcr.io/noderings/provider-upjet-noderings`
