// SPDX-License-Identifier: BSD-3-Clause
// SPDX-FileCopyrightText: 2018 M. Shulhan <ms@kilabit.info>

// Package git provide a wrapper for git command line interface.
package git

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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

// Git is a type for working with single git repository.
type Git struct {
	listIgnore map[string]*Gitignore

	// absDir define the absolute path to the repository.
	absDir string
}

// New start working on repository in directory `dir`.
// If the `dir` does not contains ".git" directory it will return an error,
// even if the parent directory may contains the ".git" directory.
func New(dir string) (git *Git, err error) {
	var logp = `New`
	var fi os.FileInfo
	var pathDotGit = filepath.Join(dir, `.git`)
	fi, err = os.Stat(pathDotGit)
	if err != nil {
		return nil, fmt.Errorf(`%s: %q is not a git repository`, logp, dir)
	}
	if !fi.IsDir() {
		return nil, fmt.Errorf(`%s: %q is not a git repository`, logp, dir)
	}
	dir, err = filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf(`%s: %w`, logp, err)
	}
	git = &Git{
		absDir:     dir,
		listIgnore: map[string]*Gitignore{},
	}
	return git, nil
}

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
		return fmt.Errorf(`CheckoutRevision: %w`, err)
	}

	cmd = exec.Command("git", "checkout")
	cmd.Args = append(cmd.Args, "--quiet")
	cmd.Args = append(cmd.Args, "--track", ref, "-B", branch)
	cmd.Dir = repoDir
	cmd.Stdout = _stdout
	cmd.Stderr = _stderr

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf(`CheckoutRevision: %w`, err)
	}

	cmd = exec.Command("git", "reset")
	cmd.Args = append(cmd.Args, "--quiet")
	cmd.Args = append(cmd.Args, "--hard", revision)
	cmd.Dir = repoDir
	cmd.Stdout = _stdout
	cmd.Stderr = _stderr

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf(`CheckoutRevision: %w`, err)
	}

	return nil
}

// Clone the repository into destination directory.
//
// If destination directory is not empty it will return an error.
func Clone(remoteURL, dest string) (err error) {
	err = os.MkdirAll(dest, 0700)
	if err != nil {
		return fmt.Errorf(`Clone: %w`, err)
	}

	cmd := exec.Command("git", "clone")
	cmd.Args = append(cmd.Args, "--quiet")
	cmd.Args = append(cmd.Args, remoteURL, ".")
	cmd.Dir = dest
	cmd.Stdout = _stdout
	cmd.Stderr = _stderr

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf(`Clone: %w`, err)
	}
	return nil
}

// Equal return true if `v` is instance of `*Git` and has the same working
// directory.
// This implement [git.sr.ht/~shulhan/pakakeh.go/lib/reflect.Equaler]
// interface.
func (git *Git) Equal(v any) bool {
	if v == nil {
		return false
	}
	got, ok := v.(*Git)
	if !ok {
		return false
	}
	if got == nil {
		return false
	}
	return git.String() == got.String()
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
		return fmt.Errorf(`FetchAll: %w`, err)
	}
	return nil
}

// FetchTags will fetch all tags from remote.
func FetchTags(repoDir string) (err error) {
	cmd := exec.Command("git", "fetch")
	cmd.Args = append(cmd.Args, "--quiet")
	cmd.Args = append(cmd.Args, "--tags", "--force")
	cmd.Dir = repoDir
	cmd.Stdout = _stdout
	cmd.Stderr = _stderr

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf(`FetchTags: %w`, err)
	}
	return nil
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
		return ``, fmt.Errorf(`GetRemote: %w`, err)
	}

	url, ok := gitIni.Get("remote", remoteName, "url", "")
	if !ok {
		return ``, errors.New(`GetRemote: Empty or invalid remote name`)
	}

	return url, nil
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
		return ``, fmt.Errorf(`GetTag: %w`, err)
	}

	tag = string(bytes.TrimSpace(btag))

	return tag, nil
}

