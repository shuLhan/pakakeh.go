// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package hunspell

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/shuLhan/share/lib/parser"
	libstrings "github.com/shuLhan/share/lib/strings"
)

type affixOptions struct {
	//
	// Affix file general options.
	//

	//
	// The "SET" option.
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
	// The "FLAG" option set
	// Default type is the extended ASCII (8-bit) character.
	//
	flag string

	//
	// The "COMPLEXPREFIXES" option set twofold prefix stripping (but
	// single suffix stripping), for example, for morphologically complex
	// languages with right-to-left writing system.
	//
	isComplexPrefixes bool

	//
	// The "LANG" option set language code for language-specific functions
	// of Hunspell.
	//
	// Use it to activate special casing of Azeri (LANG az), Turkish
	// (LANG tr) and Crimean Tatar (LANG crh), also not generalized
	// syllable-counting compounding rules of Hungarian (LANG hu).
	//
	lang string

	//
	// The "IGNORE" option sets characters to ignore dictionary words,
	// affixes, and input words.
	//
	ignore string

	//
	// The "AF" option.
	// afAliases contains list of affix that can be substituted with
	// number.  In this case, the number is index of slice.
	//
	afAliases []string

	//
	// The "AM" option.
	// amAliases contains list of affix rules that can be replaced with
	// ordinal number.
	//
	amAliases []string

	//
	// Affix file options for suggestion
	//

	// The "KEY" option.
	keys []string

	// The "TRY" option.
	// Hunspell can suggest right word forms, when they differ from the
	// bad input word by one TRY character.
	// The parameter of TRY is case sensitive.
	try string

	noSuggest          []string      // NOSUGGEST option.
	maxCompundSuggests int           // MAXCPDSUGS option.
	maxNGramSuggests   int           // MAXNGRAMSUGS option.
	maxDiff            int           // MAXDIFF option.
	isOnlyMaxDiff      bool          // ONLYMAXDIFF option.
	isNoSplitSugs      bool          // NOSPLITSUGS option.
	isSugsWithDots     bool          // SUGSWITHDOTS option.
	reps               []replacement // REP option.
	charsMaps          []charsmap    // MAP option.
	warn               string        // WARN option.
	isForbidWarn       bool          // FORBIDWARN option.

	//phone              map[string]string // PHONE option.

	//
	// Options for compounding.
	//

	breakOpts              []breakopt        // BREAK option.
	compoundRules          []compoundRule    // COMPOUNDRULE option.
	compoundMin            int               // COMPOUNDMIN option.
	compoundFlag           string            // COMPOUNDFLAG option.
	compoundBegin          string            // COMPOUNDBEGIN option.
	compoundLast           string            // COMPOUNDLAST option.
	compoundMiddle         string            // COMPOUNDMIDDLE option.
	onlyInCompound         string            // ONLYINCOMPOUND option.
	compoundPermitFlag     string            // COMPOUNDPERMITFLAG option.
	compoundForbidFlag     string            // COMPOUNDFORBIDFLAG option.
	isCompoundMoreSuffixes bool              // COMPOUNDMORESUFFIXES option.
	compoundRoot           string            // COMPOUNDROOT option.
	compoundWordMax        int               // COMPOUNDWORDMAX option.
	isCheckCompoundDup     bool              // CHECKCOMPOUNDDUP option.
	isCheckCompoundRep     bool              // CHECKCOMPOUNDREP option.
	isCheckCompoundCase    bool              // CHECKCOMPOUNDCASE option.
	isCheckCompoundTriple  bool              // CHECKCOMPOUNDTRIPLE option.
	isSimplifiedTriple     bool              // SIMPLIFIEDTRIPLE option.
	compoundPatterns       []compoundPattern // CHECKCOMPOUNDPATTERN option.
	forceUCase             string            // FORCEUCASE option.
	compoundSyllable       *compoundSyllable // COMPOUNDSYLLABLE option.
	syllableNum            string            // SYLLABLENUM option.

	//
	// Affix file options for affix creation.
	//

	prefixes map[string]*affix // PFX options.
	suffixes map[string]*affix // SFX options.

	//
	// Affix file other options.
	//

	circumfix     string       // CIRCUMFIX option.
	forbiddenWord string       // FORBIDDENWORD option.
	isFullStrip   bool         // FULLSTRIP option.
	keepCase      string       // KEEPCASE option.
	iconv         []conversion // ICONV option.
	oconv         []conversion // OCONV option.
	lemmaPresent  string       // LEMMA_PRESENT option.
	needAffix     string       // NEEDAFFIX option.
	pseudoRoot    string       // PSEUDOROOT option.
	substandard   string       // SUBSTANDARD option.
	wordchars     string       // WORDCHARS option.
	isCheckSharps bool         // CHECKSHARPS option.
}

