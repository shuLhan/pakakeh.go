// Copyright 2018, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package git provide a wrapper for git command line interface.
package git

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"git.sr.ht/~shulhan/pakakeh.go/lib/ini"
)

const (
	_defRemoteName = "origin"
	_defRef        = "origin/master"
	_defBranch     = "master"
)

var (
	_stdout io.Writer = os.Stdout
	_stderr io.Writer = os.Stderr
)

// CheckoutRevision will set the HEAD to specific revision on specific branch.
// Any untracked files and directories will be removed before checking out
// existing branch or creating new branch.
// If remoteName is empty, it will use default reference "origin".
// If branch is empty, it will use default branch "master".
// If revision is empty, it will do nothing.
//
// This function assume that repository is up-to-date with remote.
// Client may call FetchAll() before, to prevent checking out revision that
// may not exist.
func CheckoutRevision(repoDir, remoteName, branch, revision string) error {
	if len(remoteName) == 0 {
		remoteName = _defRemoteName
	}
	if len(branch) == 0 {
		branch = _defBranch
	}
	if len(revision) == 0 {
		return nil
	}
	ref := remoteName + "/" + branch

	cmd := exec.Command("git")
	cmd.Args = append(cmd.Args, "clean", "-qdff")
	cmd.Dir = repoDir
	cmd.Stdout = _stdout
	cmd.Stderr = _stderr

	err := cmd.Run()
	if err != nil {
		err = fmt.Errorf("CheckoutRevision: %s", err)
		return err
	}

	cmd = exec.Command("git", "checkout")
	cmd.Args = append(cmd.Args, "--quiet")
	cmd.Args = append(cmd.Args, "--track", ref, "-B", branch)
	cmd.Dir = repoDir
	cmd.Stdout = _stdout
	cmd.Stderr = _stderr

	err = cmd.Run()
	if err != nil {
		err = fmt.Errorf("CheckoutRevision: %s", err)
		return err
	}

	cmd = exec.Command("git", "reset")
	cmd.Args = append(cmd.Args, "--quiet")
	cmd.Args = append(cmd.Args, "--hard", revision)
	cmd.Dir = repoDir
	cmd.Stdout = _stdout
	cmd.Stderr = _stderr

	err = cmd.Run()
	if err != nil {
		err = fmt.Errorf("CheckoutRevision: %s", err)
	}

	return err
}

// Clone the repository into destination directory.
//
// If destination directory is not empty it will return an error.
func Clone(remoteURL, dest string) (err error) {
	err = os.MkdirAll(dest, 0700)
	if err != nil {
		err = fmt.Errorf("Clone: %s", err)
		return
	}

	cmd := exec.Command("git", "clone")
	cmd.Args = append(cmd.Args, "--quiet")
	cmd.Args = append(cmd.Args, remoteURL, ".")
	cmd.Dir = dest
	cmd.Stdout = _stdout
	cmd.Stderr = _stderr

	err = cmd.Run()
	if err != nil {
		err = fmt.Errorf("Clone: %s", err)
	}

	return
}

// FetchAll will fetch the latest commits and tags from remote.
func FetchAll(repoDir string) (err error) {
	cmd := exec.Command("git", "fetch")
	cmd.Args = append(cmd.Args, "--quiet")
	cmd.Args = append(cmd.Args, "--all", "--tags", "--force")
	cmd.Dir = repoDir
	cmd.Stdout = _stdout
	cmd.Stderr = _stderr

	err = cmd.Run()
	if err != nil {
		err = fmt.Errorf("FetchAll: %s", err)
	}

	return
}

// FetchTags will fetch all tags from remote.
func FetchTags(repoDir string) error {
	cmd := exec.Command("git", "fetch")
	cmd.Args = append(cmd.Args, "--quiet")
	cmd.Args = append(cmd.Args, "--tags", "--force")
	cmd.Dir = repoDir
	cmd.Stdout = _stdout
	cmd.Stderr = _stderr

	err := cmd.Run()
	if err != nil {
		err = fmt.Errorf("FetchTags: %s", err)
	}

	return err
}

// GetRemoteURL return remote URL or error if repository is not git or url is
// empty.
// If remoteName is empty, it will be set to default ("origin").
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

	url, ok := gitIni.Get("remote", remoteName, "url", "")
	if !ok {
		err = fmt.Errorf("GetRemote: Empty or invalid remote name")
	}

	return
}

// GetTag get the tag from revision.  If revision is empty it's default to
// "HEAD".
func GetTag(repoDir, revision string) (tag string, err error) {
	if len(revision) == 0 {
		revision = "HEAD"
	}

	cmd := exec.Command("git")
	cmd.Args = append(cmd.Args, "describe", "--tags", "--exact-match", revision)
	cmd.Dir = repoDir
	cmd.Stderr = _stderr

	btag, err := cmd.Output()
	if err != nil {
		err = fmt.Errorf("GetTag: %s", err)
		return
	}

	tag = string(bytes.TrimSpace(btag))

	return
}

