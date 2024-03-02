// Copyright 2022, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"git.sr.ht/~shulhan/pakakeh.go/lib/ascii"
)

const (
	defDataName       = "default"
	defDataFileSuffix = "_test.txt"
)

var (
	prefixInput  = []byte(">>>")
	prefixOutput = []byte("<<<")
)

// Data contains predefined input and output values that is loaded from
// file to be used during test.
//
// The Data provides zero or more flags, an optional description, zero or
// more input, and zero or more output.
//
// The data content use the following format,
//
//	[FLAG_KEY ":" FLAG_VALUE LF]
//	[LF DESCRIPTION]
//	LF
//	">>>" [INPUT_NAME] LF
//	INPUT_CONTENT
//	LF
//	"<<<" [OUTPUT_NAME] LF
//	OUTPUT_CONTENT
//
// The Data can contains zero or more flag.
// A Flag is map of key and value separated by ":".
// The Flag's key must not contain spaces.
//
// The Data may contain description, to describe the content of test file.
//
// The line that start with "\n>>>" (new line followed by three '>') define
// the beginning of Input.
// An Input can have a name, if its empty it will be set to "default".
// An Input can be defined multiple times, with different names.
//
// The line that start with "\n<<<" (new line followed by three '<') defined
// the beginning of Output.
// An Output can have a name, if its empty it will be set to "default".
// An Output also can be defined multiple times, with different names.
//
// All of both Input and Output content will have one new line at the end, to
// separated them with each others.
// If the content of Input or Output itself expecting empty line at the end,
// add two empty lines at the end of it.
//
// # Example
//
// The following code illustrate how to use Data when writing test.
//
// Assume that we are writing a parser that consume []byte.
// First we pass the input as defined in ">>>" and then
// we dump the result into bytes.Buffer to be compare with output "<<<".
//
//	func TestParse(t *testing.T) {
//		var buf bytes.Buffer
//		tdata, _ := LoadData("testdata/data.txt")
//		opt := tdata.Flag["env"]
//		p, err := Parse(tdata.Input["default"], opt)
//		if err != nil {
//			Assert(t, "Error", tdata.Output["error"], []byte(err.Error())
//		}
//		fmt.Fprintf(&buf, "%v", p)
//		want := tdata.Output["default"]
//		got := buf.Bytes()
//		Assert(t, tdata.Name, want, got)
//	}
//
// That is the gist, the real application can consume one or more input; or
// generate one or more output.
//
// # Articles
//
//   - https://kilabit.info/journal/2023/go_test_data/
type Data struct {
	Flag   map[string]string
	Input  map[string][]byte
	Output map[string][]byte

	// The file name of the data.
	Name string

	Desc []byte
}

// LoadData load data from file.
func LoadData(file string) (data *Data, err error) {
	var (
		logp = "LoadData"

		content []byte
	)

	content, err = os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", logp, err)
	}

	data = newData(filepath.Base(file))

	err = data.parse(content)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// LoadDataDir load all data inside a directory.
// Only file that has file name suffix "_text.txt" will be loaded.
func LoadDataDir(path string) (listData []*Data, err error) {
	var (
		logp = "LoadDataDir"

		dir         *os.File
		listfi      []os.FileInfo
		fi          os.FileInfo
		data        *Data
		name        string
		pathTestTxt string
	)

	dir, err = os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", logp, err)
	}

	listfi, err = dir.Readdir(0)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", logp, err)
	}

	var listTestTxt = make([]string, 0, len(listfi))

	for _, fi = range listfi {
		if fi.Size() == 0 || fi.IsDir() {
			continue
		}

		name = fi.Name()

		if !strings.HasSuffix(name, defDataFileSuffix) {
			continue
		}

		pathTestTxt = filepath.Join(path, name)
		listTestTxt = append(listTestTxt, pathTestTxt)
	}
	if len(listTestTxt) == 0 {
		return listData, nil
	}

	sort.Strings(listTestTxt)

	for _, pathTestTxt = range listTestTxt {
		data, err = LoadData(pathTestTxt)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", logp, err)
		}

		listData = append(listData, data)
	}

	return listData, nil
}

func isFlag(content []byte) bool {
	var c byte
	for _, c = range content {
		if ascii.IsSpace(c) {
			return false
		}
		if c == ':' {
			return true
		}
	}
	return false
}

func newData(name string) (data *Data) {
	data = &Data{
		Name:   name,
		Flag:   make(map[string]string),
		Input:  make(map[string][]byte),
		Output: make(map[string][]byte),
	}
	return data
}

func (data *Data) parse(content []byte) (err error) {
	const (
		stateFlag int = iota
		stateDesc
		stateInputOutput
		stateInput
		stateOutput
	)

	var (
		logp = "LoadData"

		name  string
		lines [][]byte
		state int
		n     int
		x     int
	)

	lines = bytes.Split(content, []byte("\n"))

	for x < len(lines) {
		content = lines[x]
		if state == stateFlag {
			if len(content) == 0 {
				x++
				continue
			}
			if isFlag(content) {
				data.parseFlag(content)
				x++
				continue
			}
			state = stateDesc
		}
		if state == stateDesc {
			if len(content) == 0 {
				x++
				continue
			}
			if !(bytes.HasPrefix(content, prefixInput) || bytes.HasPrefix(content, prefixOutput)) {
				if len(data.Desc) > 0 {
					data.Desc = append(data.Desc, '\n')
				}
				data.Desc = append(data.Desc, content...)
				x++
				continue
			}
			state = stateInputOutput
		}
		if bytes.HasPrefix(content, prefixInput) {
			name, content, n = data.parseInputOutput(lines[x:])
			data.Input[name] = content
			x += n
			continue
		}
		if bytes.HasPrefix(content, prefixOutput) {
			name, content, n = data.parseInputOutput(lines[x:])
			data.Output[name] = content
			x += n
			continue
		}
		return fmt.Errorf("%s: unknown syntax line %d: %s", logp, x, content)
	}
	return nil
}

func (data *Data) parseFlag(content []byte) {
	var (
		idx  = bytes.IndexByte(content, ':')
		bkey []byte
		bval []byte
	)
	if idx < 0 {
		return
	}

	bkey = bytes.TrimSpace(content[:idx])
	if len(bkey) == 0 {
		return
	}

	bval = bytes.TrimSpace(content[idx+1:])

	data.Flag[string(bkey)] = string(bval)
}

func (data *Data) parseInputOutput(lines [][]byte) (name string, content []byte, x int) {
	var (
		line        []byte
		bname       []byte
		bufContent  bytes.Buffer
		n           int
		isPrevEmpty bool
	)

	line = lines[0]
	bname = bytes.TrimSpace(line[3:])
	if len(bname) == 0 {
		name = defDataName
	} else {
		name = string(bname)
	}

	for x = 1; x < len(lines); x++ {
		line = lines[x]
		if len(line) == 0 {
			if isPrevEmpty {
				bufContent.WriteByte('\n')
			} else {
				isPrevEmpty = true
			}
			continue
		}
		if isPrevEmpty {
			if bytes.HasPrefix(line, prefixInput) || bytes.HasPrefix(line, prefixOutput) {
				break
			}
			bufContent.WriteByte('\n')
		}
		bufContent.Write(line)
		bufContent.WriteByte('\n')
		isPrevEmpty = false
	}

	content = bufContent.Bytes()
	n = len(content) - 1
	if n > 0 && content[n] == '\n' {
		content = content[:n]
	}

	return name, content, x
}
