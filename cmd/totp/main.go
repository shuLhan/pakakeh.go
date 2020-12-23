//
// Program totp generate Time-based One-time Password from secret key.
//
package main

import (
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base32"
	"flag"
	"fmt"
	"hash"
	"log"
	"os"

	"github.com/shuLhan/share/lib/totp"
)

func main() {
	log.SetFlags(0)

	flag.Usage = usage

	paramDigits := flag.Int("digits", 6, "number of digits to generated")
	paramHash := flag.String("alg", "sha1", "hash name, valid values is sha1, sha256, sha512")
	paramTimestep := flag.Int("period", 30, "time step in seconds")
	paramHelp := flag.Bool("help", false, "show command usage")
	flag.Parse()

	if *paramHelp {
		flag.Usage()
	}
	if len(os.Args) == 1 {
		flag.Usage()
	}

	var hashFn func() hash.Hash
	switch *paramHash {
	case "sha256":
		hashFn = sha256.New
	case "sha512":
		hashFn = sha512.New
	default:
		hashFn = sha1.New
	}

	totproto := totp.New(hashFn, *paramDigits, *paramTimestep)
	secret, err := base32.StdEncoding.DecodeString(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	listOTP, err := totproto.GenerateN(secret, 3)
	if err != nil {
		log.Fatal(err)
	}

	for _, otp := range listOTP {
		fmt.Printf("%s\n", otp)
	}
}

func usage() {
	log.Printf(`%s is command line interface to generate time-based one-time password.
Usage:
	%s [OPTIONS] <BASE32-SECRET>

Available OPTIONS:
`, os.Args[0], os.Args[0])
	flag.PrintDefaults()
	os.Exit(2)
}
