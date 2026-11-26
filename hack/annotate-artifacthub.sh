#!/usr/bin/env bash
# Annotate a published multi-arch Crossplane package (OCI index) so Artifact
# Hub can list it as a container image. Index-level annotations take precedence
# over image labels. Usage: annotate-artifacthub.sh <image:tag>
set -euo pipefail

IMAGE="${1:?usage: annotate-artifacthub.sh <image:tag>}"
TAG="${IMAGE##*:}"
VERSION="${TAG#v}"
CREATED="${CREATED:-$(date -u +"%Y-%m-%dT%H:%M:%SZ")}"

README_URL="https://raw.githubusercontent.com/noderings/provider-upjet-noderings/${TAG}/README.md"
DESCRIPTION="Crossplane provider for the NodeRings infrastructure marketplace. Manage virtual machines, DNS, firewall rulesets and access control as Kubernetes resources."
MAINTAINERS='[{"name":"NodeRings Team","email":"partners@noderings.com"}]'
ALT_LOCATION="xpkg.upbound.io/noderings/provider-upjet-noderings:${TAG}"

tmp="$(mktemp)"
annotated="$(mktemp)"
trap 'rm -f "$tmp" "$annotated"' EXIT

oras manifest fetch --pretty=false "$IMAGE" >"$tmp"

media_type="$(jq -r '.mediaType // empty' "$tmp")"
if [[ "$media_type" != "application/vnd.oci.image.index.v1+json" ]]; then
  echo "expected an OCI image index, got: ${media_type:-<missing>}" >&2
  exit 1
fi

jq \
  --arg readme "$README_URL" \
  --arg created "$CREATED" \
  --arg desc "$DESCRIPTION" \
  --arg source "https://github.com/noderings/provider-upjet-noderings" \
  --arg url "https://noderings.com" \
  --arg docs "https://noderings.com/blog/crossplane-vps" \
  --arg title "NodeRings Crossplane provider" \
  --arg vendor "NodeRings" \
  --arg version "$VERSION" \
  --arg license "Apache-2.0" \
  --arg keywords "noderings,crossplane,provider,infrastructure" \
  --arg logo "https://noderings.com/favicon.png" \
  --arg maintainers "$MAINTAINERS" \
  --arg alt "$ALT_LOCATION" \
  '.annotations = {
    "io.artifacthub.package.readme-url": $readme,
    "org.opencontainers.image.created": $created,
    "org.opencontainers.image.description": $desc,
    "org.opencontainers.image.source": $source,
    "org.opencontainers.image.url": $url,
    "org.opencontainers.image.documentation": $docs,
    "org.opencontainers.image.title": $title,
    "org.opencontainers.image.vendor": $vendor,
    "org.opencontainers.image.version": $version,
    "io.artifacthub.package.license": $license,
    "io.artifacthub.package.keywords": $keywords,
    "io.artifacthub.package.logo-url": $logo,
    "io.artifacthub.package.maintainers": $maintainers,
    "io.artifacthub.package.alternative-locations": $alt
  }' "$tmp" >"$annotated"

oras manifest push --media-type application/vnd.oci.image.index.v1+json "$IMAGE" "$annotated"

# Confirm the required Artifact Hub annotations landed on the index.
got="$(oras manifest fetch --pretty=false "$IMAGE")"
for key in \
  io.artifacthub.package.readme-url \
  org.opencontainers.image.created \
  org.opencontainers.image.description
do
  value="$(jq -r --arg k "$key" '.annotations[$k] // empty' <<<"$got")"
  if [[ -z "$value" ]]; then
    echo "missing required index annotation: $key" >&2
    exit 1
  fi
  echo "ok $key=$value"
done