//
// open open and parse the affix options from file.
// This function will cause all of the affix options will be reset back to
// its default values.
//
func (opts *affixOptions) open(file string) (err error) {
	affcontent, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("affixOptions.open: %w", err)
	}

	err = opts.load(string(affcontent))
	if err != nil {
		return err
	}

	return nil
}

//
// load affix options from string.
//
func (opts *affixOptions) load(content string) (err error) {
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
			opts.parseSet(tokens[1:])

		case optFlag:
			if len(tokens) != 2 {
				return fmt.Errorf("line %d: FLAG: missing argument", x)
			}

			opts.parseFlag(tokens[1])

		case optComplexPrefixes:
			opts.isComplexPrefixes = true

		case optLang:
			if len(tokens) == 1 {
				return fmt.Errorf("line %d: LANG: missing argument", x)
			}
			opts.lang = strings.ToLower(tokens[1])

		case optIgnore:
			if len(tokens) == 1 {
				return fmt.Errorf("line %d: IGNORE: missing argument", x)
			}
			opts.ignore = tokens[1]

		case optAF:
			if len(tokens) == 1 {
				return fmt.Errorf("line %d: AF: missing argument", x)
			}
			err = opts.parseAF(tokens[1])
			if err != nil {
				return fmt.Errorf("line %d: %s", x, err.Error())
			}

		case optAM:
			if len(tokens) == 1 {
				return fmt.Errorf("line %d: AM: missing argument", x)
			}
			err = opts.parseAM(tokens[1:])
			if err != nil {
				return fmt.Errorf("line %d: %s", x, err.Error())
			}

		case optKey:
			if len(tokens) == 1 {
				return fmt.Errorf("line %d: KEY: missing argument", x)
			}
			opts.keys = strings.Split(tokens[1], "|")

		case optTry:
			if len(tokens) == 1 {
				return fmt.Errorf("line %d: TRY: missing argument", x)
			}
			opts.try = tokens[1]

		case optNoSuggest:
			if len(tokens) == 1 {
				return fmt.Errorf("line %d: NOSUGGEST: missing argument", x)
			}
			opts.noSuggest = append(opts.noSuggest, tokens[1])

		case optMaxCPDSuggest:
			if len(tokens) == 1 {
				return fmt.Errorf("line %d: missing MAXCPDSUGS argument", x)
			}
			opts.maxCompundSuggests, err = strconv.Atoi(tokens[1])
			if err != nil {
				return fmt.Errorf("line %d: %s", x, err.Error())
			}

		case optMaxNGramSugs:
			if len(tokens) == 1 {
				return fmt.Errorf("line %d: MAXNGRAMSUGS: missing argument", x)
			}
			opts.maxNGramSuggests, err = strconv.Atoi(tokens[1])
			if err != nil {
				return fmt.Errorf("line %d: %s", x, err.Error())
			}

		case optMaxDiff:
			if len(tokens) == 1 {
				return fmt.Errorf("line %d: MAXNGRAMSUGS: missing argument", x)
			}
			opts.maxDiff, err = strconv.Atoi(tokens[1])
			if err != nil {
				return fmt.Errorf("line %d: %s", x, err.Error())
			}
			if opts.maxDiff < 1 || opts.maxDiff > 10 {
				opts.maxDiff = 5
			}

		case optOnlyMaxDiff:
			opts.isOnlyMaxDiff = true
		case optNoSplitSugs:
			opts.isNoSplitSugs = true

		case optSugsWithDots:
			opts.isSugsWithDots = true

		case optRep:
			err = opts.parseRep(tokens[1:])
			if err != nil {
				return fmt.Errorf("line %d: %s", x, err.Error())
			}

		case optMap:
			if len(tokens) != 2 {
				return fmt.Errorf("line %d: MAP: missing argument", x)
			}
			err = opts.parseMap(tokens[1])
			if err != nil {
				return fmt.Errorf("line %d: %s", x, err.Error())
			}

		case optPhone:
			// TODO

		case optWarn:
			if len(tokens) != 2 {
				return fmt.Errorf("line %d: WARN: missing argument", x)
			}
			opts.warn = tokens[1]

		case optForbidWarn:
			opts.isForbidWarn = true

		case optBreak:
			if len(tokens) != 2 {
				return fmt.Errorf("line %d: BREAK: missing argument", x)
			}
			err = opts.parseBreak(tokens[1])
			if err != nil {
				return fmt.Errorf("line %d: %s", x, err.Error())
			}

		case optCompoundRule:
			if len(tokens) != 2 {
				return fmt.Errorf("line %d:, COMPOUNDRULE: missing argument", x)
			}
			err = opts.parseCompoundRule(tokens[1])
			if err != nil {
				return fmt.Errorf("line %d: %s", x, err.Error())
			}

		case optCompoundMin:
			if len(tokens) != 2 {
				return fmt.Errorf("line %d: COMPOUNDMIN: missing argument", x)
			}
			opts.compoundMin, err = strconv.Atoi(tokens[1])
			if err != nil {
				return fmt.Errorf("line %d: COMPOUNDMIN: invalid argument %q", x, tokens[1])
			}

		case optCompoundFlag:
			if len(tokens) != 2 {
				return fmt.Errorf("line %d: COMPOUNDFLAG: missing argument", x)
			}
			opts.compoundFlag = tokens[1]

		case optCompoundBegin:
			if len(tokens) != 2 {
				return fmt.Errorf("line %d: COMPOUNDBEGIN: missing argument", x)
			}
			opts.compoundBegin = tokens[1]

		case optCompoundLast:
			if len(tokens) != 2 {
				return fmt.Errorf("line %d: COMPOUNDLAST: missing argument", x)
			}
			opts.compoundLast = tokens[1]

		case optCompoundMiddle:
			if len(tokens) != 2 {
				return fmt.Errorf("line %d: COMPOUNDMIDDLE: missing argument", x)
			}
			opts.compoundMiddle = tokens[1]

		case optOnlyInCompound:
			if len(tokens) != 2 {
				return fmt.Errorf("line %d: ONLYINCOMPOUND: missing argument", x)
			}
			opts.onlyInCompound = tokens[1]

		case optCompoundPermitFlag:
			if len(tokens) != 2 {
				return fmt.Errorf("line %d: COMPOUNDPERMITFLAG: missing argument", x)
			}
			opts.compoundPermitFlag = tokens[1]

		case optCompoundForbidFlags:
			if len(tokens) != 2 {
				return fmt.Errorf("line %d: COMPOUNDFORBIDFLAG: missing argument", x)
			}
			opts.compoundForbidFlag = tokens[1]

		case optCompoundMoreSuffixes:
			opts.isCompoundMoreSuffixes = true

		case optCompoundRoot:
			if len(tokens) != 2 {
				return fmt.Errorf("line %d: COMPOUNDROOT: missing argument", x)
			}
			opts.compoundRoot = tokens[1]

		case optCompoundWordMax:
			if len(tokens) != 2 {
				return fmt.Errorf("line %d: COMPOUNDWORDMAX: missing argument", x)
			}
			opts.compoundWordMax, err = strconv.Atoi(tokens[1])
			if err != nil {
				return fmt.Errorf("line %d: COMPOUNDWORDMAX: invalid argument %q", x, tokens[1])
			}

		case optCheckCompoundDup:
			opts.isCheckCompoundDup = true
		case optCheckCompoundRep:
			opts.isCheckCompoundRep = true
		case optCheckCompoundCase:
			opts.isCheckCompoundCase = true
		case optCheckCompoundTriple:
			opts.isCheckCompoundTriple = true
		case optSimplifiedTriple:
			opts.isSimplifiedTriple = true

		case optCheckCompoundPattern:
			if len(tokens) == 1 {
				return fmt.Errorf("line %d: CHECKCOMPOUNDPATTERN: missing argument", x)
			}
			err = opts.parseCheckCompoundPattern(tokens)
			if err != nil {
				return fmt.Errorf("line %d: %s", x, err.Error())
			}

		case optForceUcase:
			if len(tokens) != 2 {
				return fmt.Errorf("line %d: FORCEUCASE: missing argument", x)
			}
			opts.forceUCase = tokens[1]

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
			opts.compoundSyllable = cs

		case optSyllableNum:
			if len(tokens) != 2 {
				return fmt.Errorf("line %d: SYLLABLENUM: missing argument", x)
			}
			opts.syllableNum = tokens[1]

		case optPFX:
			if len(tokens) < 3 {
				return fmt.Errorf("line %d: PFX: missing arguments", x)
			}
			err = opts.parsePfx(tokens[1:])
			if err != nil {
				return fmt.Errorf("line %d: %s", x, err.Error())
			}

		case optSFX:
			if len(tokens) < 3 {
				return fmt.Errorf("line %d: SFX: missing arguments", x)
			}
			err = opts.parseSfx(tokens[1:])
			if err != nil {
				return fmt.Errorf("line %d: %s", x, err.Error())
			}

		case optCircumfix:
			if len(tokens) != 2 {
				return fmt.Errorf("line %d: CIRCUMFIX: missing argument", x)
			}
			opts.circumfix = tokens[1]

		case optForbiddenWord:
			if len(tokens) != 2 {
				return fmt.Errorf("line %d: FORBIDDENWORD: missing argument", x)
			}
			opts.forbiddenWord = tokens[1]

		case optFullStrip:
			opts.isFullStrip = true

		case optKeepCase:
			if len(tokens) != 2 {
				return fmt.Errorf("line %d: KEEPCASE: missing argument", x)
			}
			opts.keepCase = tokens[1]

		case optIconv:
			if len(tokens) == 1 {
				return fmt.Errorf("line %d: ICONV: missing argument", x)
			}
			err = opts.parseIconv(tokens[1:])
			if err != nil {
				return fmt.Errorf("line %d: %s", x, err.Error())
			}

		case optOconv:
			if len(tokens) != 3 {
				return fmt.Errorf("line %d: OCONV: missing argument", x)
			}
			err = opts.parseOconv(tokens[1:])
			if err != nil {
				return fmt.Errorf("line %d: %s", x, err.Error())
			}

		case optLemmaPresent:
			if len(tokens) != 2 {
				return fmt.Errorf("line %d: LEMMA_PRESENT: missing argument", x)
			}
			opts.lemmaPresent = tokens[1]

		case optNeedAffix:
			if len(tokens) != 2 {
				return fmt.Errorf("line %d: NEEDAFFIX: missing argument", x)
			}
			opts.needAffix = tokens[1]

		case optPseudoRoot:
			if len(tokens) != 2 {
				return fmt.Errorf("line %d: PSEUDOROOT: missing argument", x)
			}
			opts.pseudoRoot = tokens[1]

		case optSubstandard:
			if len(tokens) != 2 {
				return fmt.Errorf("line %d: SUBSTANDARD: missing argument", x)
			}
			opts.substandard = tokens[1]

		case optWordChars:
			if len(tokens) != 2 {
				return fmt.Errorf("line %d: WORDCHARS: missing argument", x)
			}
			opts.wordchars = tokens[1]

		case optCheckSharps:
			opts.isCheckSharps = true

		default:
			log.Printf("line %d: unknown option %q\n", x, tokens[0])
		}
	}

	return nil
}

