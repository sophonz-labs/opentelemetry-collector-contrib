# SOPHONZ deploy (k8s + CI)

Builds and deploys the SOPHONZ OTel collector and ClickHouse schema-migrator.

- Workflow: [`.github/workflows/sophonz-deploy.yml`](../workflows/sophonz-deploy.yml)
- Kustomize base: [`base/`](base) — collector Deployment + Service + ConfigMap, migrator Job, `v3` Namespace
- Kustomize overlay: [`overlays/v3/`](overlays/v3) — pins everything to namespace `v3`
- Kubeconfig generator: [`scripts/sophonz-gen-kubeconfig.sh`](../../scripts/sophonz-gen-kubeconfig.sh)

## Pipeline

1. Runs on the org ARC runner set `arc-runner-x64` (x86_64).
2. Cross-compiles both binaries for `linux/amd64` + `linux/arm64` (`CGO_ENABLED=0`).
3. `docker buildx` assembles a multi-arch manifest for each image and pushes to GHCR.
   The Dockerfiles only `COPY` the prebuilt per-arch binary (no `RUN`), so no QEMU is needed.
4. Deploy job applies the `v3` overlay to context `oci@afextwin`, pinned to the
   immutable `sha-<commit>` image tag.

Images:

- `ghcr.io/sophonz-labs/otel-collector`
- `ghcr.io/sophonz-labs/otel-schema-migrator`

Tags pushed per build: `<version>-sophonz` (version from `versions.yaml`), `sha-<short-sha>`, `latest`.

## One-time setup

### 1. KUBE_CONFIG secret (v3 namespace admin)

Run against an admin context for the cluster, then store the output as the secret:

```bash
./scripts/sophonz-gen-kubeconfig.sh oci@afextwin > kubeconfig-v3.yaml
gh secret set KUBE_CONFIG < kubeconfig-v3.yaml   # repo or org scope
```

The script creates `ServiceAccount v3/v3-deployer`, binds the built-in `admin`
ClusterRole to it via a namespaced RoleBinding (admin rights limited to `v3`),
mints a long-lived token, and emits a kubeconfig whose context is `oci@afextwin`.

### 2. ClickHouse credentials secret (in-cluster)

```bash
kubectl -n v3 create secret generic sophonz-clickhouse \
  --from-literal=CLICKHOUSE_HOST=clickhouse \
  --from-literal=CLICKHOUSE_PORT=9000 \
  --from-literal=CLICKHOUSE_USERNAME=sophonz \
  --from-literal=CLICKHOUSE_PASSWORD=sophonz
```

Both the collector and the migrator load these via `envFrom`.

## Run

- Automatic: push to `main` touching the sophonz `cmd/`, `exporter/`, `processor/`,
  `pkg/` paths or these CI/k8s files.
- Manual: Actions → "SOPHONZ build & deploy" → Run workflow (uncheck *deploy* to
  push images only).

## Local checks

```bash
kubectl kustomize .github/k8s/overlays/v3      # render manifests
cmd/sophonzcollector/build-image.sh --push     # local multi-arch build+push
```

## Notes

- The migrator runs as a `Job` (immutable spec); the deploy step deletes the old
  Job before re-applying so each deploy re-runs migrations.
- Edit [`base/collector-config.yaml`](base/collector-config.yaml) to wire the real
  ClickHouse exporters — the committed config is a minimal OTLP→debug starter.
- Compiled binaries under `cmd/*/_build/` are git-ignored; CI recompiles from the
  committed OCB sources, so never commit the binaries.
