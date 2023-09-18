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
	pkey, err = LoadPrivateKey(file)
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
	pkey, err = LoadPrivateKey(file)
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
	pkey, err = LoadPrivateKey(file)
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
