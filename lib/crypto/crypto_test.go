package crypto

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"io"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/crypto/ssh"

	"git.sr.ht/~shulhan/pakakeh.go/lib/test"
	"git.sr.ht/~shulhan/pakakeh.go/lib/test/mock"
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
		desc     string
		file     string
		secret   string
		termrw   io.ReadWriter
		expError string

		// Environment variables for testing with SSH_ASKPASS and
		// SSH_ASKPASS_REQUIRE.
		envDisplay           string
		envSSHAskpass        string
		envSSHAskpassRequire string
	}

	var (
		wd  string
		err error
	)

	wd, err = os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	var (
		mockrw         = mock.ReadWriter{}
		askpassProgram = filepath.Join(wd, `testdata`, `askpass.sh`)

		pkey crypto.PrivateKey
		ok   bool
	)

	var cases = []testCase{{
		desc:   `withValidPassphrase`,
		file:   `testdata/openssl_rsa_pass.key`,
		secret: "s3cret\r\n",
		termrw: &mockrw,
	}, {
		desc:     `withMockedrw`,
		file:     `testdata/openssl_rsa_pass.key`,
		termrw:   &mockrw,
		expError: `LoadPrivateKeyInteractive: empty passphrase`,
	}, {
		desc:     `withDefaultTermrw`,
		file:     `testdata/openssl_rsa_pass.key`,
		termrw:   nil, // Using nil default to os.Stdin for termrw.
		expError: `LoadPrivateKeyInteractive: cannot read passhprase from stdin`,
	}, {
		desc:                 `withAskpassRequire=prefer`,
		file:                 `testdata/openssl_rsa_pass.key`,
		envDisplay:           `:0`,
		envSSHAskpass:        askpassProgram,
		envSSHAskpassRequire: `prefer`,
	}, {
		desc:                 `withAskpassRequire=prefer, no DISPLAY`,
		file:                 `testdata/openssl_rsa_pass.key`,
		envSSHAskpass:        askpassProgram,
		envSSHAskpassRequire: `prefer`,
		expError:             `LoadPrivateKeyInteractive: cannot read passhprase from stdin`,
	}, {
		desc:                 `withAskpassRequire=prefer, empty SSH_ASKPASS`,
		file:                 `testdata/openssl_rsa_pass.key`,
		envDisplay:           `:0`,
		envSSHAskpassRequire: `prefer`,
		expError:             `LoadPrivateKeyInteractive: cannot read passhprase from stdin`,
	}, {
		desc:                 `withAskpassRequire=prefer, invalid program`,
		file:                 `testdata/openssl_rsa_pass.key`,
		envDisplay:           `:0`,
		envSSHAskpass:        `/invalid/program`,
		envSSHAskpassRequire: `prefer`,
		expError:             `LoadPrivateKeyInteractive: fork/exec /invalid/program: no such file or directory`,
	}, {
		desc:                 `withAskpassRequire=force`,
		file:                 `testdata/openssl_rsa_pass.key`,
		envDisplay:           `:0`,
		envSSHAskpass:        askpassProgram,
		envSSHAskpassRequire: `force`,
	}}

	var c testCase

	for _, c = range cases {
		os.Setenv(envKeyDisplay, c.envDisplay)
		os.Setenv(envKeySSHAskpass, c.envSSHAskpass)
		os.Setenv(envKeySSHAskpassRequire, c.envSSHAskpassRequire)

		_, err = mockrw.BufRead.WriteString(c.secret)
		if err != nil {
			t.Fatalf(`%s: %s`, c.desc, err)
		}

		pkey, err = LoadPrivateKeyInteractive(c.termrw, c.file)
		if err != nil {
			test.Assert(t, c.desc+` error`, c.expError, err.Error())
			continue
		}

		_, ok = pkey.(*rsa.PrivateKey)
		if !ok {
			test.Assert(t, c.desc+` cast to *rsa.PrivateKey`, c.expError, err.Error())
			continue
		}
	}
}
