package usage

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"
	"time"

	"go.opencensus.io/metric/metricdata"
	"go.opencensus.io/tag"
	"go.opentelemetry.io/collector/pdata/pcommon"

	sophonzsemconv "github.com/open-telemetry/opentelemetry-collector-contrib/pkg/sophonz/semconv"
)

const (
	TenantKey                 = "tenant"
	ExporterIDKey             = "exporterId"
	DefaultCollectionInterval = 1 * time.Hour
	UsageTableName            = "usage"
)

var (
	TagTenantKey, _     = tag.NewKey(TenantKey)
	TagExporterIdKey, _ = tag.NewKey(ExporterIDKey)
)

// Metric is an in-flight per-tenant usage accumulator.
type Metric struct {
	Size  int64
	Count int64
}

// UsageDB mirrors the ClickHouse usage row layout.
type UsageDB struct {
	CollectorID string    `ch:"collector_id" json:"collectorId"`
	ExporterID  string    `ch:"exporter_id" json:"exporterId"`
	TimeStamp   time.Time `ch:"timestamp" json:"timestamp"`
	Tenant      string    `ch:"tenant" json:"tenant"`
	Data        string    `ch:"data" json:"data"`
}

// Encrypt AES-CFB encrypts text using key and returns the iv-prefixed bytes.
func Encrypt(key, text []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	b := base64.StdEncoding.EncodeToString(text)
	ciphertext := make([]byte, aes.BlockSize+len(b))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}
	cfb := cipher.NewCFBEncrypter(block, iv)
	cfb.XORKeyStream(ciphertext[aes.BlockSize:], []byte(b))
	return ciphertext, nil
}

// AddMetric accumulates count/size for a tenant.
func AddMetric(metrics map[string]Metric, tenant string, count int64, size int64) {
	if entry, ok := metrics[tenant]; ok {
		entry.Count += count
		entry.Size += size
		metrics[tenant] = entry
	} else {
		metrics[tenant] = Metric{
			Size:  size,
			Count: count,
		}
	}
}

// GetTenantNameFromResource extracts the tenant resource attribute, defaulting
// to "default".
//
// The collector-derived sophonz.tenant.id wins over the legacy "tenant"
// attribute: it is resolved from the SDK app key rather than taken from the
// payload, so it is the only one a client cannot forge. The legacy key stays as
// a fallback for pipelines that set it explicitly and for resources that pass
// through with the tenant mode off.
func GetTenantNameFromResource(resource pcommon.Resource) *string {
	if tenant, found := resource.Attributes().Get(sophonzsemconv.TenantID); found && tenant.Str() != "" {
		val := tenant.AsString()
		return &val
	}
	if tenant, found := resource.Attributes().Get("tenant"); found {
		val := tenant.AsString()
		return &val
	}
	def := "default"
	return &def
}

// GetIndexOfLabel returns the index of key within labelKeys, or -1.
func GetIndexOfLabel(labelKeys []metricdata.LabelKey, key string) int {
	for i, label := range labelKeys {
		if label.Key == key {
			return i
		}
	}
	return -1
}
