// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package strings

// wikiMarkup define the markup for Wikimedia software.
type wikiMarkup struct {
	begin string
	end   string
}

//
// listWikiMarkup contains list of common markup in Wikimedia software.
//
var listWikiMarkup = []wikiMarkup{{
	begin: "[[Category:",
	end:   "]]",
}, {
	begin: "[[:Category:",
	end:   "]]",
}, {
	begin: "[[File:",
	end:   "]]",
}, {
	begin: "[[Help:",
	end:   "]]",
}, {
	begin: "[[Image:",
	end:   "]]",
}, {
	begin: "[[Special:",
	end:   "]]",
}, {
	begin: "[[Wikipedia:",
	end:   "]]",
}, {
	begin: "{{DEFAULTSORT:",
	end:   "}}",
}, {
	begin: "{{Template:",
	end:   "}}",
}, {
	begin: "<ref",
	end:   "/>",
}}