// IsIgnored return true if the `path` is ignored by git.
// This is processed by matching it with all of the patterns in the
// ".gitignore" file inside the path directory and its parent, until the root
// of Git repository.
func (git *Git) IsIgnored(path string) (b bool) {
	path = strings.TrimSpace(path) // a/b/b1
	if path == `` {
		return true
	}
	// Traverse each directory from bottom to the top of git directory to
	// load ".gitignore" file and match it with path.
	var absPath = filepath.Join(git.absDir, path)        // $git/a/b/b1
	var dirGitignore = filepath.Dir(absPath)             // $git/a/b/
	var name = strings.TrimPrefix(absPath, dirGitignore) // b1
	name = strings.TrimLeft(name, `/`)
	for strings.HasPrefix(dirGitignore, git.absDir) {
		var ign *Gitignore
		var ok bool
		ign, ok = git.listIgnore[dirGitignore]
		if !ok {
			ign, _ = LoadGitignore(dirGitignore)
			git.listIgnore[dirGitignore] = ign
		}
		if ign != nil && ign.IsIgnored(name) {
			return true
		}
		dirGitignore = filepath.Dir(dirGitignore)
		name = strings.TrimPrefix(absPath, dirGitignore)
		name = strings.TrimLeft(name, `/`)
	}
	return false
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
		return ``, fmt.Errorf(`LatestCommit: %w`, err)
	}

	commit = string(bytes.TrimSpace(bcommit))

	return commit, nil
}

// LatestTag get latest tag.
func LatestTag(repoDir string) (tag string, err error) {
	cmd := exec.Command("git")
	cmd.Args = append(cmd.Args, "rev-list", "--tags", "--max-count=1")
	cmd.Dir = repoDir
	cmd.Stderr = _stderr

	bout, err := cmd.Output()
	if err != nil {
		return ``, fmt.Errorf(`LatestTag: %w`, err)
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
		return ``, fmt.Errorf(`LatestTag: %w`, err)
	}

	tag = string(bytes.TrimSpace(bout))

	return tag, nil
}

// LatestVersion will try to get latest tag from repository.
// If it's fail get the latest commit hash.
func LatestVersion(repoDir string) (version string, err error) {
	var logp = `LatestVersion`

	version, err = LatestTag(repoDir)
	if err == nil && len(version) > 0 {
		return version, nil
	}

	version, err = LatestCommit(repoDir, "")
	if err != nil {
		return ``, fmt.Errorf(`%s: %w`, logp, err)
	}

	return version, nil
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
		return nil, fmt.Errorf(`ListTag: %w`, err)
	}

	sep := []byte{'\n'}
	btags := bytes.Split(bout, sep)

	for x := range len(btags) {
		if len(btags[x]) == 0 {
			continue
		}
		tags = append(tags, string(btags[x]))
	}

	return tags, nil
}

// LogRevisions get commits between two revisions.
func LogRevisions(repoDir, prevRevision, nextRevision string) (err error) {
	cmd := exec.Command("git")
	cmd.Args = append(cmd.Args, "--no-pager", "log", "--oneline",
		prevRevision+"..."+nextRevision)
	cmd.Dir = repoDir
	cmd.Stdout = _stdout
	cmd.Stderr = _stderr

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf(`CompareRevisions: %w`, err)
	}

	return nil
}

// RemoteChange change current repository remote name (e.g. "origin") to new
// remote name and URL.
func RemoteChange(repoDir, oldName, newName, newURL string) (err error) {
	if len(repoDir) == 0 {
		return nil
	}

	cmd := exec.Command("git")
	cmd.Args = append(cmd.Args, "remote", "remove", oldName)
	cmd.Dir = repoDir
	cmd.Stdout = _stdout
	cmd.Stderr = _stderr

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf(`RemoteChange: %w`, err)
	}

	cmd = exec.Command("git")
	cmd.Args = append(cmd.Args, "remote", "add", newName, newURL)
	cmd.Dir = repoDir
	cmd.Stdout = _stdout
	cmd.Stderr = _stderr

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf(`RemoteChange: %w`, err)
	}

	return nil
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
		return nil, fmt.Errorf(`RemoteBranches: %w`, err)
	}

	bbranches := bytes.Split(bout, []byte{'\n'})
	if len(bbranches) == 0 {
		return nil, nil
	}

	var branches []string
	bHEAD := []byte("HEAD")
	for x := range len(bbranches) {
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

// String return the working directory with prefix "git+file://".
func (git *Git) String() string {
	return `git+file://` + git.absDir
}
