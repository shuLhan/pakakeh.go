// Copyright 2019, Shulhan <m.shulhan@gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package hunspell

//
// Spell contains list of options, root words, expanded words, and affixes.
//
type Spell struct {
	opts affixOptions
	dict dictionary
}

//
// New create and initialize default Spell.
//
func New() (spell *Spell) {
	spell = &Spell{
		opts: affixOptions{
			encoding:    DefaultEncoding,
			flag:        DefaultFlag,
			compoundMin: defaultMinimumCompound,
			prefixes:    make(map[string]*affix),
			suffixes:    make(map[string]*affix),
		},
		dict: dictionary{
			stems: make(map[string]*Stem),
		},
	}
	return spell
}

//
// Open create and initialize new Spell from affix and dictionary files.
//
func Open(affpath, dpath string) (spell *Spell, err error) {
	spell = New()

	if len(affpath) > 0 {
		err = spell.opts.open(affpath)
		if err != nil {
			return nil, err
		}
	}

	if len(dpath) > 0 {
		err = spell.AddDictionary(dpath)
		if err != nil {
			return nil, err
		}
	}

	return spell, nil
}

//
// AddDictionary from file "path".
//
func (spell *Spell) AddDictionary(path string) (err error) {
	return spell.dict.open(path, &spell.opts)
}

//
// Spell return the stem of "word" if its recognized by Spell;
// otherwise it will return nil.
//
func (spell *Spell) Spell(word string) (stem *Stem) {
	return spell.dict.stems[word]
}
