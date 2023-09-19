package crypto

import (
	"crypto"
	"crypto/rsa"
	"testing"

	"github.com/shuLhan/share/lib/test"
	"github.com/shuLhan/share/lib/test/mock"
)

func TestLoadPrivateKeyInteractive(t *testing.T) {
	var (
		mockrw = mock.ReadWriter{}
		file   = `testdata/openssl_rsa_pass.key`

		pkey     crypto.PrivateKey
		expError string
		err      error
		ok       bool
	)

	_, err = mockrw.BufRead.WriteString("s3cret\r\n")
	if err != nil {
		t.Fatal(err)
	}

	pkey, err = LoadPrivateKeyInteractive(&mockrw, file)
	if err != nil {
		t.Fatal(err)
	}

	_, ok = pkey.(*rsa.PrivateKey)
	if !ok {
		t.Fatalf(`expecting *rsa.PrivateKey, got %T`, pkey)
	}

	// Using nil (os.Stdin) for termrw.

	file = `testdata/openssh_ed25519_pass.key`
	expError = `LoadPrivateKeyInteractive: MakeRaw: inappropriate ioctl for device`

	pkey, err = LoadPrivateKeyInteractive(nil, file)
	if err != nil {
		test.Assert(t, `using os.Stdin in test`, expError, err.Error())
	}
}
