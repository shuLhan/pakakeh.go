package http

import "fmt"

type RangePosition struct {
	Start int64
	End   int64

	// Length of zero means read until the end.
	Length int64
}

// ContentRange return the string that can be used for HTTP Content-Range
// header value.
func (pos RangePosition) ContentRange(unit string, size int64) (v string) {
	if size == 0 {
		v = fmt.Sprintf(`%s %s/*`, unit, pos.String())
	} else {
		v = fmt.Sprintf(`%s %s/%d`, unit, pos.String(), size)
	}
	return v
}

func (pos RangePosition) String() string {
	if pos.Start < 0 {
		return fmt.Sprintf(`%d`, pos.Start)
	}
	if pos.Start > 0 && pos.End == 0 {
		return fmt.Sprintf(`%d-`, pos.Start)
	}
	return fmt.Sprintf(`%d-%d`, pos.Start, pos.End)
}
