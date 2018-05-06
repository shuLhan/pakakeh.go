package ini

type lineMode uint

const (
	lineModeNewline lineMode = 1 << iota
	lineModeComment
	lineModeSection
	lineModeSubsection
	lineModeVar
	lineModeVarMulti
)

//
// parsedLine define the single line, where `m` contain mode of line, `n`
// contain line number, and `v` contain the content of line itself.
//
type parsedLine struct {
	m lineMode
	n int
	v []byte
}
