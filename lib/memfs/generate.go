// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package memfs

import (
	"fmt"
	"os"
)

//
// GoGenerate write the tree nodes as Go generated source file.
// If pkgName is not defined it will be default to "main".
// If out is not defined it will be default "memfs_generate.go" and saved in
// current directory.
//
func (mfs *MemFS) GoGenerate(pkgName, out string) (err error) {
	if len(pkgName) == 0 {
		pkgName = "main"
	}
	if len(out) == 0 {
		out = "memfs_generate.go"
	}

	tmpl, err := generateTemplate()
	if err != nil {
		return err
	}

	f, err := os.Create(out)
	if err != nil {
		return fmt.Errorf("memfs: GoGenerate: " + err.Error())
	}

	names := mfs.ListNames()

	err = tmpl.ExecuteTemplate(f, "HEADER", pkgName)
	if err != nil {
		goto fail
	}

	for x := 0; x < len(names); x++ {
		node := mfs.pn.v[names[x]]
		err = tmpl.ExecuteTemplate(f, "GENERATE_NODE", node)
		if err != nil {
			goto fail
		}
	}

	err = tmpl.ExecuteTemplate(f, "PATHFUNCS", mfs.pn.v)
	if err != nil {
		goto fail
	}

	err = f.Close()
	if err != nil {
		return fmt.Errorf("memfs: GoGenerate: " + err.Error())
	}

	return nil
fail:
	_ = f.Close()
	return fmt.Errorf("memfs: GoGenerate: " + err.Error())
}
