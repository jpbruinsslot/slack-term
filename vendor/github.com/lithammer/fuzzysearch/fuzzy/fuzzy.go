// Fuzzy searching allows for flexibly matching a string with partial input,
// useful for filtering data very quickly based on lightweight user input.
package fuzzy

import (
	"bytes"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

var foldTransformer = unicodeFoldTransformer{}
var noopTransformer = transform.Nop
var normalizeTransformer = transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
var normalizeFoldTransformer = transform.Chain(normalizeTransformer, foldTransformer)

// Match returns true if source matches target using a fuzzy-searching
// algorithm. Note that it doesn't implement Levenshtein distance (see
// RankMatch instead), but rather a simplified version where there's no
// approximation. The method will return true only if each character in the
// source can be found in the target and occurs after the preceding matches.
func Match(source, target string) bool {
	return match(source, target, noopTransformer)
}

// MatchFold is a case-insensitive version of Match.
func MatchFold(source, target string) bool {
	return match(source, target, foldTransformer)
}

// MatchNormalized is a unicode-normalized version of Match.
func MatchNormalized(source, target string) bool {
	return match(source, target, normalizeTransformer)
}

// MatchNormalizedFold is a unicode-normalized and case-insensitive version of Match.
func MatchNormalizedFold(source, target string) bool {
	return match(source, target, normalizeFoldTransformer)
}

func match(source, target string, transformer transform.Transformer) bool {
	source = stringTransform(source, transformer)
	target = stringTransform(target, transformer)

	lenDiff := len(target) - len(source)

	if lenDiff < 0 {
		return false
	}

	if lenDiff == 0 && source == target {
		return true
	}

Outer:
	for _, r1 := range source {
		for i, r2 := range target {
			if r1 == r2 {
				target = target[i+utf8.RuneLen(r2):]
				continue Outer
			}
		}
		return false
	}

	return true
}

// Find will return a list of strings in targets that fuzzy matches source.
func Find(source string, targets []string) []string {
	return find(source, targets, noopTransformer)
}

// FindFold is a case-insensitive version of Find.
func FindFold(source string, targets []string) []string {
	return find(source, targets, foldTransformer)
}

// FindNormalized is a unicode-normalized version of Find.
func FindNormalized(source string, targets []string) []string {
	return find(source, targets, normalizeTransformer)
}

// FindNormalizedFold is a unicode-normalized and case-insensitive version of Find.
func FindNormalizedFold(source string, targets []string) []string {
	return find(source, targets, normalizeFoldTransformer)
}

func find(source string, targets []string, transformer transform.Transformer) []string {
	var matches []string

	for _, target := range targets {
		if match(source, target, transformer) {
			matches = append(matches, target)
		}
	}

	return matches
}

// RankMatch is similar to Match except it will measure the Levenshtein
// distance between the source and the target and return its result. If there
// was no match, it will return -1.
// Given the requirements of match, RankMatch only needs to perform a subset of
// the Levenshtein calculation, only deletions need be considered, required
// additions and substitutions would fail the match test.
func RankMatch(source, target string) int {
	return rank(source, target, noopTransformer)
}

// RankMatchFold is a case-insensitive version of RankMatch.
func RankMatchFold(source, target string) int {
	return rank(source, target, foldTransformer)
}

// RankMatchNormalized is a unicode-normalized version of RankMatch.
func RankMatchNormalized(source, target string) int {
	return rank(source, target, normalizeTransformer)
}

// RankMatchNormalizedFold is a unicode-normalized and case-insensitive version of RankMatch.
func RankMatchNormalizedFold(source, target string) int {
	return rank(source, target, normalizeFoldTransformer)
}

func rank(source, target string, transformer transform.Transformer) int {
	lenDiff := len(target) - len(source)

	if lenDiff < 0 {
		return -1
	}

	source = stringTransform(source, transformer)
	target = stringTransform(target, transformer)

	if lenDiff == 0 && source == target {
		return 0
	}

	runeDiff := 0

Outer:
	for _, r1 := range source {
		for i, r2 := range target {
			if r1 == r2 {
				target = target[i+utf8.RuneLen(r2):]
				continue Outer
			} else {
				runeDiff++
			}
		}
		return -1
	}

	// Count up remaining char
	runeDiff += utf8.RuneCountInString(target)

	return runeDiff
}

// RankFind is similar to Find, except it will also rank all matches using
// Levenshtein distance.
func RankFind(source string, targets []string) Ranks {
	return rankFind(source, targets, noopTransformer)
}

// RankFindFold is a case-insensitive version of RankFind.
func RankFindFold(source string, targets []string) Ranks {
	return rankFind(source, targets, foldTransformer)
}

// RankFindNormalized is a unicode-normalizedversion of RankFind.
func RankFindNormalized(source string, targets []string) Ranks {
	return rankFind(source, targets, normalizeTransformer)
}

// RankFindNormalizedFold is a unicode-normalized and case-insensitive version of RankFind.
func RankFindNormalizedFold(source string, targets []string) Ranks {
	return rankFind(source, targets, normalizeFoldTransformer)
}

func rankFind(source string, targets []string, transformer transform.Transformer) Ranks {
	var r Ranks

	for index, target := range targets {
		if match(source, target, transformer) {
			distance := LevenshteinDistance(source, target)
			r = append(r, Rank{source, target, distance, index})
		}
	}
	return r
}

type Rank struct {
	// Source is used as the source for matching.
	Source string

	// Target is the word matched against.
	Target string

	// Distance is the Levenshtein distance between Source and Target.
	Distance int

	// Location of Target in original list
	OriginalIndex int
}

type Ranks []Rank

func (r Ranks) Len() int {
	return len(r)
}

func (r Ranks) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

func (r Ranks) Less(i, j int) bool {
	return r[i].Distance < r[j].Distance
}

func stringTransform(s string, t transform.Transformer) (transformed string) {
	var err error
	transformed, _, err = transform.String(t, s)
	if err != nil {
		transformed = s
	}

	return
}

type unicodeFoldTransformer struct{}

func (unicodeFoldTransformer) Transform(dst, src []byte, atEOF bool) (nDst, nSrc int, err error) {
	runes := bytes.Runes(src)
	var lowerRunes []rune
	for _, r := range runes {
		lowerRunes = append(lowerRunes, unicode.ToLower(r))
	}

	srcBytes := []byte(string(lowerRunes))
	n := copy(dst, srcBytes)
	if n < len(srcBytes) {
		err = transform.ErrShortDst
	}

	return n, n, err
}

func (unicodeFoldTransformer) Reset() {}
