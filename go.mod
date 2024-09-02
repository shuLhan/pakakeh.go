module git.sr.ht/~shulhan/pakakeh.go

go 1.21

require (
	golang.org/x/crypto v0.26.0
	golang.org/x/net v0.28.0
	golang.org/x/sys v0.24.0
	golang.org/x/term v0.23.0
)

replace golang.org/x/crypto => git.sr.ht/~shulhan/go-x-crypto v0.22.1-0.20240504075244-918d40784a11

//replace golang.org/x/crypto => ../go-x-crypto

//replace golang.org/x/term => ../../../golang.org/x/term
