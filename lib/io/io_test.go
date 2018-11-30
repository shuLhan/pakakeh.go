package io

import (
	"os"
	"testing"
)

func cleanup() {
	// Cleaning up TestRmdirEmptyAll
	_ = os.Remove("testdata/file")
	_ = os.RemoveAll("testdata/a")
	_ = os.RemoveAll("testdata/dirempty")
}

func TestMain(m *testing.M) {
	cleanup()

	s := m.Run()

	cleanup()

	os.Exit(s)
}
