// Copyright 2019, Shulhan <m.shulhan@gmail.com>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package hunspell

import (
	"fmt"
	"io/ioutil"
	"log"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/shuLhan/share/lib/parser"
	libstrings "github.com/shuLhan/share/lib/strings"
)

//
// Spell contains list of options, root words, expanded words, and affixes.
//
type Spell struct {
	//
	// Affix file general options.
	//

	//
	// Set character encodings of words and morphemes in affix and
	// dictionary files.
	// Its value is derived from SET option in affix file.
	// Possible values: UTF-8, ISO8859-1 ... ISO8859-10,
	// ISO8859-13 ... IS8859-15, KOI8-R, KOI8-U, cp1251, or
	// ISCII-DEVANAGRI.
	//
	encoding string

	//
	// Flag type.
	// Default type  is  the  extended  ASCII  (8-bit) character.
	//
	flag string

	//
	// Set twofold prefix stripping (but single suffix stripping) eg.
	// for morphologically complex languages with right-to-left writing
	// system.
	//
	isComplexPrefixes bool

	//
	// Set language code for language-specific functions of Hunspell.
	//
	// Use it to activate special casing of Azeri (LANG az), Turkish
	// (LANG tr) and Crimean Tatar (LANG crh), also not generalized
	// syllable-counting compounding rules of Hungarian (LANG hu).
	//
	lang string

	//
	// Sets characters to ignore dictionary words, affixes, and input
	// words.
	//
	ignore string

	//
	// afAliases contains list of affix that can be substituted with
	// number.  In this case, the number is index of slice.
	//
	afAliases []string

	//
	// amAliases contains list of affix rules that can be replaced with
	// ordinal number.
	//
	amAliases []string

	//
	// Affix file options for suggestion
	//

	keys []string

	// Hunspell can suggest right word forms, when they differ from the
	// bad input word by one TRY character. The parameter of TRY is case
	// sensitive.
	try string

	noSuggest          []string
	maxCompundSuggests int
	maxNGramSuggests   int
	maxDiff            int
	isOnlyMaxDiff      bool
	isNoSplitSugs      bool
	isSugsWithDots     bool
	reps               []replacement
	charsMaps          []charsmap
	warn               string
	isForbidWarn       bool

	//
	// Options for compounding.
	//

	breakOpts              []breakopt
	compoundRules          []compoundRule
	compoundMin            int
	compoundFlag           string
	compoundBegin          string
	compoundLast           string
	compoundMiddle         string
	onlyInCompound         string
	compoundPermitFlag     string
	compoundForbidFlag     string
	isCompoundMoreSuffixes bool
	compoundRoot           string
	compoundWordMax        int
	isCheckCompoundDup     bool
	isCheckCompoundRep     bool
	isCheckCompoundCase    bool
	isCheckCompoundTriple  bool
	isSimplifiedTriple     bool
	compoundPatterns       []compoundPattern
	forceUCase             string
	compoundSyllable       *compoundSyllable
	syllableNum            string

	//
	// Affix file options for affix creation.
	//

	prefixes map[string]*affix
	suffixes map[string]*affix

	//
	// Affix file other options.
	//

	circumfix     string
	forbiddenWord string
	isFullStrip   bool
	keepCase      string
	iconv         []convertion
	oconv         []convertion
	lemmaPresent  string
	needAffix     string
	pseudoRoot    string
	substandard   string
	wordchars     string
	isCheckSharps bool

	// stems contains mapping between root words and its attributes.
	stems map[string]*stem

	// derivatives contains the mapping of combination of derivative
	// word (root word plus prefix and/or suffix) and its root word.
	derivatives map[string]*stem
}

//
// New create and initialize default Spell.
//
func New() (spell *Spell) {
	spell = &Spell{
		encoding:    DefaultEncoding,
		flag:        DefaultFlag,
		compoundMin: 3,
		prefixes:    make(map[string]*affix),
		suffixes:    make(map[string]*affix),
		stems:       make(map[string]*stem),
		derivatives: make(map[string]*stem),
	}
	return spell
}

