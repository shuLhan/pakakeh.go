package dns

import (
	"testing"

	"github.com/shuLhan/share/lib/test"
)

func TestParseHostsFile(t *testing.T) {
	hostsFile, err := ParseHostsFile("testdata/hosts")
	if err != nil {
		t.Fatal(err)
	}

	test.Assert(t, "Length", 10, len(hostsFile.Records))
}

func TestHostsLoad2(t *testing.T) {
	_, err := ParseHostsFile("testdata/hosts.block")
	if err != nil {
		t.Fatal(err)
	}
}
