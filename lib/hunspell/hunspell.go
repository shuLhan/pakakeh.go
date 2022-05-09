// Copyright 2019, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package hunspell is a library to parse the Hunspell file format.
package hunspell

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/shuLhan/share/lib/ascii"
	libio "github.com/shuLhan/share/lib/io"
	"github.com/shuLhan/share/lib/parser"
)

// List of affix file general options.
const (
	optSet             = "SET"
	optFlag            = "FLAG"
	optComplexPrefixes = "COMPLEXPREFIXES"
	optLang            = "LANG"
	optIgnore          = "IGNORE"
	optAF              = "AF"
	optAM              = "AM"
)

// List of affix file options for suggestion.
const (
	optKey           = "KEY"
	optTry           = "TRY"
	optNoSuggest     = "NOSUGGEST"
	optMaxCPDSuggest = "MAXCPDSUGS"
	optMaxNGramSugs  = "MAXNGRAMSUGS"
	optMaxDiff       = "MAXDIFF"
	optOnlyMaxDiff   = "ONLYMAXDIFF"
	optNoSplitSugs   = "NOSPLITSUGS"
	optSugsWithDots  = "SUGSWITHDOTS"
	optRep           = "REP"
	optMap           = "MAP"
	optPhone         = "PHONE"
	optWarn          = "WARN"
	optForbidWarn    = "FORBIDWARN"
)

// List of affix file options for compounding.
const (
	optBreak                = "BREAK"
	optCompoundRule         = "COMPOUNDRULE"
	optCompoundMin          = "COMPOUNDMIN"
	optCompoundFlag         = "COMPOUNDFLAG"
	optCompoundBegin        = "COMPOUNDBEGIN"
	optCompoundLast         = "COMPOUNDLAST"
	optCompoundMiddle       = "COMPOUNDMIDDLE"
	optOnlyInCompound       = "ONLYINCOMPOUND"
	optCompoundPermitFlag   = "COMPOUNDPERMITFLAG"
	optCompoundForbidFlags  = "COMPOUNDFORBIDFLAG"
	optCompoundMoreSuffixes = "COMPOUNDMORESUFFIXES"
	optCompoundRoot         = "COMPOUNDROOT"
	optCompoundWordMax      = "COMPOUNDWORDMAX"
	optCheckCompoundDup     = "CHECKCOMPOUNDDUP"
	optCheckCompoundRep     = "CHECKCOMPOUNDREP"
	optCheckCompoundCase    = "CHECKCOMPOUNDCASE"
	optCheckCompoundTriple  = "CHECKCOMPOUNDTRIPLE"
	optSimplifiedTriple     = "SIMPLIFIEDTRIPLE"
	optCheckCompoundPattern = "CHECKCOMPOUNDPATTERN"
	optForceUcase           = "FORCEUCASE"
	optCompoundSyllable     = "COMPOUNDSYLLABLE"
	optSyllableNum          = "SYLLABLENUM"
)

// List of affix file options for affix creation.
const (
	optPFX = "PFX"
	optSFX = "SFX"
)

// List of affix file other options.
const (
	optCircumfix     = "CIRCUMFIX"
	optForbiddenWord = "FORBIDDENWORD"
	optFullStrip     = "FULLSTRIP"
	optKeepCase      = "KEEPCASE"
	optIconv         = "ICONV"
	optOconv         = "OCONV"
	optLemmaPresent  = "LEMMA_PRESENT"
	optNeedAffix     = "NEEDAFFIX"
	optPseudoRoot    = "PSEUDOROOT"
	optSubstandard   = "SUBSTANDARD"
	optWordChars     = "WORDCHARS"
	optCheckSharps   = "CHECKSHARPS"
)

// List of morphological and other IDs.
//
// nolint: deadcode,varcheck
const (
	morphKeyPH = "ph"
	morphKeyST = "st"
	morphKeyAL = "al"
	morphKeyPO = "po"
	morphKeyDS = "ds"
	morphKeyIS = "is"
	morphKeyTS = "ts"
	morphKeySP = "sp"
	morphKeyPA = "pa"
	morphKeyDP = "dp"
	morphKeyIP = "ip"
	morphKeyTP = "tp"
)

// List of valid SET values.
const (
	EncodingUTF8           = "UTF-8" // Default
	EncodingISO8859_1      = "ISO8859-1"
	EncodingISO8859_2      = "ISO8859-2"
	EncodingISO8859_3      = "ISO8859-3"
	EncodingISO8859_4      = "ISO8859-4"
	EncodingISO8859_5      = "ISO8859-5"
	EncodingISO8859_6      = "ISO8859-6"
	EncodingISO8859_7      = "ISO8859-7"
	EncodingISO8859_8      = "ISO8859-8"
	EncodingISO8859_9      = "ISO8859-9"
	EncodingISO8859_10     = "ISO8859-10"
	EncodingISO8859_13     = "ISO8859-13"
	EncodingISO8859_14     = "ISO8859-14"
	EncodingISO8859_15     = "ISO8859-15"
	EncodingKOI8R          = "KOI8-R"
	EncodingKOI8U          = "KOI8-U"
	EncodingCP1251         = "CP1251"
	EncodingISCIIDevanagri = "ISCII-DEVANAGRI"
)

// List of valid flag values.
const (
	// Default flag with single character.
	FlagASCII = "ascii"

	// `UTF-8' parameter  sets  UTF-8  encoded Unicode character flags,
	// single character.
	FlagUTF8 = "utf-8"

	//  The `long' value sets the double extended ASCII character flag
	//  type, double ASCII characters.
	FlagLong = "long"

	// Decimal flags numbered from 1 to 65000, and in flag fields are
	// separated by comma.
	FlagNum = "num"
)

// List of default values.
const (
	DefaultEncoding        = EncodingUTF8
	DefaultFlag            = FlagASCII
	defaultMinimumCompound = 3
)

func errInvalidMorpheme(in string) error {
	return fmt.Errorf("invalid morpheme %q", in)
}

// MergeDictionaries merge two or more dictionaries into single file.
// The outFile define the output of merged dictionaries.
// If the outFile already exist it will be truncated, otherwise it will be
// created.
// The inFiles contains list of input dictionary files.
//
// On success it will return number of words merged into output file.
func MergeDictionaries(outFile string, inFiles ...string) (n int, err error) {
	if len(inFiles) == 0 {
		return 0, nil
	}

	if len(inFiles) == 1 {
		err = libio.Copy(outFile, inFiles[0])
		return 0, err
	}

	dict := make(map[string]string, 1024)

	for x := 0; x < len(inFiles); x++ {
		lines, err := parser.Lines(inFiles[x])
		if err != nil {
			return 0, err
		}

		// Skip the first line that may contains number of words.
		y := 0
		if ascii.IsDigit(lines[y][0]) {
			y = 1
		}

		for ; y < len(lines); y++ {
			ss := strings.Split(lines[y], "/")
			key := ss[0]
			attr := dict[key] + strings.Join(ss[1:], "")
			dict[key] = attr
		}
	}

	words := make([]string, 0, len(dict))

	for word, attr := range dict {
		if len(attr) == 0 {
			words = append(words, word)
		} else {
			words = append(words, word+"/"+attr)
		}
	}

	sort.Strings(words)

	fout, err := os.OpenFile(outFile, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return 0, err
	}

	fmt.Fprintf(fout, "%d\n", len(words))
	for x := 0; x < len(words); x++ {
		fmt.Fprintf(fout, "%s\n", words[x])
	}

	err = fout.Close()
	if err != nil {
		return 0, err
	}

	return len(words), nil
}
