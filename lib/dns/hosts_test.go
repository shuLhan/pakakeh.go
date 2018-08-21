package dns

import (
	"testing"
)

func TestHostsLoad(t *testing.T) {
	msgs, err := HostsLoad("testdata/hosts")
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("msgs: %s\n", msgs)
}
