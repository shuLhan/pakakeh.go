// Copyright 2023, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package telemetry

import (
	"bytes"
	"fmt"

	libbytes "git.sr.ht/~shulhan/pakakeh.go/lib/bytes"
)

const (
	ilpFormatterName = `ilp`
)

// IlpFormatter format the Metric using the Influxdata Line Protocol, [ILP].
// Syntax,
//
//	ILP      = measurement [METADATA] " " METRIC [" " timestamp] LF
//	METADATA = *("," key "=" value)
//	METRIC   = key "=" value *("," METRIC)
//
// [ILP]: https://docs.influxdata.com/influxdb/cloud/reference/syntax/line-protocol/
type IlpFormatter struct {
	name        string
	measurement string
	mdRaw       []byte
	mdVersion   int // Latest version of metadata.

}

// NewIlpFormatter create and initialize new IlpFormatter.
func NewIlpFormatter(measurement string) (ilp *IlpFormatter) {
	ilp = &IlpFormatter{
		name:        ilpFormatterName,
		measurement: measurement,
	}
	return ilp
}

// BulkFormat format list of Metric with metadata.
func (ilp *IlpFormatter) BulkFormat(list []Metric, md *Metadata) []byte {
	if len(list) == 0 {
		return nil
	}

	ilp.generateMetadata(md)

	var (
		bb bytes.Buffer
		m  Metric
		x  int
	)

	bb.WriteString(ilp.measurement)
	bb.Write(ilp.mdRaw)
	bb.WriteByte(' ')

	for _, m = range list {
		if len(m.Name) == 0 {
			continue
		}
		if x > 0 {
			bb.WriteByte(',')
		}
		fmt.Fprintf(&bb, `%s=%f`, m.Name, m.Value)
		x++
	}

	fmt.Fprintf(&bb, " %d\n", m.Timestamp)

	return libbytes.Copy(bb.Bytes())
}

// Format single Metric.
func (ilp *IlpFormatter) Format(m Metric, md *Metadata) []byte {
	if len(m.Name) == 0 {
		return nil
	}
	return ilp.BulkFormat([]Metric{m}, md)
}

func (ilp *IlpFormatter) generateMetadata(md *Metadata) {
	var mdVersion = md.Version()
	if ilp.mdVersion == mdVersion {
		return
	}

	var (
		keys, vals = md.KeysMap()

		bb bytes.Buffer
		k  string
		v  string
	)
	for _, k = range keys {
		v = vals[k]
		fmt.Fprintf(&bb, `,%s=%s`, k, v)
	}
	ilp.mdRaw = libbytes.Copy(bb.Bytes())
	ilp.mdVersion = mdVersion
}

// Name return the unique name of the formatter, "ilp".
func (ilp *IlpFormatter) Name() string {
	return ilp.name
}