//
// Open create and initialize new Spell from affix and dictionary files.
//
func Open(affpath, dpath string) (spell *Spell, err error) {
	spell = New()

	if len(affpath) > 0 {
		err = spell.openAffix(affpath)
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
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("AddDictionary: %s", err.Error())
	}

	err = spell.loadDictionary(string(content))
	if err != nil {
		return fmt.Errorf("AddDictionary: %s", err.Error())
	}

	return nil
}

//
// Spell return the root word of "s" if its recognized by spell checked;
// otherwise it will return empty string.
//
func (spell *Spell) Spell(word string) (root string, ok bool) {
	s, ok := spell.stems[word]
	if ok {
		return s.value, true
	}

	s, ok = spell.derivatives[word]
	if ok {
		return s.value, true
	}

	return "", false
}

//
// isValidFlag check whether the flag value conform the FLAG type.
//
func (spell *Spell) isValidFlag(flag string) error {
	switch spell.flag {
	case FlagASCII:
		if len(flag) != 1 {
			return fmt.Errorf("invalid ASCII flag: %q", flag)
		}
	case FlagUTF8:
		if utf8.RuneCountInString(flag) != 1 {
			return fmt.Errorf("invalid UTF-8 flag: %q", flag)
		}
		r, _ := utf8.DecodeRuneInString(flag)
		if r == utf8.RuneError {
			return fmt.Errorf("invalid UTF-8 flag: %q", flag)
		}
	case FlagLong:
		if len(flag) != 2 {
			return fmt.Errorf("invalid long flag: %q", flag)
		}
	case FlagNum:
		_, err := strconv.Atoi(flag)
		if err != nil {
			return fmt.Errorf("invalud num flag: %q: %s", flag, err.Error())
		}
	}
	return nil
}

//
// load affix and dictionary from string.
//
func (spell *Spell) load(affContent, dictContent string) (err error) {
	err = spell.loadAffix(affContent)
	if err != nil {
		return fmt.Errorf("Load: " + err.Error())
	}
	err = spell.loadDictionary(dictContent)
	if err != nil {
		return fmt.Errorf("Load: " + err.Error())
	}
	return nil
}

