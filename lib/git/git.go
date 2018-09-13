// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package git provide a wrapper for git comman line interface.
package git

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/shuLhan/share/lib/debug"
	"github.com/shuLhan/share/lib/ini"
)

const (
	_defRemoteName = "origin"
	_defRef        = "origin/master"
	_defBranch     = "master"
)

var (
	_stdout = os.Stdout
	_stderr = os.Stderr
)

//
// CheckoutRevision will set the HEAD to specific revision on specific branch.
// Any untracked files and directories will be removed before checking out
// existing branch or creating new branch.
// If ref is empty, it will use default reference "origin/master".
// If branch is empty, it will use default branch "master".
// If revision is empty, it will do nothing.
//
func CheckoutRevision(repoDir, ref, branch, revision string) error {
	if len(revision) == 0 {
		return nil
	}

	cmd := exec.Command("git")
	cmd.Args = append(cmd.Args, "clean", "-qdff")
	cmd.Dir = repoDir
	cmd.Stdout = _stdout
	cmd.Stderr = _stderr
	if debug.Value >= 1 {
		fmt.Printf("= CheckoutRevision %s %s\n", cmd.Dir, cmd.Args)
	}
	err := cmd.Run()
	if err != nil {
		err = fmt.Errorf("CheckoutRevision: %s", err)
		return err
	}

	err = FetchAll(repoDir)
	if err != nil {
		err = fmt.Errorf("CheckoutRevision: %s", err)
		return err
	}

	cmd = exec.Command("git")
	cmd.Dir = repoDir
	cmd.Stdout = _stdout
	cmd.Stderr = _stderr

	if len(ref) == 0 {
		ref = _defRef
	}
	cmd.Args = append(cmd.Args, "checkout", "--quiet", "--track", ref)

	if len(branch) == 0 {
		branch = _defBranch
	}
	cmd.Args = append(cmd.Args, "-B", branch)

	if debug.Value >= 1 {
		fmt.Printf("= CheckoutRevision %s %s\n", cmd.Dir, cmd.Args)
	}
	err = cmd.Run()
	if err != nil {
		err = fmt.Errorf("CheckoutRevision: %s", err)
		return err
	}

	cmd = exec.Command("git")
	cmd.Dir = repoDir
	cmd.Stdout = _stdout
	cmd.Stderr = _stderr

	cmd.Args = append(cmd.Args, "reset", "--quiet", "--hard", revision)
	if debug.Value >= 1 {
		fmt.Printf("= CheckoutRevision %s %s\n", cmd.Dir, cmd.Args)
	}
	err = cmd.Run()
	if err != nil {
		err = fmt.Errorf("CheckoutRevision: %s", err)
	}

	return err
}

//
// Clone the repository into destination directory.
//
// If destination directory is not empty it will return an error.
//
func Clone(remoteURL, dest string) (err error) {
	err = os.MkdirAll(dest, 0700)
	if err != nil {
		err = fmt.Errorf("Clone: %s", err)
		return
	}

	cmd := exec.Command("git")
	cmd.Args = append(cmd.Args, "clone", "--quiet", remoteURL, ".")
	cmd.Dir = dest
	cmd.Stdout = _stdout
	cmd.Stderr = _stderr

	if debug.Value >= 1 {
		fmt.Printf("= Clone %s %s %s\n", remoteURL, cmd.Dir, cmd.Args)
	}

	err = cmd.Run()
	if err != nil {
		err = fmt.Errorf("Clone: %s", err)
	}

	return
}

//
// FetchAll will fetch the latest commits from remote.
//
func FetchAll(repoDir string) error {
	cmd := exec.Command("git")
	cmd.Args = append(cmd.Args, "fetch", "--quiet", "--all")
	cmd.Dir = repoDir
	cmd.Stdout = _stdout
	cmd.Stderr = _stderr

	if debug.Value >= 1 {
		fmt.Printf("= FetchAll %s %s\n", cmd.Dir, cmd.Args)
	}

	err := cmd.Run()
	if err != nil {
		err = fmt.Errorf("FetchAll: %s", err)
	}

	return err
}

//
// GetRemoteURL return remote URL or error if repository is not git or url is
// empty.
// If remoteName is empty, it will be set to default ("origin").
//
func GetRemoteURL(repoDir, remoteName string) (url string, err error) {
	if len(remoteName) == 0 {
		remoteName = _defRemoteName
	}

	gitConfig := filepath.Join(repoDir, ".git", "config")

	gitIni, err := ini.Open(gitConfig)
	if err != nil {
		err = fmt.Errorf("GetRemote: %s", err)
		return
	}

	url, ok := gitIni.Get("remote", remoteName, "url")
	if !ok {
		err = fmt.Errorf("GetRemote: Empty or invalid remote name")
	}

	return
}

