package crypto

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
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
