// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package sophonzclickhouselogsexporter

import (
	"database/sql/driver"
	"encoding/json"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.uber.org/zap"
)

// ClickHouse accepts maps as bound parameters. Encode them for the SQL mock.
type logValueConverter struct{}

func (logValueConverter) ConvertValue(value any) (driver.Value, error) {
	if attrs, ok := value.(map[string]string); ok {
		return json.Marshal(attrs)
	}
	return driver.DefaultParameterConverter.ConvertValue(value)
}

func TestPushLogEventFields(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.ValueConverterOption(logValueConverter{}))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	cfg := createDefaultConfig().(*Config)
	e := &logsExporter{client: db, insertSQL: renderInsertLogsSQL(cfg), logger: zap.NewNop()}
	input := plog.NewLogs()
	resource := input.ResourceLogs().AppendEmpty()
	attrs := resource.Resource().Attributes()
	attrs.PutStr("service.namespace", "app")
	attrs.PutStr("service.name", "reader")
	attrs.PutStr("service.version", "1.2")
	attrs.PutStr("web.version", "3.4")
	attrs.PutStr("sophonz.tenant.id", "tenant-a")
	record := resource.ScopeLogs().AppendEmpty().LogRecords().AppendEmpty()
	record.SetEventName("session.start")
	record.Body().SetStr("started")
	mock.ExpectBegin()
	prepared := mock.ExpectPrepare(regexp.QuoteMeta(e.insertSQL))
	prepared.ExpectExec().WithArgs(sqlmock.AnyArg(), "", "", "", 0, "", 0, "session.start", "app", "reader", "1.2", "", "3.4", "started", "", sqlmock.AnyArg(), "", "", "", sqlmock.AnyArg(), sqlmock.AnyArg(), "tenant-a").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	if err := e.pushLogsData(t.Context(), input); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
