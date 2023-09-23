package crypto

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"io"
	"testing"

	"golang.org/x/crypto/ssh"

	"github.com/shuLhan/share/lib/test"
	"github.com/shuLhan/share/lib/test/mock"
)

func TestEncryptOaep(t *testing.T) {
	var (
		tdata *test.Data
		err   error
	)

	tdata, err = test.LoadData(`testdata/encrypt_oaep_test.txt`)
	if err != nil {
		t.Fatal(err)
	}

	var privkey crypto.PrivateKey

	privkey, err = ssh.ParseRawPrivateKey(tdata.Input[`private_key.pem`])
	if err != nil {
		t.Fatal(err)
	}

	var (
		rsakey *rsa.PrivateKey
		ok     bool
	)

	rsakey, ok = privkey.(*rsa.PrivateKey)
	if !ok {
		t.Fatalf(`expecting %T, got %T`, rsakey, privkey)
	}

	var (
		rsapub   = &rsakey.PublicKey
		hash     = sha256.New()
		expPlain = tdata.Input[`plain.txt`]
		maxSize  = rsapub.Size() - 2*hash.Size() - 2

		cipher []byte
	)

	t.Logf(`message size = %d`, len(expPlain))
	t.Logf(`max message size = public key size (%d) - 2*hash.Size (%d) - 2 = %d`,
		rsapub.Size(), 2*hash.Size(), maxSize)

	cipher, err = rsa.EncryptOAEP(hash, rand.Reader, rsapub, expPlain, nil)
	if err != nil {
		var expError = string(tdata.Output[`error_too_long`])
		test.Assert(t, `rsa.EncryptOAEP`, expError, err.Error())
	}

	cipher, err = EncryptOaep(hash, rand.Reader, rsapub, expPlain, nil)
	if err != nil {
		t.Fatal(err)
	}

	var gotPlain []byte

	gotPlain, err = DecryptOaep(hash, rand.Reader, rsakey, cipher, nil)
	if err != nil {
		t.Fatal(err)
	}

	test.Assert(t, `DecryptOaep`, expPlain, gotPlain)
}

func TestLoadPrivateKeyInteractive(t *testing.T) {
	type testCase struct {
		file     string
		secret   string
		termrw   io.ReadWriter
		expError string
	}

	var (
		mockrw = mock.ReadWriter{}

		pkey crypto.PrivateKey
		err  error
		ok   bool
	)

	var cases = []testCase{{
		file:   `testdata/openssl_rsa_pass.key`,
		secret: "s3cret\r\n",
		termrw: &mockrw,
	}, {
		file:     `testdata/openssl_rsa_pass.key`,
		termrw:   &mockrw,
		expError: `LoadPrivateKeyInteractive: empty passphrase`,
	}, {
		file: `testdata/openssl_rsa_pass.key`,
		// Using nil (default to os.Stdin for termrw.
		termrw:   nil,
		expError: `LoadPrivateKeyInteractive: MakeRaw: inappropriate ioctl for device`,
	}}

	var c testCase

	for _, c = range cases {
		_, err = mockrw.BufRead.WriteString(c.secret)
		if err != nil {
			t.Fatal(err)
		}

		pkey, err = LoadPrivateKeyInteractive(c.termrw, c.file)
		if err != nil {
			test.Assert(t, `using os.Stdin in test`, c.expError, err.Error())
			continue
		}

		_, ok = pkey.(*rsa.PrivateKey)
		if !ok {
			test.Assert(t, `cast to *rsa.PrivateKey`, c.expError, err.Error())
			continue
		}
	}
}