//
// loadAffix options from string.
//
func (spell *Spell) loadAffix(content string) (err error) {
	p := parser.New(content, "")

	lines := p.Lines()

	for x := 0; x < len(lines); x++ {
		// Skip comment, a line starting with '#'.
		if lines[x][0] == '#' {
			continue
		}

		tokens := libstrings.Split(lines[x], false, false)

		opt := strings.ToUpper(tokens[0])

		switch opt {
		case optSet:
			spell.parseSet(tokens[1:])

		case optFlag:
			if len(tokens) != 2 {
				return fmt.Errorf("line %d: FLAG: missing argument", x)
			}

			spell.parseFlag(tokens[1])

		case optComplexPrefixes:
			spell.isComplexPrefixes = true

		case optLang:
			if len(tokens) == 1 {
				return fmt.Errorf("line %d: LANG: missing argument", x)
			}
			spell.lang = strings.ToLower(tokens[1])

		case optIgnore:
			if len(tokens) == 1 {
				return fmt.Errorf("line %d: IGNORE: missing argument", x)
			}
			spell.ignore = tokens[1]

		case optAF:
			if len(tokens) == 1 {
				return fmt.Errorf("line %d: AF: missing argument", x)
			}
			err = spell.parseAF(tokens[1])
			if err != nil {
				return fmt.Errorf("line %d: %s", x, err.Error())
			}

		case optAM:
			if len(tokens) == 1 {
				return fmt.Errorf("line %d: AM: missing argument", x)
			}
			err = spell.parseAM(tokens[1])
			if err != nil {
				return fmt.Errorf("line %d: %s", x, err.Error())
			}

		case optKey:
			if len(tokens) == 1 {
				return fmt.Errorf("line %d: KEY: missing argument", x)
			}
			spell.keys = strings.Split(tokens[1], "|")

		case optTry:
			if len(tokens) == 1 {
				return fmt.Errorf("line %d: TRY: missing argument", x)
			}
			spell.try = tokens[1]

		case optNoSuggest:
			if len(tokens) == 1 {
				return fmt.Errorf("line %d: NOSUGGEST: missing argument", x)
			}
			spell.noSuggest = append(spell.noSuggest, tokens[1])

		case optMaxCPDSuggest:
			if len(tokens) == 1 {
				return fmt.Errorf("line %d: missing MAXCPDSUGS argument", x)
			}
			spell.maxCompundSuggests, err = strconv.Atoi(tokens[1])
			if err != nil {
				return fmt.Errorf("line %d: %s", x, err.Error())
			}

		case optMaxNGramSugs:
			if len(tokens) == 1 {
				return fmt.Errorf("line %d: MAXNGRAMSUGS: missing argument", x)
			}
			spell.maxNGramSuggests, err = strconv.Atoi(tokens[1])
			if err != nil {
				return fmt.Errorf("line %d: %s", x, err.Error())
			}

		case optMaxDiff:
			if len(tokens) == 1 {
				return fmt.Errorf("line %d: MAXNGRAMSUGS: missing argument", x)
			}
			spell.maxDiff, err = strconv.Atoi(tokens[1])
			if err != nil {
				return fmt.Errorf("line %d: %s", x, err.Error())
			}
			if spell.maxDiff < 1 || spell.maxDiff > 10 {
				spell.maxDiff = 5
			}

		case optOnlyMaxDiff:
			spell.isOnlyMaxDiff = true
		case optNoSplitSugs:
			spell.isNoSplitSugs = true

		case optSugsWithDots:
			spell.isSugsWithDots = true

		case optRep:
			err = spell.parseRep(tokens[1:])
			if err != nil {
				return fmt.Errorf("line %d: %s", x, err.Error())
			}

		case optMap:
			if len(tokens) != 2 {
				return fmt.Errorf("line %d: MAP: missing argument", x)
			}
			err = spell.parseMap(tokens[1])
			if err != nil {
				return fmt.Errorf("line %d: %s", x, err.Error())
			}

		case optPhone:
			// TODO

		case optWarn:
			if len(tokens) != 2 {
				return fmt.Errorf("line %d: WARN: missing argument", x)
			}
			spell.warn = tokens[1]

		case optForbidWarn:
			spell.isForbidWarn = true

		case optBreak:
			if len(tokens) != 2 {
				return fmt.Errorf("line %d: BREAK: missing argument", x)
			}
			err = spell.parseBreak(tokens[1])
			if err != nil {
				return fmt.Errorf("line %d: %s", x, err.Error())
			}

		case optCompoundRule:
			if len(tokens) != 2 {
				return fmt.Errorf("line %d:, COMPOUNDRULE: missing argument", x)
			}
			err = spell.parseCompoundRule(tokens[1])
			if err != nil {
				return fmt.Errorf("line %d: %s", x, err.Error())
			}

		case optCompoundMin:
			if len(tokens) != 2 {
				return fmt.Errorf("line %d: COMPOUNDMIN: missing argument", x)
			}
			spell.compoundMin, err = strconv.Atoi(tokens[1])
			if err != nil {
				return fmt.Errorf("line %d: COMPOUNDMIN: invalid argument %q", x, tokens[1])
			}

		case optCompoundFlag:
			if len(tokens) != 2 {
				return fmt.Errorf("line %d: COMPOUNDFLAG: missing argument", x)
			}
			spell.compoundFlag = tokens[1]

		case optCompoundBegin:
			if len(tokens) != 2 {
				return fmt.Errorf("line %d: COMPOUNDBEGIN: missing argument", x)
			}
			spell.compoundBegin = tokens[1]

		case optCompoundLast:
			if len(tokens) != 2 {
				return fmt.Errorf("line %d: COMPOUNDLAST: missing argument", x)
			}
			spell.compoundLast = tokens[1]

		case optCompoundMiddle:
			if len(tokens) != 2 {
				return fmt.Errorf("line %d: COMPOUNDMIDDLE: missing argument", x)
			}
			spell.compoundMiddle = tokens[1]

		case optOnlyInCompound:
			if len(tokens) != 2 {
				return fmt.Errorf("line %d: ONLYINCOMPOUND: missing argument", x)
			}
			spell.onlyInCompound = tokens[1]

		case optCompoundPermitFlag:
			if len(tokens) != 2 {
				return fmt.Errorf("line %d: COMPOUNDPERMITFLAG: missing argument", x)
			}
			spell.compoundPermitFlag = tokens[1]

		case optCompoundForbidFlags:
			if len(tokens) != 2 {
				return fmt.Errorf("line %d: COMPOUNDFORBIDFLAG: missing argument", x)
			}
			spell.compoundForbidFlag = tokens[1]

		case optCompoundMoreSuffixes:
			spell.isCompoundMoreSuffixes = true

		case optCompoundRoot:
			if len(tokens) != 2 {
				return fmt.Errorf("line %d: COMPOUNDROOT: missing argument", x)
			}
			spell.compoundRoot = tokens[1]

		case optCompoundWordMax:
			if len(tokens) != 2 {
				return fmt.Errorf("line %d: COMPOUNDWORDMAX: missing argument", x)
			}
			spell.compoundWordMax, err = strconv.Atoi(tokens[1])
			if err != nil {
				return fmt.Errorf("line %d: COMPOUNDWORDMAX: invalid argument %q", x, tokens[1])
			}

		case optCheckCompoundDup:
			spell.isCheckCompoundDup = true
		case optCheckCompoundRep:
			spell.isCheckCompoundRep = true
		case optCheckCompoundCase:
			spell.isCheckCompoundCase = true
		case optCheckCompoundTriple:
			spell.isCheckCompoundTriple = true
		case optSimplifiedTriple:
			spell.isSimplifiedTriple = true

		case optCheckCompoundPattern:
			if len(tokens) == 1 {
				return fmt.Errorf("line %d: CHECKCOMPOUNDPATTERN: missing argument", x)
			}
			err = spell.parseCheckCompoundPattern(tokens)
			if err != nil {
				return fmt.Errorf("line %d: %s", x, err.Error())
			}

		case optForceUcase:
			if len(tokens) != 2 {
				return fmt.Errorf("line %d: FORCEUCASE: missing argument", x)
			}
			spell.forceUCase = tokens[1]

		case optCompoundSyllable:
			if len(tokens) != 3 {
				return fmt.Errorf("line %d: COMPOUNDSYLLABLE: missing argument", x)
			}
			cs := &compoundSyllable{}

			cs.max, err = strconv.Atoi(tokens[1])
			if err != nil {
				return fmt.Errorf("line %d: COMPOUNDSYLLABLE: invalid argument %q: %s", x, tokens[1], err.Error())
			}
			cs.vowels = tokens[2]
			spell.compoundSyllable = cs

		case optSyllableNum:
			if len(tokens) != 2 {
				return fmt.Errorf("line %d: SYLLABLENUM: missing argument", x)
			}
			spell.syllableNum = tokens[1]

		case optPFX:
			if len(tokens) < 3 {
				return fmt.Errorf("line %d: PFX: missing arguments", x)
			}
			err = spell.parsePfx(tokens[1:])
			if err != nil {
				return fmt.Errorf("line %d: %s", x, err.Error())
			}

		case optSFX:
			if len(tokens) < 3 {
				return fmt.Errorf("line %d: SFX: missing arguments", x)
			}
			err = spell.parseSfx(tokens[1:])
			if err != nil {
				return fmt.Errorf("line %d: %s", x, err.Error())
			}

		case optCircumfix:
			if len(tokens) != 2 {
				return fmt.Errorf("line %d: CIRCUMFIX: missing argument", x)
			}
			spell.circumfix = tokens[1]

		case optForbiddenWord:
			if len(tokens) != 2 {
				return fmt.Errorf("line %d: FORBIDDENWORD: missing argument", x)
			}
			spell.forbiddenWord = tokens[1]

		case optFullStrip:
			spell.isFullStrip = true

		case optKeepCase:
			if len(tokens) != 2 {
				return fmt.Errorf("line %d: KEEPCASE: missing argument", x)
			}
			spell.keepCase = tokens[1]

		case optIconv:
			if len(tokens) == 1 {
				return fmt.Errorf("line %d: ICONV: missing argument", x)
			}
			err = spell.parseIconv(tokens[1:])
			if err != nil {
				return fmt.Errorf("line %d: %s", x, err.Error())
			}

		case optOconv:
			if len(tokens) != 3 {
				return fmt.Errorf("line %d: OCONV: missing argument", x)
			}
			err = spell.parseOconv(tokens[1:])
			if err != nil {
				return fmt.Errorf("line %d: %s", x, err.Error())
			}

		case optLemmaPresent:
			if len(tokens) != 2 {
				return fmt.Errorf("line %d: LEMMA_PRESENT: missing argument", x)
			}
			spell.lemmaPresent = tokens[1]

		case optNeedAffix:
			if len(tokens) != 2 {
				return fmt.Errorf("line %d: NEEDAFFIX: missing argument", x)
			}
			spell.needAffix = tokens[1]

		case optPseudoRoot:
			if len(tokens) != 2 {
				return fmt.Errorf("line %d: PSEUDOROOT: missing argument", x)
			}
			spell.pseudoRoot = tokens[1]

		case optSubstandard:
			if len(tokens) != 2 {
				return fmt.Errorf("line %d: SUBSTANDARD: missing argument", x)
			}
			spell.substandard = tokens[1]

		case optWordChars:
			if len(tokens) != 2 {
				return fmt.Errorf("line %d: WORDCHARS: missing argument", x)
			}
			spell.wordchars = tokens[1]

		case optCheckSharps:
			spell.isCheckSharps = true

		default:
			log.Printf("line %d: unknown option %q\n", x, tokens[0])
		}
	}

	return nil
}

