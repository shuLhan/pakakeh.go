module github.com/shuLhan/share

go 1.20

require (
	golang.org/x/crypto v0.18.0
	golang.org/x/net v0.20.0
	golang.org/x/sys v0.16.0
	golang.org/x/term v0.16.0
)

replace golang.org/x/crypto => git.sr.ht/~shulhan/go-x-crypto v0.18.1-0.20240119171712-4b35f92ea767

//replace golang.org/x/term => ../../../golang.org/x/term
