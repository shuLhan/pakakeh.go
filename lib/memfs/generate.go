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
	defGeneratedVarName = "memfsPathNode"
)

type generateData struct {
	VarName string
	Node    *Node
	Nodes   map[string]*Node
}

//
// GoGenerate write the tree nodes as Go generated source file.
//
// If pkgName is not defined it will be default to "main".
//
// varName is the global variable name that will return the memfs root
// PathNode, which can be used to initilize New() function.
//
// If out is not defined it will be default "memfs_generate.go" and saved in
// current directory.
//
// If contentEncoding is not empty, it will encode the content of node and set
// the node ContentEncoding.
// List of available encoding is "gzip".
// For example, if contentEncoding is "gzip" it will compress the content of
// file using gzip and set "ContentEncoding" to "gzip".
//
func (mfs *MemFS) GoGenerate(pkgName, varName, out, contentEncoding string) (err error) {
	if len(pkgName) == 0 {
		pkgName = "main"
	}
	if len(varName) == 0 {
		varName = defGeneratedVarName
	}
	if len(out) == 0 {
		out = "memfs_generate.go"
	}
	genData := &generateData{
		VarName: varName,
		Nodes:   mfs.pn.v,
	}

	tmpl, err := generateTemplate()
	if err != nil {
		return err
	}

	f, err := os.Create(out)
	if err != nil {
		return fmt.Errorf("memfs: GoGenerate: %w", err)
	}

	if len(contentEncoding) > 0 {
		err = mfs.ContentEncode(contentEncoding)
		if err != nil {
			return fmt.Errorf("GoGenerate: %w", err)
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
			delete(mfs.pn.v, names[x])
			continue
		}

		genData.Node = mfs.pn.v[names[x]]

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

	err = f.Close()
	if err != nil {
		return fmt.Errorf("memfs: GoGenerate: %w", err)
	}

	return nil
fail:
	_ = f.Close()
	return fmt.Errorf("memfs: GoGenerate: %w", err)
}
