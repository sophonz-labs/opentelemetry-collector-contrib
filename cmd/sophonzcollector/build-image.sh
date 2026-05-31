#!/usr/bin/env bash
# Build (and optionally push) the SOPHONZ collector + schema-migrator images.
#
# Images:
#   ghcr.io/sophonz-labs/opentelemetry-collector:<tag>
#   ghcr.io/sophonz-labs/opentelemetry-schema-migrator:<tag>
#
# Usage:
#   ./build-image.sh [arch] [--regen] [--push]
#     arch     linux target arch for LOCAL builds: arm64 (default) or amd64.
#              Ignored when --push is given (push always builds amd64+arm64).
#     --regen  re-run OCB to regenerate the collector distribution sources
#              (_build/) before compiling. Without it, the existing _build is
#              compiled as-is (faster; preserves the grok operator wiring).
#     --push   build linux/amd64 + linux/arm64 and push a multi-arch manifest
#              to ghcr via buildx. Requires: docker login ghcr.io.
#
# Binaries are static (CGO_ENABLED=0) with tz data embedded (-tags timetzdata),
# so the images use a minimal distroless static base.
set -euo pipefail

REGISTRY="ghcr.io/sophonz-labs"
TAG="${TAG:-0.153.0-sophonz}"
COLLECTOR_IMAGE="$REGISTRY/otel-collector"
MIGRATOR_IMAGE="$REGISTRY/otel-schema-migrator"

ARCH="arm64"
REGEN=0
PUSH=0
for a in "$@"; do
  case "$a" in
    amd64|arm64) ARCH="$a" ;;
    --regen) REGEN=1 ;;
    --push) PUSH=1 ;;
    *) echo "unknown arg: $a" >&2; exit 2 ;;
  esac
done

HERE="$(cd "$(dirname "$0")" && pwd)"            # cmd/sophonzcollector
ROOT="$(cd "$HERE/../.." && pwd)"                 # contrib fork root
BUILD="$HERE/_build"
MIGRATOR="$ROOT/cmd/sophonzschemamigrator"
GROK_MOD="github.com/open-telemetry/opentelemetry-collector-contrib/pkg/sophonzgrok"

# 1. (optional) regenerate the OCB distribution sources.
if [[ "$REGEN" == "1" ]]; then
  echo ">> running OCB (regenerating _build/)"
  ( cd "$ROOT/../opentelemetry-collector/cmd/builder" && go run . --config "$HERE/builder-config.yaml" )
fi

# 2. ensure the grok stanza operator is registered (OCB has no operators
#    section, so it is wired in via a blank import). Idempotent.
if ! grep -q "sophonzgrok" "$BUILD/components.go"; then
  echo ">> injecting grok operator registration into _build/"
  perl -0pi -e 's#import \(\n#import (\n\t_ "'"$GROK_MOD"'"\n#' "$BUILD/components.go"
  ( cd "$BUILD" \
    && go mod edit -require="$GROK_MOD@v0.0.1" -replace="$GROK_MOD=$ROOT/pkg/sophonzgrok" \
    && go mod tidy )
fi

# compile_for <arch>: produce static linux binaries for both binaries.
compile_for() {
  local arch="$1"
  echo ">> compiling collector -> _build/sophonz-collector-linux-$arch"
  ( cd "$BUILD" && GOOS=linux GOARCH="$arch" CGO_ENABLED=0 \
      go build -tags timetzdata -ldflags "-s -w" -o "sophonz-collector-linux-$arch" . )
  echo ">> compiling migrator  -> cmd/sophonzschemamigrator/_build/sophonz-migrator-linux-$arch"
  mkdir -p "$MIGRATOR/_build"
  ( cd "$MIGRATOR" && GOOS=linux GOARCH="$arch" CGO_ENABLED=0 \
      go build -tags timetzdata -ldflags "-s -w" -o "_build/sophonz-migrator-linux-$arch" ./cmd )
}

if [[ "$PUSH" == "1" ]]; then
  echo ">> multi-arch build + push (amd64, arm64) to $REGISTRY"
  compile_for amd64
  compile_for arm64
  docker buildx build --platform linux/amd64,linux/arm64 \
    -t "$COLLECTOR_IMAGE:$TAG" -t "$COLLECTOR_IMAGE:latest" --push "$HERE"
  docker buildx build --platform linux/amd64,linux/arm64 \
    -t "$MIGRATOR_IMAGE:$TAG" -t "$MIGRATOR_IMAGE:latest" --push "$MIGRATOR"
  echo ">> pushed:"
  echo "   $COLLECTOR_IMAGE:$TAG (+latest)"
  echo "   $MIGRATOR_IMAGE:$TAG (+latest)"
else
  echo ">> local build: linux/$ARCH"
  compile_for "$ARCH"
  docker build --build-arg TARGETARCH="$ARCH" -t "$COLLECTOR_IMAGE:$TAG" "$HERE"
  docker build --build-arg TARGETARCH="$ARCH" -t "$MIGRATOR_IMAGE:$TAG" "$MIGRATOR"
  echo ">> built (loaded locally):"
  docker images "$COLLECTOR_IMAGE:$TAG" --format '   {{.Repository}}:{{.Tag}}  {{.Size}}'
  docker images "$MIGRATOR_IMAGE:$TAG" --format '   {{.Repository}}:{{.Tag}}  {{.Size}}'
fi