// LatestCommit get the latest commit hash in short format from "ref".
// If ref is empty, its default to "origin/master".
func LatestCommit(repoDir, ref string) (commit string, err error) {
	if len(ref) == 0 {
		ref = _defRef
	}

	cmd := exec.Command("git")
	cmd.Args = append(cmd.Args, "rev-parse", "--short", ref)
	cmd.Dir = repoDir
	cmd.Stderr = _stderr

	bcommit, err := cmd.Output()
	if err != nil {
		err = fmt.Errorf("LatestCommit: %s", err)
		return
	}

	commit = string(bytes.TrimSpace(bcommit))

	return
}

// LatestTag get latest tag.
func LatestTag(repoDir string) (tag string, err error) {
	cmd := exec.Command("git")
	cmd.Args = append(cmd.Args, "rev-list", "--tags", "--max-count=1")
	cmd.Dir = repoDir
	cmd.Stderr = _stderr

	bout, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("LatestTag: %s", err)
	}

	out := string(bytes.TrimSpace(bout))
	if len(out) == 0 {
		return "", nil
	}

	cmd = exec.Command("git")
	cmd.Args = append(cmd.Args, "describe", "--tags", "--abbrev=0", out)
	cmd.Dir = repoDir
	cmd.Stderr = _stderr

	bout, err = cmd.Output()
	if err != nil {
		return "", fmt.Errorf("LatestTag: %s", err)
	}

	tag = string(bytes.TrimSpace(bout))

	return tag, nil
}

// LatestVersion will try to get latest tag from repository.
// If it's fail get the latest commit hash.
func LatestVersion(repoDir string) (version string, err error) {
	version, err = LatestTag(repoDir)
	if err == nil && len(version) > 0 {
		return
	}

	version, err = LatestCommit(repoDir, "")
	if err == nil {
		return
	}

	err = fmt.Errorf("GetVersion: %s", err)
	return
}

// ListTags get all tags from repository.
func ListTags(repoDir string) (tags []string, err error) {
	err = FetchTags(repoDir)
	if err != nil {
		return nil, err
	}

	cmd := exec.Command("git")
	cmd.Args = append(cmd.Args, "tag", "--list")
	cmd.Dir = repoDir
	cmd.Stderr = _stderr

	bout, err := cmd.Output()
	if err != nil {
		err = fmt.Errorf("ListTag: %s", err)
		return nil, err
	}

	sep := []byte{'\n'}
	btags := bytes.Split(bout, sep)

	for x := 0; x < len(btags); x++ {
		if len(btags[x]) == 0 {
			continue
		}
		tags = append(tags, string(btags[x]))
	}

	return tags, nil
}

// LogRevisions get commits between two revisions.
func LogRevisions(repoDir, prevRevision, nextRevision string) error {
	cmd := exec.Command("git")
	cmd.Args = append(cmd.Args, "--no-pager", "log", "--oneline",
		prevRevision+"..."+nextRevision)
	cmd.Dir = repoDir
	cmd.Stdout = _stdout
	cmd.Stderr = _stderr

	err := cmd.Run()
	if err != nil {
		err = fmt.Errorf("CompareRevisions: %s", err)
	}

	return err
}

// RemoteChange change current repository remote name (e.g. "origin") to new
// remote name and URL.
func RemoteChange(repoDir, oldName, newName, newURL string) error {
	if len(repoDir) == 0 {
		return nil
	}

	cmd := exec.Command("git")
	cmd.Args = append(cmd.Args, "remote", "remove", oldName)
	cmd.Dir = repoDir
	cmd.Stdout = _stdout
	cmd.Stderr = _stderr

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

	err = cmd.Run()
	if err != nil {
		err = fmt.Errorf("RemoteChange: %s", err)
	}

	return err
}

// RemoteBranches return list of remote branches.
func RemoteBranches(repoDir string) ([]string, error) {
	if len(repoDir) == 0 {
		return nil, nil
	}

	cmd := exec.Command("git", "--no-pager", "branch", "-r", "--format", "%(refname:lstrip=3)")
	cmd.Dir = repoDir
	cmd.Stderr = _stderr

	bout, err := cmd.Output()
	if err != nil {
		err = fmt.Errorf("RemoteBranches: %s", err)
		return nil, err
	}

	bbranches := bytes.Split(bout, []byte{'\n'})
	if len(bbranches) == 0 {
		return nil, err
	}

	var branches []string
	bHEAD := []byte("HEAD")
	for x := 0; x < len(bbranches); x++ {
		if len(bbranches[x]) == 0 {
			continue
		}
		if bytes.Equal(bbranches[x], bHEAD) {
			continue
		}
		branches = append(branches, string(bbranches[x]))
	}

	return branches, nil
}