//
// loadDictionary from string.
//
func (spell *Spell) loadDictionary(content string) (err error) {
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

		_, ok := spell.stems[s.value]
		if ok {
			log.Printf("duplicate stem %q", s.value)
		}

		derivatives, err := s.unpack(spell)
		if err != nil {
			return fmt.Errorf("line %d: %s", x, err.Error())
		}

		spell.stems[s.value] = s

		for _, w := range derivatives {
			spell.derivatives[w] = s
		}
	}

	return nil
}

//
// openAffix open and parse the affix from file.
// This function will cause all of the Spell options will be reset back to
// default values.
//
func (spell *Spell) openAffix(path string) (err error) {
	affcontent, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("openAffix: %s", err.Error())
	}

	err = spell.loadAffix(string(affcontent))
	if err != nil {
		return fmt.Errorf("openAffix: %s", err.Error())
	}

	return nil
}

//
// parseSet option from affix file.
//
func (spell *Spell) parseSet(args []string) {
	if len(args) == 0 {
		return
	}

	encoding := strings.ToUpper(args[0])

	switch encoding {
	case EncodingUTF8,
		EncodingISO8859_1, EncodingISO8859_2, EncodingISO8859_3,
		EncodingISO8859_4, EncodingISO8859_5, EncodingISO8859_6,
		EncodingISO8859_7, EncodingISO8859_8, EncodingISO8859_9,
		EncodingISO8859_10,
		EncodingISO8859_13, EncodingISO8859_14, EncodingISO8859_15,
		EncodingKOI8R, EncodingKOI8U, EncodingCP1251,
		EncodingISCIIDevanagri:
		spell.encoding = encoding
	default:
		log.Printf("hunspell: invalid SET value %q\n", encoding)
	}
}

