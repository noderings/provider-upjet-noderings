# The NodeRings Terraform provider is compiled into the provider binary
# (terraform-plugin-framework, in-process), so the image needs no Terraform CLI
# and no provider download.
FROM gcr.io/distroless/static-debian12:nonroot

ARG TARGETOS
ARG TARGETARCH

COPY bin/${TARGETOS}_${TARGETARCH}/provider /usr/local/bin/provider

USER 65532:65532
EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/provider"]
