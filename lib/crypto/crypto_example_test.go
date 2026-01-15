// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2023 M. Shulhan <ms@kilabit.info>

package crypto

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"fmt"
	"log"
)

// ExampleLoadPrivateKey test loading private key from file and convert the
// returned type to original type.
func ExampleLoadPrivateKey() {
	var (
		file string
		err  error
		pkey crypto.PrivateKey
		ok   bool
	)

	// RSA key with PKCS#8 generated using openssl v3.1.2,
	//	$ openssl genpkey -algorithm rsa -out openssl_rsa.key

	file = `testdata/openssl_rsa.key`
	pkey, err = LoadPrivateKey(file, nil)
	if err != nil {
		log.Fatalf(`%s: %s`, file, err)
	}

	_, ok = pkey.(*rsa.PrivateKey)
	if !ok {
		log.Fatalf(`expecting *rsa.PrivateKey, got %T`, pkey)
	}
	fmt.Printf("Loaded %T from %s\n", pkey, file)

	// ecdsa key generated using openssl v3.1.2,
	//	$ openssl ecparam -name prime256v1 -genkey -noout -out openssl_ecdsa.key

	file = `testdata/openssl_ecdsa.key`
	pkey, err = LoadPrivateKey(file, nil)
	if err != nil {
		log.Fatalf(`%s: %s`, file, err)
	}

	_, ok = pkey.(*ecdsa.PrivateKey)
	if !ok {
		log.Fatalf(`expecting *ecdsa.PrivateKey, got %T`, pkey)
	}
	fmt.Printf("Loaded %T from %s\n", pkey, file)

	// ed25519 key generated using openssh v9.4p1-4.

	file = `testdata/openssh_ed25519.key`
	pkey, err = LoadPrivateKey(file, nil)
	if err != nil {
		log.Fatalf(`%s: %s`, file, err)
	}

	_, ok = pkey.(*ed25519.PrivateKey)
	if !ok {
		log.Fatalf(`expecting *ed25519.PrivateKey, got %T`, pkey)
	}
	fmt.Printf("Loaded %T from %s\n", pkey, file)

	// Output:
	// Loaded *rsa.PrivateKey from testdata/openssl_rsa.key
	// Loaded *ecdsa.PrivateKey from testdata/openssl_ecdsa.key
	// Loaded *ed25519.PrivateKey from testdata/openssh_ed25519.key
}

// RSA key with PKCS#8 generated using openssl v3.1.2,
//
//	$ openssl genpkey -algorithm rsa -out openssl_rsa.key
//
// and then encrypted using passphrase,
//
//	$ cp openssl_rsa.key openssl_rsa_pass.key
//	$ ssh-keygen -p -f openssl_rsa_pass.key -N s3cret
//
// Using openssl to encrypt private key will cause the
// LoadPrivateKey return an error,
//
//	unsupported key type "ENCRYPTED PRIVATE KEY"
//
// ecdsa key generated using openssl v3.1.2,
//
//	$ openssl ecparam -name prime256v1 -genkey -noout \
//		-out openssl_ecdsa.key
//
// and then ecrypted using passphrase,
//
//	$ openssl ec -aes256 -in openssl_ecdsa.key -out \
//		openssl_ecdsa_pass.key -passout pass:s3cret
//
// ed25519 key generated using openssh v9.4p1-4,
//
//	$ ssh-keygen -t ed25519 -f openssh_ed25519.key
//
// and then encrypted using passphrase,
//
//	$ cp openssh_ed25519.key openssh_ed25519_pass.key
//	$ ssh-keygen -p -f openssh_ed25519_pass.key -N s3cret
func ExampleLoadPrivateKey_withPassphrase() {
	var (
		passphrase = []byte(`s3cret`)

		file string
		err  error
		pkey crypto.PrivateKey
		ok   bool
	)

	file = `testdata/openssl_rsa_pass.key`
	pkey, err = LoadPrivateKey(file, passphrase)
	if err != nil {
		log.Fatalf(`%s: %s`, file, err)
	}

	_, ok = pkey.(*rsa.PrivateKey)
	if !ok {
		log.Fatalf(`expecting *rsa.PrivateKey, got %T`, pkey)
	}
	fmt.Printf("Loaded %T with passphrase from %s\n", pkey, file)

	file = `testdata/openssl_ecdsa_pass.key`
	pkey, err = LoadPrivateKey(file, passphrase)
	if err != nil {
		log.Fatalf(`%s: %s`, file, err)
	}

	_, ok = pkey.(*ecdsa.PrivateKey)
	if !ok {
		log.Fatalf(`expecting *ecdsa.PrivateKey, got %T`, pkey)
	}
	fmt.Printf("Loaded %T with passphrase from %s\n", pkey, file)

	file = `testdata/openssh_ed25519_pass.key`
	pkey, err = LoadPrivateKey(file, passphrase)
	if err != nil {
		log.Fatalf(`%s: %s`, file, err)
	}

	_, ok = pkey.(*ed25519.PrivateKey)
	if !ok {
		log.Fatalf(`expecting *ed25519.PrivateKey, got %T`, pkey)
	}
	fmt.Printf("Loaded %T with passphrase from %s\n", pkey, file)

	// Output:
	// Loaded *rsa.PrivateKey with passphrase from testdata/openssl_rsa_pass.key
	// Loaded *ecdsa.PrivateKey with passphrase from testdata/openssl_ecdsa_pass.key
	// Loaded *ed25519.PrivateKey with passphrase from testdata/openssh_ed25519_pass.key
}
