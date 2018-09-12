package io

import (
	"os"
	"strconv"
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

	sdebug := os.Getenv("DEBUG")
	if len(sdebug) > 0 {
		_debug, _ = strconv.Atoi(sdebug)
	}

	s := m.Run()

	cleanup()

	os.Exit(s)
}
