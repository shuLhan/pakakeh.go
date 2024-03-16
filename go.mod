module git.sr.ht/~shulhan/pakakeh.go

go 1.21

require (
	golang.org/x/crypto v0.21.0
	golang.org/x/net v0.22.0
	golang.org/x/sys v0.18.0
	golang.org/x/term v0.18.0
)

replace golang.org/x/crypto => git.sr.ht/~shulhan/go-x-crypto v0.21.1-0.20240316083930-db093b454c7e

//replace golang.org/x/term => ../../../golang.org/x/term