//
// parseFlag parse the FLAG option from affix .
//
func (spell *Spell) parseFlag(flag string) {
	flag = strings.ToLower(flag)

	switch flag {
	case FlagASCII, FlagUTF8, FlagLong, FlagNum:
		spell.flag = flag
	default:
		log.Printf("hunspell: invalid FLAG value %q\n", flag)
	}
}

func (spell *Spell) parseAF(arg string) (err error) {
	if cap(spell.afAliases) == 0 {
		n, err := strconv.Atoi(arg)
		if err != nil {
			return err
		}

		spell.afAliases = make([]string, 0, n)
		spell.afAliases = append(spell.afAliases, "")
	} else {
		spell.afAliases = append(spell.afAliases, arg)
	}
	return nil
}

func (spell *Spell) parseAM(arg string) (err error) {
	if cap(spell.amAliases) == 0 {
		n, err := strconv.Atoi(arg)
		if err != nil {
			return err
		}

		spell.amAliases = make([]string, 0, n)
	} else {
		spell.amAliases = append(spell.amAliases, arg)
	}
	return nil
}

func (spell *Spell) parseRep(args []string) (err error) {
	if cap(spell.reps) == 0 {
		if len(args) != 1 {
			return fmt.Errorf("REP: missing number of replacement")
		}

		n, err := strconv.Atoi(args[0])
		if err != nil {
			return err
		}

		spell.reps = make([]replacement, 0, n)
	} else {
		if len(args) != 2 {
			return fmt.Errorf("REP: invalid arguments")
		}

		rep, err := newReplacement(args[0], args[1])
		if err != nil {
			return fmt.Errorf("REP: invalid argument %q", args[0])
		}

		spell.reps = append(spell.reps, rep)
	}
	return nil
}

