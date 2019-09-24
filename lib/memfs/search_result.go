package memfs

//
// SearchResult contains the result of searching where the Path will be
// filled with absolute path of file system in memory and the Snippet will
// filled with part of the text before and after the search string.
//
type SearchResult struct {
	Path     string
	Snippets []string
}
