// Copyright 2020, Shulhan <m.shulhan@gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package hunspell

import (
	"fmt"
	"io/ioutil"
	"log"
	"strconv"

	"github.com/shuLhan/share/lib/parser"
)

type dictionary struct {
	// stems contains mapping between root words and its attributes.
	stems map[string]*stem

	// derivatives contains the mapping of combination of derivative
	// word (root word plus prefix and/or suffix) and its root word.
	derivatives map[string]*stem
}

func (dict *dictionary) open(file string, opts *affixOptions) (err error) {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return fmt.Errorf("dictionary.open: %w", err)
	}

	err = dict.load(string(content), opts)
	if err != nil {
		return err
	}

	return nil
}

//
// load dictionary from string.
//
func (dict *dictionary) load(content string, opts *affixOptions) (err error) {
	p := parser.New(content, "")

	// The string splitted into lines and then parsed one by one.
	lines := p.Lines()
	if len(lines) == 0 {
		return fmt.Errorf("empty file")
	}

	// The first line is approximately number of words.
	// The idea is to allow the parser to allocated hash map before
	// parsing all lines.
	_, err = strconv.Atoi(lines[0])
	if err != nil {
		return fmt.Errorf("invalid words count %q", lines[0])
	}

	for x := 1; x < len(lines); x++ {
		s, err := newStem(lines[x])
		if err != nil {
			return fmt.Errorf("line %d: %s", x, err.Error())
		}
		if s == nil {
			continue
		}

		_, ok := dict.stems[s.value]
		if ok {
			log.Printf("duplicate stem %q", s.value)
		}

		derivatives, err := s.unpack(opts)
		if err != nil {
			return fmt.Errorf("line %d: %s", x, err.Error())
		}

		dict.stems[s.value] = s

		for _, w := range derivatives {
			dict.derivatives[w] = s
		}
	}

	return nil
}
