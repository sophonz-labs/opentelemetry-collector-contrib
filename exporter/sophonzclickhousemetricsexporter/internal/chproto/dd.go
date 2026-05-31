// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

// Package chproto provides a self-contained port of the DDSketch ClickHouse
// AggregateFunction wire model used by the legacy SOPHONZ metrics exporter.
//
// The legacy exporter relied on the SigNoz `ch-go` fork
// (github.com/SigNoz/ch-go, package proto) for the DD/IndexMapping/Store types
// and on the SigNoz `clickhouse-go` fork for the matching custom column codec.
// Those forks are pinned to old versions that are incompatible with the
// upgraded ClickHouse driver (clickhouse-go/v2 v2.46.0) used by the ported
// SOPHONZ exporters. To keep this module self-contained on the standard
// toolchain, the data model and DDSketch wire format are reproduced here.
package chproto // import "github.com/open-telemetry/opentelemetry-collector-contrib/exporter/sophonzclickhousemetricsexporter/internal/chproto"

import (
	"encoding/binary"
	"errors"
	"math"
	"strconv"
)

// ErrInvalidFlag is returned when an unexpected flag byte is decoded.
var ErrInvalidFlag = errors.New("invalid flag")

// Flag layout (ported from the SigNoz ch-go fork, package proto):
//
//	type flag = 2 low bits, sub flag = 6 high bits.
const (
	numBitsForType byte = 2
	flagTypeMask   byte = (1 << numBitsForType) - 1

	flagTypeSketchFeatures byte = 0b00
	flagTypeIndexMapping   byte = 0b10
	flagTypePositiveStore  byte = 0b01
	flagTypeNegativeStore  byte = 0b11
)

func newFlag(t, sub byte) byte { return t | (sub << numBitsForType) }

var (
	flagZeroCountVarFloat           = newFlag(flagTypeSketchFeatures, 1)
	flagIndexMappingBaseLogarithmic = newFlag(flagTypeIndexMapping, 0)
	flagTypePositiveStoreByte       = flagTypePositiveStore
	flagTypeNegativeStoreByte       = flagTypeNegativeStore

	binEncodingIndexDeltasAndCounts = byte(1) << numBitsForType
	binEncodingContiguousCounts     = byte(3) << numBitsForType
)

// Buffer is a minimal little-endian append buffer compatible with the subset
// of operations used by the DDSketch encoding.
type Buffer struct {
	Buf []byte
}

// PutByte appends a single byte.
func (b *Buffer) PutByte(x byte) { b.Buf = append(b.Buf, x) }

// PutFloat64 appends a little-endian IEEE-754 float64.
func (b *Buffer) PutFloat64(v float64) {
	var tmp [8]byte
	binary.LittleEndian.PutUint64(tmp[:], math.Float64bits(v))
	b.Buf = append(b.Buf, tmp[:]...)
}

// PutUVarInt appends an unsigned varint.
func (b *Buffer) PutUVarInt(x uint64) {
	var tmp [binary.MaxVarintLen64]byte
	n := binary.PutUvarint(tmp[:], x)
	b.Buf = append(b.Buf, tmp[:n]...)
}

// PutVarInt appends a signed (zig-zag) varint.
func (b *Buffer) PutVarInt(x int64) {
	var tmp [binary.MaxVarintLen64]byte
	n := binary.PutVarint(tmp[:], x)
	b.Buf = append(b.Buf, tmp[:n]...)
}

// DD is a quantile sketch in which the bins have a size that is proportional to
// the fractional rank error that they incur. It is compatible with the DDSketch
// protobuf model.
type DD struct {
	Mapping        *IndexMapping
	PositiveValues *Store
	NegativeValues *Store
	ZeroCount      float64
}

// Encode encodes DD to buffer.
func (d DD) Encode(b *Buffer) {
	if b == nil {
		return
	}
	b.PutByte(flagIndexMappingBaseLogarithmic)
	d.Mapping.Encode(b)
	b.PutByte(flagTypePositiveStoreByte)
	d.PositiveValues.Encode(b)
	b.PutByte(flagTypeNegativeStoreByte)
	d.NegativeValues.Encode(b)
	b.PutByte(flagZeroCountVarFloat)
	b.PutFloat64(d.ZeroCount)
}

// Debug returns debug string.
func (d DD) Debug() string {
	var s string
	s += "Mapping:\n"
	if d.Mapping != nil {
		s += d.Mapping.Debug()
	}
	s += "\nPositive values:\n"
	if d.PositiveValues != nil {
		s += d.PositiveValues.Debug()
	}
	s += "\nNegative values:\n"
	if d.NegativeValues != nil {
		s += d.NegativeValues.Debug()
	}
	s += "\nZero count: "
	s += strconv.FormatFloat(d.ZeroCount, 'f', -1, 64)
	return s
}

// IndexMapping is a mapping from a bin index to a value.
type IndexMapping struct {
	Gamma       float64
	IndexOffset float64
}

// Encode encodes IndexMapping to buffer.
func (m IndexMapping) Encode(b *Buffer) {
	if b == nil {
		return
	}
	b.PutFloat64(m.Gamma)
	b.PutFloat64(m.IndexOffset)
}

// Debug returns debug string.
func (m IndexMapping) Debug() string {
	var s string
	s += "Gamma: "
	s += strconv.FormatFloat(m.Gamma, 'f', -1, 64)
	s += "\nIndex offset: "
	s += strconv.FormatFloat(m.IndexOffset, 'f', -1, 64)
	return s
}

// Store is a store of bin counts.
type Store struct {
	BinCounts map[int32]float64

	ContiguousBinCounts      []float64
	ContiguousBinIndexOffset int32
}

// Encode encodes Store to buffer.
func (s Store) Encode(b *Buffer) {
	if b == nil {
		return
	}
	if len(s.ContiguousBinCounts) > 0 {
		b.PutByte(binEncodingContiguousCounts)
		b.PutUVarInt(uint64(len(s.ContiguousBinCounts)))
		b.PutVarInt(int64(s.ContiguousBinIndexOffset))
		b.PutVarInt(1)
		for _, v := range s.ContiguousBinCounts {
			b.PutFloat64(v)
		}
	} else {
		b.PutByte(binEncodingIndexDeltasAndCounts)
		b.PutUVarInt(uint64(len(s.BinCounts)))
		for k, v := range s.BinCounts {
			b.PutVarInt(int64(k))
			b.PutFloat64(v)
		}
	}
}

// Debug returns debug string.
func (store Store) Debug() string {
	var s string
	if len(store.ContiguousBinCounts) > 0 {
		s += "Contiguous bin counts:\n"
		for i, v := range store.ContiguousBinCounts {
			s += strconv.Itoa(int(store.ContiguousBinIndexOffset) + i)
			s += ": "
			s += strconv.FormatFloat(v, 'f', -1, 64)
			s += ", "
		}
	} else {
		s += "Bin counts:\n"
		for k, v := range store.BinCounts {
			s += strconv.Itoa(int(k))
			s += ": "
			s += strconv.FormatFloat(v, 'f', -1, 64)
			s += ", "
		}
	}
	return s
}
