package memfs

const (
	defaultMaxFileSize = 1024 * 1024 * 5
)

type Options struct {
	// Root contains path to directory where its contents will be mapped
	// to memory.
	Root string

	// GeneratedPathNode define the root directory.
	// If GeneratedPathNode is not nil, the Dir, Includes, and Excludes
	// options will not have any effects.
	GeneratedPathNode *PathNode

	// The includes and excludes pattern applied to path of file in file
	// system, not to the path in memory.
	//
	// If GeneratedPathNode is not nil, the includes and excludes does not
	// have any effect, since the content of path and nodes will be
	// overwritten by it.
	Includes []string
	Excludes []string

	// MaxFileSize define maximum file size that can be stored on memory.
	// The default value is 5 MB.
	// If its value is negative, the content of file will not be mapped to
	// memory, the MemFS will behave as directory tree.
	MaxFileSize int64

	// Development define a flag to bypass file in memory.
	// If its true, any call to Get will result in direct read to file
	// system.
	Development bool
}

//
// init initialize the options with default value.
//
func (opts *Options) init() {
	if opts.MaxFileSize == 0 {
		opts.MaxFileSize = defaultMaxFileSize
	}
}
