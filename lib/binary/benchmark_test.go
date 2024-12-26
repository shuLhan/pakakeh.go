// SPDX-FileCopyrightText: 2024 M. Shulhan <ms@kilabit.info>
//
// SPDX-License-Identifier: BSD-3-Clause

package binary

import (
	"bytes"
	"encoding/binary"
	"testing"
)

// This benchmark is taken from https://github.com/golang/go/issues/27757 .
func BenchmarkStd_BinaryWrite_SliceByte(b *testing.B) {
	var data = make([]byte, 2e6)
	var buf = bytes.NewBuffer([]byte{})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		binary.Write(buf, binary.BigEndian, data)
	}
}

func BenchmarkBigEndianBuffer_Write_SliceByte(b *testing.B) {
	var data = make([]byte, 2e6)
	var beb BigEndianBuffer
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		beb.Write(data)
	}
}

// This benchmark is taken from https://github.com/golang/go/issues/70503 .
func BenchmarkStd_BinaryWrite_SliceFloat32(b *testing.B) {
	var data = make([]float32, 1000)
	for i := range data {
		data[i] = float32(i) / 42
	}
	var buf bytes.Buffer
	b.SetBytes(int64(4 * len(data)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		binary.Write(&buf, binary.BigEndian, data)
	}
}

func BenchmarkBigEndianBuffer_Write_SliceFloat32(b *testing.B) {
	data := make([]float32, 1000)
	for i := range data {
		data[i] = float32(i) / 42
	}
	var beb BigEndianBuffer
	b.SetBytes(int64(4 * len(data)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		beb.Write(data)
	}
}
