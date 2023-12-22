module github.com/shuLhan/share

go 1.20

require (
	golang.org/x/crypto v0.17.0
	golang.org/x/net v0.19.0
	golang.org/x/sys v0.15.0
	golang.org/x/term v0.15.0
)

//replace golang.org/x/term => ../../../golang.org/x/term
