# SOPHONZ OTel Collector image.
#
# Packages the OCB-generated, statically-linked (CGO_ENABLED=0, -tags timetzdata)
# collector binary into a minimal distroless base.
#
# Prerequisite: the linux binary must already be built into ./_build by
# build-image.sh (or the OCB + cross-compile steps it runs). This Dockerfile
# only packages it — it does not run OCB or the Go toolchain.
#
# Build (from this directory):
#   ./build-image.sh                                      # binary + image, host arch
#   docker build -t ghcr.io/sophonz-labs/opentelemetry-collector . # if the binary is already built
FROM gcr.io/distroless/static-debian12:nonroot

# TARGETARCH is auto-populated by BuildKit per --platform (and defaults to the
# host arch for a plain `docker build`). Do NOT declare it globally (before FROM)
# with a default: a global default shadows the per-platform value and every
# arch ends up with the same (default) binary -> exec format error.
ARG TARGETARCH
COPY _build/sophonz-collector-linux-${TARGETARCH} /sophonz-collector

# OTLP grpc/http, health_check, pprof, zpages
EXPOSE 4317 4318 13133 1777 55679

ENTRYPOINT ["/sophonz-collector"]
CMD ["--config=/etc/otel/config.yaml"]
