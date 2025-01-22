// SPDX-FileCopyrightText: 2023 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package telemetry

import (
	"bytes"
	"fmt"
)

// DsvFormatter format the [Metric] in single line where each value is separated
// by single character.
// The metric are formatted in the following order,
//
//	Timestamp SEP Name SEP Value *(SEP metadata)
//	metadata = Metadata.Key "=" Metadata.Value *("," metadata)
//
// The Name, Value, and metadata are enclosed with double quoted.
type DsvFormatter struct {
	metricsAlias map[string]string
	name         string
	mdRaw        []byte
	sep          rune
	mdVersion    int // Latest version of metadata.
}

// NewDsvFormatter create new DsvFormatter using sep as separater and options
// to change the metric output name using metricsAlias.
// See [RuntimeMetricsAlias] for example.
func NewDsvFormatter(sep rune, metricsAlias map[string]string) (dsv *DsvFormatter) {
	dsv = &DsvFormatter{
		name:         `dsv`,
		sep:          sep,
		metricsAlias: metricsAlias,
	}
	return dsv
}

// BulkFormat bulk format list of Metric with [Metadata].
func (dsv *DsvFormatter) BulkFormat(listm []Metric, md *Metadata) []byte {
	if len(listm) == 0 {
		return nil
	}

	dsv.generateMetadata(md)

	var (
		bb bytes.Buffer
		m  Metric
	)
	for _, m = range listm {
		dsv.formatMetric(&bb, m)
	}
	return bytes.Clone(bb.Bytes())
}

// Format single Metric into single line DSV.
func (dsv *DsvFormatter) Format(m Metric, md *Metadata) []byte {
	return dsv.BulkFormat([]Metric{m}, md)
}

func (dsv *DsvFormatter) formatMetric(bb *bytes.Buffer, m Metric) {
	if len(m.Name) == 0 {
		return
	}

	var name = dsv.metricsAlias[m.Name]
	if len(name) == 0 {
		name = m.Name
	}
	fmt.Fprintf(bb, "%d%c%q%c%f%c%q\n", m.Timestamp, dsv.sep, name, dsv.sep, m.Value, dsv.sep, dsv.mdRaw)
}

func (dsv *DsvFormatter) generateMetadata(md *Metadata) {
	var mdVersion = md.Version()
	if dsv.mdVersion == mdVersion {
		return
	}

	var (
		keys, vals = md.KeysMap()

		bb bytes.Buffer
		k  string
		v  string
		x  int
	)
	for x, k = range keys {
		if x > 0 {
			bb.WriteByte(',')
		}
		v = vals[k]
		fmt.Fprintf(&bb, `%s=%s`, k, v)
	}
	dsv.mdRaw = bytes.Clone(bb.Bytes())
	dsv.mdVersion = mdVersion
}

// Name return the Name of DsvFormatter as "dsv".
func (dsv *DsvFormatter) Name() string {
	return dsv.name
}
