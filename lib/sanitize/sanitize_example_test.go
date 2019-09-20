package sanitize

import (
	"fmt"
)

func ExampleHTML() {
	input := `
<html>
	<title>Test</title>
	<head>
	</head>
	<body>
		This
		<p> is </p>
		a
		<a href="/">link</a>.
		An another
		<a href="/">link</a>.
	</body>
</html>
`

	out := HTML([]byte(input))
	fmt.Printf("%s", out)
	// Output:
	// This is a link. An another link.
}