//
// parseSet option from affix file.
//
func (opts *affixOptions) parseSet(args []string) {
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
		opts.encoding = encoding
	default:
		log.Printf("hunspell: invalid SET value %q\n", encoding)
	}
}

//
// parseFlag parse the FLAG option from affix .
//
func (opts *affixOptions) parseFlag(flag string) {
	flag = strings.ToLower(flag)

	switch flag {
	case FlagASCII, FlagUTF8, FlagLong, FlagNum:
		opts.flag = flag
	default:
		log.Printf("hunspell: invalid FLAG value %q\n", flag)
	}
}

func (opts *affixOptions) parseAF(arg string) (err error) {
	if cap(opts.afAliases) == 0 {
		n, err := strconv.Atoi(arg)
		if err != nil {
			return err
		}

		opts.afAliases = make([]string, 0, n)
		opts.afAliases = append(opts.afAliases, "")
	} else {
		opts.afAliases = append(opts.afAliases, arg)
	}
	return nil
}

func (opts *affixOptions) parseAM(args []string) (err error) {
	if cap(opts.amAliases) == 0 {
		n, err := strconv.Atoi(args[0])
		if err != nil {
			return err
		}

		opts.amAliases = make([]string, 0, n+1)
		opts.amAliases = append(opts.amAliases, "")
	} else {
		opts.amAliases = append(opts.amAliases, strings.Join(args, " "))
	}
	return nil
}

