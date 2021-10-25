//
// Program totp generate Time-based One-time Password from secret key.
//
package main

import (
	"encoding/base32"
	"flag"
	"fmt"
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

	var cryptoHash totp.CryptoHash
	switch *paramHash {
	case "sha256":
		cryptoHash = totp.CryptoHashSHA256
	case "sha512":
		cryptoHash = totp.CryptoHashSHA512
	default:
		cryptoHash = totp.CryptoHashSHA1
	}

	totproto := totp.New(cryptoHash, *paramDigits, *paramTimestep)
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
