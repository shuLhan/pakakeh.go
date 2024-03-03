// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package memfs

import (
	"fmt"
	"os"
	"strings"
)

// DefaultEmbedPackageName default package name for GoEmbed.
const DefaultEmbedPackageName = `main`

// DefaultEmbedVarName default variable name for GoEmbed.
const DefaultEmbedVarName = `memFS`

// DefaultEmbedGoFileName default file output for GoEmbed.
const DefaultEmbedGoFileName = `memfs_generate.go`

type generateData struct {
	Opts     *Options
	Node     *Node
	PathNode *PathNode
}

// GoEmbed write the tree nodes as Go generated source file.
// This method assume that the files inside the mfs instance is already
// up-to-date.
// If you are not sure, call Remount.
func (mfs *MemFS) GoEmbed() (err error) {
	var (
		logp = "GoEmbed"

		node    *Node
		genData *generateData
		name    string
	)

	if len(mfs.Opts.Embed.PackageName) == 0 {
		mfs.Opts.Embed.PackageName = DefaultEmbedPackageName
	}
	if len(mfs.Opts.Embed.VarName) == 0 {
		mfs.Opts.Embed.VarName = DefaultEmbedVarName
	}
	if len(mfs.Opts.Embed.GoFileName) == 0 {
		mfs.Opts.Embed.GoFileName = DefaultEmbedGoFileName
	}

	genData = &generateData{
		Opts:     mfs.Opts,
		PathNode: mfs.PathNodes,
	}

	tmpl, err := generateTemplate()
	if err != nil {
		return fmt.Errorf("%s: %w", logp, err)
	}

	f, err := os.Create(mfs.Opts.Embed.GoFileName)
	if err != nil {
		return fmt.Errorf("%s: %w", logp, err)
	}

	names := mfs.ListNames()

	err = tmpl.ExecuteTemplate(f, templateNameHeader, mfs.Opts.Embed)
	if err != nil {
		goto fail
	}

	for _, name = range names {
		node = mfs.PathNodes.Get(name)

		// Ignore and delete the file from map if its the output
		// itself.
		if strings.HasSuffix(name, mfs.Opts.Embed.GoFileName) {
			mfs.PathNodes.Delete(name)
			continue
		}

		if len(node.GenFuncName) == 0 {
			// Node is watched only, not included.
			continue
		}

		genData.Node = node

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
			return fmt.Errorf(`%s: %w: %w`, logp, errClose, err)
		}
		return fmt.Errorf("%s: %w", logp, errClose)
	}
	if err != nil {
		return fmt.Errorf("%s: %w", logp, err)
	}
	return nil
}