func (opts *affixOptions) parseRep(args []string) (err error) {
	if cap(opts.reps) == 0 {
		if len(args) != 1 {
			return fmt.Errorf("REP: missing number of replacement")
		}

		n, err := strconv.Atoi(args[0])
		if err != nil {
			return err
		}

		opts.reps = make([]replacement, 0, n)
	} else {
		if len(args) != 2 {
			return fmt.Errorf("REP: invalid arguments")
		}

		rep, err := newReplacement(args[0], args[1])
		if err != nil {
			return fmt.Errorf("REP: invalid argument %q", args[0])
		}

		opts.reps = append(opts.reps, rep)
	}
	return nil
}

func (opts *affixOptions) parseMap(arg string) (err error) {
	if cap(opts.charsMaps) == 0 {
		n, err := strconv.Atoi(arg)
		if err != nil {
			return fmt.Errorf("MAP: invalid argument %q: %s", arg, err.Error())
		}
		opts.charsMaps = make([]charsmap, 0, n)
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

	opts.charsMaps = append(opts.charsMaps, cmap)

	return nil
}

func (opts *affixOptions) parseBreak(arg string) (err error) {
	if cap(opts.breakOpts) == 0 {
		n, err := strconv.Atoi(arg)
		if err != nil {
			return fmt.Errorf("BREAK: invalid argument %q: %s", arg, err.Error())
		}

		opts.breakOpts = make([]breakopt, 0, n)

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

	opts.breakOpts = append(opts.breakOpts, breakrole)

	return nil
}

func (opts *affixOptions) parseCompoundRule(arg string) (err error) {
	if cap(opts.compoundRules) == 0 {
		n, err := strconv.Atoi(arg)
		if err != nil {
			return fmt.Errorf("COMPOUNDRULE: invalid argument %q: %s", arg, err.Error())
		}

		opts.compoundRules = make([]compoundRule, 0, n)

		return nil
	}

	cr := compoundRule{}

	cr.pattern, err = regexp.Compile(arg)
	if err != nil {
		return fmt.Errorf("COMPOUNDRULE: invalid argument %q", arg)
	}

	opts.compoundRules = append(opts.compoundRules, cr)

	return nil
}

func (opts *affixOptions) parseCheckCompoundPattern(args []string) (err error) {
	if cap(opts.compoundPatterns) == 0 {
		n, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("CHECKCOMPOUNDPATTERN: invalid argument %q: %s", args[0], err.Error())
		}

		opts.compoundPatterns = make([]compoundPattern, 0, n)

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
func (opts *affixOptions) parsePfx(args []string) (err error) {
	flag := args[0]

	err = opts.isValidFlag(flag)
	if err != nil {
		return err
	}

	pfx, ok := opts.prefixes[flag]
	if !ok {
		// Parse the first line of prefix.
		isCrossProduct := (strings.ToLower(args[1]) == "y")

		n, err := strconv.Atoi(args[2])
		if err != nil {
			return fmt.Errorf("PFX: invalid number %q", args[2])
		}

		opts.prefixes[flag] = newAffix(flag, true, isCrossProduct, n)
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

		err = pfx.addRule(opts, stripping, prefix, condition, morphemes)
		if err != nil {
			return fmt.Errorf("PFX: %s", err.Error())
		}
	}

	return nil
}

func (opts *affixOptions) parseSfx(args []string) (err error) {
	flag := args[0]

	err = opts.isValidFlag(flag)
	if err != nil {
		return err
	}

	sfx, ok := opts.suffixes[flag]
	if !ok {
		isCrossProduct := (strings.ToLower(args[1]) == "y")

		n, err := strconv.Atoi(args[2])
		if err != nil {
			return fmt.Errorf("SFX: invalid number %q", args[2])
		}

		opts.suffixes[flag] = newAffix(flag, false, isCrossProduct, n)
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

		err = sfx.addRule(opts, stripping, suffix, condition, morphemes)
		if err != nil {
			return fmt.Errorf("SFX: %s", err.Error())
		}
	}

	return nil
}

func (opts *affixOptions) parseIconv(args []string) (err error) {
	if cap(opts.iconv) == 0 {
		n, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("ICONV: invalid argument %q: %s",
				args[0], err)
		}

		opts.iconv = make([]conversion, 0, n)
		return nil
	}
	if len(args) != 2 {
		return fmt.Errorf("ICONV: invalid arguments %q", args)
	}

	c := conversion{
		pattern:  args[0],
		pattern2: args[1],
	}

	opts.iconv = append(opts.iconv, c)

	return nil
}

func (opts *affixOptions) parseOconv(args []string) (err error) {
	if cap(opts.oconv) == 0 {
		n, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("ICONV: invalid argument %q: %s",
				args[0], err)
		}

		opts.oconv = make([]conversion, 0, n)
		return nil
	}
	if len(args) != 2 {
		return fmt.Errorf("ICONV: invalid arguments %q", args)
	}

	c := conversion{
		pattern:  args[0],
		pattern2: args[1],
	}

	opts.oconv = append(opts.oconv, c)

	return nil
}

//
// isValidFlag check whether the flag value conform the FLAG type.
//
func (opts *affixOptions) isValidFlag(flag string) error {
	switch opts.flag {
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
