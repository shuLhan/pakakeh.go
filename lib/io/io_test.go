package io

import (
	"os"
	"strconv"
	"testing"
)

func TestMain(m *testing.M) {
	sdebug := os.Getenv("DEBUG")
	if len(sdebug) > 0 {
		_debug, _ = strconv.Atoi(sdebug)
	}
	os.Exit(m.Run())
}
