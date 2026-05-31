module github.com/open-telemetry/opentelemetry-collector-contrib/processor/sophonzattributeprocessor

go 1.25.0

require (
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/sophonz v0.0.0-00010101000000-000000000000
	github.com/segmentio/ksuid v1.0.4
	go.opentelemetry.io/collector/client v1.59.1-0.20260528150546-fe2cf23ff222
	go.opentelemetry.io/collector/component v1.59.1-0.20260528150546-fe2cf23ff222
	go.opentelemetry.io/collector/consumer v1.59.1-0.20260528150546-fe2cf23ff222
	go.opentelemetry.io/collector/pdata v1.59.1-0.20260528150546-fe2cf23ff222
	go.opentelemetry.io/collector/processor v1.59.1-0.20260528150546-fe2cf23ff222
	go.opentelemetry.io/collector/processor/processorhelper v0.153.1-0.20260528150546-fe2cf23ff222
	go.opentelemetry.io/otel v1.43.0
	go.uber.org/zap v1.28.0
)

require (
	github.com/ClickHouse/ch-go v0.71.0 // indirect
	github.com/ClickHouse/clickhouse-go/v2 v2.46.0 // indirect
	github.com/andybalholm/brotli v1.2.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/go-faster/city v1.0.1 // indirect
	github.com/go-faster/errors v0.7.1 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/hashicorp/go-version v1.9.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.18.3 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.3-0.20250322232337-35a7c28c31ee // indirect
	github.com/paulmach/orb v0.12.0 // indirect
	github.com/pierrec/lz4/v4 v4.1.25 // indirect
	github.com/segmentio/asm v1.2.1 // indirect
	github.com/shopspring/decimal v1.4.0 // indirect
	go.opentelemetry.io/collector/featuregate v1.59.0 // indirect
	go.opentelemetry.io/collector/internal/componentalias v0.153.0 // indirect
	go.opentelemetry.io/collector/pipeline v1.59.0 // indirect
	go.opentelemetry.io/otel/metric v1.43.0 // indirect
	go.opentelemetry.io/otel/trace v1.43.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/sys v0.45.0 // indirect
)

replace github.com/open-telemetry/opentelemetry-collector-contrib/pkg/sophonz => ../../pkg/sophonz
