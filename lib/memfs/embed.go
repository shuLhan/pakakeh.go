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
	Opts     *Options
	Node     *Node
	PathNode *PathNode
}

//
// GoEmbed write the tree nodes as Go generated source file.
//
func (mfs *MemFS) GoEmbed() (err error) {
	logp := "GoEmbed"

	if len(mfs.Opts.Embed.PackageName) == 0 {
		mfs.Opts.Embed.PackageName = DefaultEmbedPackageName
	}
	if len(mfs.Opts.Embed.VarName) == 0 {
		mfs.Opts.Embed.VarName = DefaultEmbedVarName
	}
	if len(mfs.Opts.Embed.GoFileName) == 0 {
		mfs.Opts.Embed.GoFileName = DefaultEmbedGoFileName
	}
	genData := &generateData{
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

	err = tmpl.ExecuteTemplate(f, templateNameHeader, mfs.Opts.Embed.PackageName)
	if err != nil {
		goto fail
	}

	for x := 0; x < len(names); x++ {
		// Ignore and delete the file from map if its the output
		// itself.
		if strings.HasSuffix(names[x], mfs.Opts.Embed.GoFileName) {
			mfs.PathNodes.Delete(names[x])
			continue
		}

		genData.Node = mfs.PathNodes.Get(names[x])

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