func (spell *Spell) parseMap(arg string) (err error) {
	if cap(spell.charsMaps) == 0 {
		n, err := strconv.Atoi(arg)
		if err != nil {
			return fmt.Errorf("MAP: invalid argument %q: %s", arg, err.Error())
		}
		spell.charsMaps = make([]charsmap, 0, n)
		return nil
	}

	var (
		isGroup bool
		s       []rune
		cmap    charsmap
	)

	for _, r := range arg {
		if r == '(' {
			isGroup = true
			s = s[:0]
			continue
		}
		if isGroup {
			if r == ')' {
				isGroup = false
				cmap = append(cmap, string(s))
				continue
			}
			s = append(s, r)
			continue
		}
		cmap = append(cmap, string(r))
	}

	spell.charsMaps = append(spell.charsMaps, cmap)

	return nil
}

func (spell *Spell) parseBreak(arg string) (err error) {
	if cap(spell.breakOpts) == 0 {
		n, err := strconv.Atoi(arg)
		if err != nil {
			return fmt.Errorf("BREAK: invalid argument %q: %s", arg, err.Error())
		}

		spell.breakOpts = make([]breakopt, 0, n)

		return nil
	}

	breakrole := breakopt{}

	if arg[0] == '^' {
		breakrole.delEnd = true
		arg = arg[1:]
	}
	if len(arg) > 0 && arg[len(arg)-1] == '$' {
		breakrole.delStart = true
		arg = arg[:len(arg)-1]
	}
	if len(arg) == 0 {
		return fmt.Errorf("BREAK: empty character sequences")
	}

	breakrole.token = arg

	spell.breakOpts = append(spell.breakOpts, breakrole)

	return nil
}

func (spell *Spell) parseCompoundRule(arg string) (err error) {
	if cap(spell.compoundRules) == 0 {
		n, err := strconv.Atoi(arg)
		if err != nil {
			return fmt.Errorf("COMPOUNDRULE: invalid argument %q: %s", arg, err.Error())
		}

		spell.compoundRules = make([]compoundRule, 0, n)

		return nil
	}

	cr := compoundRule{}

	cr.pattern, err = regexp.Compile(arg)
	if err != nil {
		return fmt.Errorf("COMPOUNDRULE: invalid argument %q", arg)
	}

	spell.compoundRules = append(spell.compoundRules, cr)

	return nil
}

func (spell *Spell) parseCheckCompoundPattern(args []string) (err error) {
	if cap(spell.compoundPatterns) == 0 {
		n, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("CHECKCOMPOUNDPATTERN: invalid argument %q: %s", args[0], err.Error())
		}

		spell.compoundPatterns = make([]compoundPattern, 0, n)

		return nil
	}

	if len(args) < 2 {
		return fmt.Errorf("CHECKCOMPOUNDPATTERN: invalid argument %q", args)
	}

	cp := compoundPattern{}

	ss := strings.Split(args[0], "/")
	if len(ss) >= 1 {
		cp.end = ss[0]
	}
	if len(ss) >= 2 {
		cp.endFlag = ss[1]
	}

	ss = strings.Split(args[1], "/")
	if len(ss) >= 1 {
		cp.begin = ss[0]
	}
	if len(ss) == 2 {
		cp.beginFlag = ss[1]
	}

	if len(args) == 3 {
		cp.rep = args[2]
	}

	return nil
}

