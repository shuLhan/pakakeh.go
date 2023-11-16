module github.com/shuLhan/share

go 1.20

require (
	golang.org/x/crypto v0.15.0
	golang.org/x/net v0.18.0
	golang.org/x/sys v0.14.0
	golang.org/x/term v0.14.0
)

//replace golang.org/x/term => ../../../golang.org/x/term
