// Command bcrypt implements command-line interface to generate and compare
// bcrypt hash.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"golang.org/x/crypto/bcrypt"
)

const (
	cmd        = `bcrypt`
	cmdCompare = `compare`
	cmdGen     = `gen`
)

func main() {
	log.SetFlags(0)

	var optCost int

	flag.Usage = usage
	flag.IntVar(&optCost, `cost`, 10, `the hashing cost used to generate hash`)
	flag.Parse()

	var cmd = flag.Arg(0)

	switch cmd {
	case cmdCompare:
		doCompare(flag.Arg(1), flag.Arg(2))
	case cmdGen:
		doGen(flag.Arg(1), optCost)
	default:
		usage()
		os.Exit(1)
	}
}

// doCompare compare hashed password with plain text password.
func doCompare(hash, pass string) {
	var err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(pass))
	if err != nil {
		log.Fatalf(`%s: %s`, cmd, err)
	}
	fmt.Println(`OK`)
}

func doGen(pass string, optCost int) {
	if pass == `-` {
		fmt.Fscanln(os.Stdin, &pass)
	}

	var (
		hash []byte
		err  error
	)

	hash, err = bcrypt.GenerateFromPassword([]byte(pass), optCost)
	if err != nil {
		log.Fatalf(`%s: %s`, cmd, err)
	}

	fmt.Printf("%s", hash)
}

func usage() {
	fmt.Println(`bcrypt - compare or generate hash using bcrypt.

== SYNOPSIS

	bcrypt [OPTIONS] [COMMAND]

=== OPTIONS

-cost <number>
	The hashing cost used to generate hash, default to 10.

=== COMMANDS

compare <hash> <plain>
	Compare the hashed password with its plain text.

gen <string> | -

	<string>
		The string to be hashed.
	-
		Read from string to be hashed from standard input.`)
}
