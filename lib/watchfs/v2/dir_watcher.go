// SPDX-FileCopyrightText: 2024 M. Shulhan <ms@kilabit.info>
// SPDX-License-Identifier: BSD-3-Clause

package watchfs

import (
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"time"
)

// DirWatcher scan the content of directory in
// [watchfs.DirWatcherOptions.Root] recursively for the files to be watched,
// using the [watchfs.DirWatcherOptions.Includes] field.
// A single file, [watchfs.DirWatcherOptions.File], will be watched for
// changes that will trigger re-scanning the content of Root recursively.
//
// The result of re-scanning is list of the Includes files (only files not
// new directory) that are changes, which will be send to channel C.
// On each [os.FileInfo] received from C, a deleted file have
// [os.FileInfo.Size] equal to [FileFlagDeleted].
// The channel will send an empty slice if no changes.
//
// The implementation of file changes in this code is naive, using loop and
// comparison of mode, modification time, and size; at least it should works
// on most operating system.
type DirWatcher struct {
	// idxDir contains index of directory.
	// It is used to detect new or deleted file inside that directory.
	idxDir map[string]node

	// idxFile contains index of files.
	idxFile map[string]node

	idxNewFile map[string]node

	// C received the new, updated, and deleted files.
	C <-chan []os.FileInfo
	c chan []os.FileInfo

	fwatch *FileWatcher
	opts   DirWatcherOptions
}

// WatchDir create and start scanning directory for changes.
func WatchDir(opts DirWatcherOptions) (dwatch *DirWatcher, err error) {
	err = opts.init()
	if err != nil {
		return nil, fmt.Errorf(`WatchDir: %w`, err)
	}

	dwatch = &DirWatcher{
		c:          make(chan []os.FileInfo, 1),
		opts:       opts,
		idxDir:     map[string]node{},
		idxFile:    map[string]node{},
		idxNewFile: map[string]node{},
	}
	dwatch.C = dwatch.c
	dwatch.initialScan(dwatch.opts.Root)
	dwatch.fwatch = WatchFile(dwatch.opts.FileWatcherOptions)
	go dwatch.watch()
	return dwatch, nil
}

// Files return all the files currently being watched, the one that filtered
// by [watchfs.DirWatcherOptions.Includes], with its file information.
// This method is not safe when called when DirWatcher has been running.
func (dwatch *DirWatcher) Files() (files map[string]os.FileInfo) {
	files = make(map[string]os.FileInfo)
	for key, node := range dwatch.idxFile {
		if node.size == nodeFlagExcluded {
			continue
		}
		files[key] = &node
	}
	return files
}

// ForceRescan force to rescan for changes without waiting for
// [watchfs.DirWatcherOptions.File] to be updated.
func (dwatch *DirWatcher) ForceRescan() {
	if dwatch.fwatch != nil {
		dwatch.fwatch.c <- &node{
			name:  `.watchfs_v2_forced`,
			size:  nodeFlagForced,
			mtime: time.Now(),
		}
	}
}

// Stop watching the file and re-scanning the Root directory.
func (dwatch *DirWatcher) Stop() {
	if dwatch.fwatch != nil {
		dwatch.fwatch.Stop()
	}
}

func (dwatch *DirWatcher) indexingFile(apath string) (anode *node) {
	if dwatch.opts.isExcluded(apath) {
		return &nodeExcluded
	}
	if !dwatch.opts.isIncluded(apath) {
		return &nodeExcluded
	}
	if apath == dwatch.opts.FileWatcherOptions.File {
		return &nodeExcluded
	}
	anode, _ = newNode(apath)
	// The newNode may return nil, so we will
	// let the next re-scan to include them
	// later.
	return anode
}

