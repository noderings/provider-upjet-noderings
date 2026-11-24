# Security policy

## Reporting a vulnerability

Please do not open a public issue for security problems. Email **partners@noderings.com** with:

- a description of the issue and its impact
- steps to reproduce, or a proof of concept
- the provider version (`kubectl get providers.pkg`) and Crossplane version

We acknowledge reports within three working days and keep you updated until the issue is fixed. Credit is given in the release notes unless you ask us not to.

## Supported versions

Security fixes are made for the latest minor release.

## Handling credentials

The provider reads the NodeRings service account token from a Kubernetes Secret referenced by a `ProviderConfig` or `ClusterProviderConfig`. Restrict access to that Secret with RBAC, and give the service account the narrowest NodeRings role that covers the resources you manage.
