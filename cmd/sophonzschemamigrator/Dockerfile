# SOPHONZ ClickHouse schema migrator image.
#
# Packages the statically-linked migrator binary together with the migration
# .sql files. golang-migrate loads them via file:// at a path RELATIVE to the
# working directory (migrationmanager/migrators/<signal>/migrations), so the
# SQL tree must sit under WORKDIR.
#
# Prerequisite: the linux binary must already be built into ./_build
# (CGO_ENABLED=0 -tags timetzdata). See build-image.sh.
#
# Run (against a ClickHouse reachable from the container):
#   docker run --rm \
#     -e CLICKHOUSE_HOST=clickhouse -e CLICKHOUSE_PORT=9000 \
#     -e CLICKHOUSE_USERNAME=sophonz -e CLICKHOUSE_PASSWORD=sophonz \
#     ghcr.io/sophonz-labs/opentelemetry-schema-migrator --cluster-name=cluster --replication=false
ARG TARGETARCH=arm64
FROM gcr.io/distroless/static-debian12:nonroot

ARG TARGETARCH
WORKDIR /app
COPY _build/sophonz-migrator-linux-${TARGETARCH} /app/sophonz-schema-migrator
# Migration SQL must be present at the cwd-relative path the migrator expects.
COPY migrationmanager/migrators /app/migrationmanager/migrators

ENTRYPOINT ["/app/sophonz-schema-migrator"]