//
// GetTag get the tag from revision.  If revision is empty it's default to
// "HEAD".
//
func GetTag(repoDir, revision string) (tag string, err error) {
	if len(revision) == 0 {
		revision = "HEAD"
	}

	cmd := exec.Command("git")
	cmd.Args = append(cmd.Args, "describe", "--tags", "--exact-match", revision)
	cmd.Dir = repoDir
	cmd.Stderr = _stderr

	if debug.Value >= 1 {
		fmt.Printf("= GetTag %s %s\n", cmd.Dir, cmd.Args)
	}

	btag, err := cmd.Output()
	if err != nil {
		err = fmt.Errorf("GetTag: %s", err)
		return
	}

	tag = string(bytes.TrimSpace(btag))

	return
}

//
// LatestCommit get the latest commit hash in short format from "ref".
// If ref is empty, its default to "origin/master".
//
func LatestCommit(repoDir, ref string) (commit string, err error) {
	if len(ref) == 0 {
		ref = _defRef
	}

	cmd := exec.Command("git")
	cmd.Args = append(cmd.Args, "rev-parse", "--short", ref)
	cmd.Dir = repoDir
	cmd.Stderr = _stderr

	if debug.Value >= 1 {
		fmt.Printf("= LatestCommit %s %s\n", cmd.Dir, cmd.Args)
	}

	bcommit, err := cmd.Output()
	if err != nil {
		err = fmt.Errorf("LatestCommit: %s", err)
		return
	}

	commit = string(bytes.TrimSpace(bcommit))

	return
}

//
// LatestTag get latest tag.
//
func LatestTag(repoDir string) (tag string, err error) {
	cmd := exec.Command("git")
	cmd.Args = append(cmd.Args, "rev-list", "--tags", "--max-count=1")
	cmd.Dir = repoDir
	cmd.Stderr = _stderr

	if debug.Value >= 1 {
		fmt.Printf("= LatestTag %s %s\n", cmd.Dir, cmd.Args)
	}

	bout, err := cmd.Output()
	if err != nil {
		err = fmt.Errorf("LatestTag: %s", err)
		return
	}

	out := string(bytes.TrimSpace(bout))

	cmd = exec.Command("git")
	cmd.Args = append(cmd.Args, "describe", "--tags", "--abbrev=0", out)
	cmd.Dir = repoDir
	cmd.Stderr = _stderr

	if debug.Value >= 1 {
		fmt.Printf("= LatestTag %s %s\n", cmd.Dir, cmd.Args)
	}

	bout, err = cmd.Output()
	if err != nil {
		err = fmt.Errorf("LatestTag: %s", err)
		return
	}

	tag = string(bytes.TrimSpace(bout))

	return
}

//
// LatestVersion will try to get latest tag from repository.
// If it's fail get the latest commit hash.
//
func LatestVersion(repoDir string) (version string, err error) {
	version, err = LatestTag(repoDir)
	if err == nil {
		return
	}

	version, err = LatestCommit(repoDir, "")
	if err == nil {
		return
	}

	err = fmt.Errorf("GetVersion: %s", err)
	return
}

//
// LogRevisions get commits between two revisions.
//
func LogRevisions(repoDir, prevRevision, nextRevision string) error {
	cmd := exec.Command("git")
	cmd.Args = append(cmd.Args, "--no-pager", "log", "--oneline",
		prevRevision+"..."+nextRevision)
	cmd.Dir = repoDir
	cmd.Stdout = _stdout
	cmd.Stderr = _stderr

	if debug.Value >= 1 {
		fmt.Printf("= CompareRevisions %s %s\n", cmd.Dir, cmd.Args)
	}

	err := cmd.Run()
	if err != nil {
		err = fmt.Errorf("CompareRevisions: %s", err)
	}

	return err
}

//
// RemoteChange change current repository remote name (e.g. "origin") to new
// remote name and URL.
//
func RemoteChange(repoDir, oldName, newName, newURL string) error {
	if len(repoDir) == 0 {
		return nil
	}

	cmd := exec.Command("git")
	cmd.Args = append(cmd.Args, "remote", "remove", oldName)
	cmd.Dir = repoDir
	cmd.Stdout = _stdout
	cmd.Stderr = _stderr

	if debug.Value >= 1 {
		fmt.Printf("= RemoteChange %s %s\n", cmd.Dir, cmd.Args)
	}

	err := cmd.Run()
	if err != nil {
		err = fmt.Errorf("RemoteChange: %s", err)
		return err
	}

	cmd = exec.Command("git")
	cmd.Args = append(cmd.Args, "remote", "add", newName, newURL)
	cmd.Dir = repoDir
	cmd.Stdout = _stdout
	cmd.Stderr = _stderr

	if debug.Value >= 1 {
		fmt.Printf("= RemoteChange %s %s\n", cmd.Dir, cmd.Args)
	}

	err = cmd.Run()
	if err != nil {
		err = fmt.Errorf("RemoteChange: %s", err)
	}

	return err
}
