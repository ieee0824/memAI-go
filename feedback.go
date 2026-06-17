package memai

import "strings"

const (
	FeedbackBoostPositive = 0.05
	FeedbackBoostNegative = -0.05
)

var positivePatternsJA = []string{
	"ありがとう", "そうそう", "正解", "それそれ", "そうだよ",
	"そうだね", "合ってる", "当たり", "さすが", "よく覚えてる",
	"覚えてくれ",
}

var negativePatternsJA = []string{
	"違うよ", "違う!", "違う！", "それじゃない", "間違い",
	"間違えてる", "ハズレ", "覚えてない", "忘れてる", "そうじゃなく",
}

// negationSuffixesJA are clause-level negators. When one of these immediately
// follows a positive pattern, the positive token is actually part of a
// correction (e.g. "正解じゃない", "覚えてくれてない") and must be treated as
// negative rather than positive feedback.
var negationSuffixesJA = []string{
	"じゃない", "ではない", "じゃなく", "ではなく",
	"でない", "てない", "ていない", "ない",
}

// English patterns are matched on a lowercased copy of the message.
var positivePatternsEN = []string{
	"thank", "exactly", "correct", "that's right", "thats right",
	"you remembered", "you got it", "spot on", "good memory", "well remembered",
}

var negativePatternsEN = []string{
	"wrong", "not right", "incorrect", "that's not it", "thats not it",
	"you forgot", "don't remember", "dont remember", "didn't remember",
	"didnt remember", "misremember",
}

// DetectFeedback analyzes a message for positive/negative feedback signals
// about memory accuracy. lang selects the keyword set (LangJapanese or
// LangEnglish; English matching is case-insensitive). Returns a boost delta:
//
//	positive → FeedbackBoostPositive (+0.05)
//	negative → FeedbackBoostNegative (-0.05)
//	neutral  → 0
//
// Negative signals take precedence over positive ones, and (for Japanese) a
// positive pattern immediately followed by a negation suffix is treated as
// negative, so corrections like "正解じゃない" or "覚えてくれてない" are not
// misclassified as praise.
func DetectFeedback(message string, lang Language) float64 {
	pos, neg := positivePatternsJA, negativePatternsJA
	target := message
	negationAware := true

	if lang == LangEnglish {
		pos, neg = positivePatternsEN, negativePatternsEN
		target = strings.ToLower(message)
		negationAware = false
	}

	// Negation wins: an explicit negative correction takes precedence over any
	// positive token in the same message (e.g. "ありがとう、でもそれじゃない").
	if containsAny(target, neg) {
		return FeedbackBoostNegative
	}
	// A positive pattern immediately negated is a correction, not praise.
	if negationAware && negatedPositive(target, pos) {
		return FeedbackBoostNegative
	}
	if containsAny(target, pos) {
		return FeedbackBoostPositive
	}
	return 0
}

// containsAny reports whether s contains any of the given patterns.
func containsAny(s string, patterns []string) bool {
	for _, p := range patterns {
		if strings.Contains(s, p) {
			return true
		}
	}
	return false
}

// negatedPositive reports whether any positive pattern in s is immediately
// followed by a Japanese negation suffix.
func negatedPositive(s string, positives []string) bool {
	for _, p := range positives {
		_, rest, found := strings.Cut(s, p)
		if !found {
			continue
		}
		for _, neg := range negationSuffixesJA {
			if strings.HasPrefix(rest, neg) {
				return true
			}
		}
	}
	return false
}