// initialScan scan the directory dir and its sub-directories to get the
// list of directory and the list of included files.
func (dwatch *DirWatcher) initialScan(dir string) (changes []os.FileInfo) {
	var (
		dirq   = []string{dir}
		name   string
		apath  string
		anode  *node
		listde []os.DirEntry
		err    error
		de     os.DirEntry
	)
	for len(dirq) > 0 {
		dir = dirq[0]
		dirq = dirq[1:]

		anode, err = newNode(dir)
		if err != nil {
			continue
		}
		if dwatch.opts.isExcluded(dir) {
			dwatch.idxDir[dir] = nodeExcluded
			continue
		}

		listde, err = os.ReadDir(dir)
		if err != nil {
			continue
		}
		anode.size = int64(len(listde))
		dwatch.idxDir[dir] = *anode

		for _, de = range listde {
			name = de.Name()
			apath = filepath.Join(dir, name)
			if de.IsDir() {
				dirq = append(dirq, apath)
				continue
			}
			anode = dwatch.indexingFile(apath)
			if anode != nil {
				dwatch.idxFile[apath] = *anode
				changes = append(changes, anode)
			}
		}
	}
	return changes
}

func (dwatch *DirWatcher) scanDir() (changes []os.FileInfo) {
	// Fetch the current keys first, so in case there is new directory
	// it does not included twice.
	var keys = slices.Sorted(maps.Keys(dwatch.idxDir))
	var (
		dir     string
		anode   node
		newnode *node
		err     error
		listde  []os.DirEntry
	)
	for _, dir = range keys {
		anode = dwatch.idxDir[dir]
		if anode.size == nodeFlagExcluded {
			continue
		}

		newnode, err = newNode(dir)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				// The directory has been deleted.
				anode = node{
					name: dir,
					size: FileFlagDeleted,
					mode: os.ModeDir,
				}
				changes = append(changes, &anode)
				delete(dwatch.idxDir, dir)
			}
			continue
		}

		listde, err = os.ReadDir(dir)
		if err != nil {
			continue
		}
		newnode.size = int64(len(listde))

		if anode.equal(newnode) {
			continue
		}
		dwatch.idxDir[dir] = *newnode

		var (
			de    os.DirEntry
			name  string
			apath string
		)
		for _, de = range listde {
			name = de.Name()
			apath = filepath.Join(dir, name)

			if de.IsDir() {
				anode = dwatch.idxDir[apath]
				if anode.mtime.IsZero() {
					// New directory created.
					var newFiles = dwatch.initialScan(apath)
					changes = append(changes, newFiles...)
				}
				// anode is a directory that has been
				// indexed.
				continue
			}

			anode = dwatch.idxFile[apath]
			if anode.mtime.IsZero() {
				// New file created.
				newnode = dwatch.indexingFile(apath)
				if newnode != nil {
					dwatch.idxNewFile[apath] = *newnode
				}
				continue
			}
		}
	}
	return changes
}

func (dwatch *DirWatcher) scanFile() (fileChanges []os.FileInfo) {
	var (
		newnode *node
		err     error
	)
	for apath, anode := range dwatch.idxFile {
		if anode.size == nodeFlagExcluded {
			continue
		}
		newnode, err = newNode(apath)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				// File has been deleted.
				newnode = &node{
					name: apath,
					size: FileFlagDeleted,
				}
				fileChanges = append(fileChanges, newnode)
				delete(dwatch.idxFile, apath)
			}
			continue
		}
		if anode.equal(newnode) {
			continue
		}
		// File has been updated.
		dwatch.idxFile[apath] = *newnode
		fileChanges = append(fileChanges, newnode)
	}
	return fileChanges
}

func (dwatch *DirWatcher) watch() {
	var (
		dirChanges  []os.FileInfo
		fileChanges []os.FileInfo
	)
	for range dwatch.fwatch.C {
		// Scan new files on each directory.
		dirChanges = dwatch.scanDir()

		// Scan update or delete files.
		fileChanges = dwatch.scanFile()

		if len(dwatch.idxNewFile) != 0 {
			for apath, anode := range dwatch.idxNewFile {
				dwatch.idxFile[apath] = anode
				if anode.size == nodeFlagExcluded {
					continue
				}
				fileChanges = append(fileChanges, &anode)
			}
			clear(dwatch.idxNewFile)
		}

		dirChanges = append(dirChanges, fileChanges...)
		dwatch.c <- dirChanges
	}
	close(dwatch.c)
}