//
// parsePfx parse the prefix header and rules.
//
// The first line is the prefix header,
//
//	PFX flag cross_product number
//
// The flag option define the name of the affix class.
// The cross_product option define whether to allow to combine prefixes and
// suffixes.
// Possible values are: Y (yes) or N (no).
// The number option define the number of rules.
//
// The prefix rule,
//
//	PFX flag stripping prefix [condition [morphological_fields ...]]
//
// The stripping option contains list of characters to be stripped from
// beginning.  Empty stripping are indicated by zero character.
//
// The prefix option define string to be pre-prended to root word.
// Empty prefix are indicated by zero character.
//
// The condition option is simplified, regular expression like pattern, which
// must be met before the prefix can be applied.
// The dot ('.') condition means always true.
// Characters in braces, for example '[aiu]', sign that the stem must be
// prefixed by one of the characters.
// If circumflex ('^') is negation, which means the prefix will be applied if
// one of the character in braces is not the first character of stem.
//
// The morphological_fields are separated by spaces or tab.
//
func (spell *Spell) parsePfx(args []string) (err error) {
	flag := args[0]

	err = spell.isValidFlag(flag)
	if err != nil {
		return err
	}

	pfx, ok := spell.prefixes[flag]
	if !ok {
		// Parse the first line of prefix.
		isCrossProduct := (strings.ToLower(args[1]) == "y")

		n, err := strconv.Atoi(args[2])
		if err != nil {
			return fmt.Errorf("PFX: invalid number %q", args[2])
		}

		spell.prefixes[flag] = newAffix(flag, true, isCrossProduct, n)
	} else {
		// Parse the prefix rule.
		var (
			condition string
			morphemes []string
		)
		stripping := args[1]
		prefix := args[2]
		if len(args) >= 4 {
			condition = args[3]
		}
		if len(args) >= 5 {
			morphemes = args[4:]
		}

		err = pfx.addRule(spell, stripping, prefix, condition, morphemes)
		if err != nil {
			return fmt.Errorf("PFX: %s", err.Error())
		}
	}

	return nil
}

func (spell *Spell) parseSfx(args []string) (err error) {
	flag := args[0]

	err = spell.isValidFlag(flag)
	if err != nil {
		return err
	}

	sfx, ok := spell.suffixes[flag]
	if !ok {
		isCrossProduct := (strings.ToLower(args[1]) == "y")

		n, err := strconv.Atoi(args[2])
		if err != nil {
			return fmt.Errorf("SFX: invalid number %q", args[2])
		}

		spell.suffixes[flag] = newAffix(flag, false, isCrossProduct, n)
	} else {
		// Parse the prefix rule.
		var (
			condition string
			morphemes []string
		)
		stripping := args[1]
		suffix := args[2]
		if len(args) >= 4 {
			condition = args[3]
		}
		if len(args) >= 5 {
			morphemes = args[4:]
		}

		err = sfx.addRule(spell, stripping, suffix, condition, morphemes)
		if err != nil {
			return fmt.Errorf("SFX: %s", err.Error())
		}
	}

	return nil
}

func (spell *Spell) parseIconv(args []string) (err error) {
	if cap(spell.iconv) == 0 {
		n, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("ICONV: invalid argument %q: %s",
				args[0], err)
		}

		spell.iconv = make([]convertion, 0, n)
		return nil
	}
	if len(args) != 2 {
		return fmt.Errorf("ICONV: invalid arguments %q", args)
	}

	c := convertion{
		pattern:  args[0],
		pattern2: args[1],
	}

	spell.iconv = append(spell.iconv, c)

	return nil
}

func (spell *Spell) parseOconv(args []string) (err error) {
	if cap(spell.oconv) == 0 {
		n, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("ICONV: invalid argument %q: %s",
				args[0], err)
		}

		spell.oconv = make([]convertion, 0, n)
		return nil
	}
	if len(args) != 2 {
		return fmt.Errorf("ICONV: invalid arguments %q", args)
	}

	c := convertion{
		pattern:  args[0],
		pattern2: args[1],
	}

	spell.oconv = append(spell.oconv, c)

	return nil
}
