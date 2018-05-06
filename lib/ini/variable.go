package ini

type varMode uint

const (
	varModeNewline varMode = 1 << iota
	varModeComment
	varModeNormal
)

var (
	varNewline   = &variable{m: varModeNewline, k: nil, v: nil, c: nil}
	varValueTrue = []byte("true")
)

type variable struct {
	m varMode
	k []byte
	v []byte
	c []byte
}
