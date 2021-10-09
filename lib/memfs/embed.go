// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package memfs

import (
	"fmt"
	"os"
	"strings"
)

const (
	DefaultEmbedPackageName = "main"              // Default package name for GoEmbed().
	DefaultEmbedVarName     = "memFS"             // Default variable name for GoEmbed().
	DefaultEmbedGoFileName  = "memfs_generate.go" // Default file output for GoEmbed().
)

type generateData struct {
	Opts    *Options
	VarName string
	Node    *Node
	Nodes   map[string]*Node
}

//
// GoEmbed write the tree nodes as Go generated source file.
//
// If pkgName is not defined it will be default to "main".
//
// varName is the global variable name with type *memfs.MemFS which will be
// initialize by generated Go source code on init().
// The varName default to "memFS" if its empty.
//
// If out is not defined it will be default to "memfs_generate.go" and saved
// in current directory from where its called.
//
// If contentEncoding is not empty, it will encode the content of node and set
// the node ContentEncoding.
// List of available encoding is "gzip".
// For example, if contentEncoding is "gzip" it will compress the content of
// file using gzip and set Node.ContentEncoding to "gzip".
//
func (mfs *MemFS) GoEmbed(pkgName, varName, out, contentEncoding string) (err error) {
	logp := "MemFS.GoEmbed"

	if len(pkgName) == 0 {
		pkgName = DefaultEmbedPackageName
	}
	if len(varName) == 0 {
		varName = DefaultEmbedVarName
	}
	if len(out) == 0 {
		out = DefaultEmbedGoFileName
	}
	genData := &generateData{
		Opts:    mfs.Opts,
		VarName: varName,
		Nodes:   mfs.PathNodes.v,
	}

	tmpl, err := generateTemplate()
	if err != nil {
		return fmt.Errorf("%s: %w", logp, err)
	}

	f, err := os.Create(out)
	if err != nil {
		return fmt.Errorf("%s: %w", logp, err)
	}

	if len(contentEncoding) > 0 {
		err = mfs.ContentEncode(contentEncoding)
		if err != nil {
			return fmt.Errorf("%s: %w", logp, err)
		}
	}

	names := mfs.ListNames()

	err = tmpl.ExecuteTemplate(f, templateNameHeader, pkgName)
	if err != nil {
		goto fail
	}

	for x := 0; x < len(names); x++ {
		// Ignore and delete the file from map if its the output
		// itself.
		if strings.HasSuffix(names[x], out) {
			delete(mfs.PathNodes.v, names[x])
			continue
		}

		genData.Node = mfs.PathNodes.v[names[x]]

		err = tmpl.ExecuteTemplate(f, templateNameGenerateNode, genData)
		if err != nil {
			goto fail
		}
	}

	err = tmpl.ExecuteTemplate(f, templateNamePathFuncs, genData)
	if err != nil {
		goto fail
	}

	err = f.Sync()
	if err != nil {
		goto fail
	}

fail:
	errClose := f.Close()
	if errClose != nil {
		if err != nil {
			return fmt.Errorf("%s: %s: %w", logp, errClose, err)
		}
		return fmt.Errorf("%s: %w", logp, errClose)
	}
	if err != nil {
		return fmt.Errorf("%s: %w", logp, err)
	}
	return nil
}
