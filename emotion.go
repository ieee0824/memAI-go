package memai

import (
	"context"
	"strings"
	"unicode"
)

// emotionKeywordsJA maps emotion types to Japanese keyword triggers.
var emotionKeywordsJA = map[EmotionType][]string{
	EmotionJoy: {
		"嬉しい", "うれしい", "楽しい", "たのしい", "ありがとう",
		"最高", "やった", "良かった", "よかった", "素晴らしい",
		"素敵", "すてき", "幸せ", "しあわせ", "わくわく",
		"好き", "大好き", "笑",
	},
	EmotionSadness: {
		"悲しい", "かなしい", "辛い", "つらい", "寂しい",
		"さみしい", "残念", "泣", "落ち込", "しんどい",
		"切ない", "せつない", "がっかり", "凹", "へこ",
	},
	EmotionAnger: {
		"むかつく", "ふざけ", "ありえない", "最悪", "怒",
		"イライラ", "いらいら", "うざ", "腹立", "ムカ",
		"ひどい", "酷い", "許せない", "ゆるせない",
	},
	EmotionFear: {
		"心配", "不安", "怖い", "こわい",
		"やばい", "ヤバい", "焦", "あせ", "緊張",
		"ドキドキ", "どきどき", "恐",
	},
	EmotionSurprise: {
		"まじ", "マジ", "え？", "えっ", "信じられない",
		"びっくり", "驚", "すごい", "スゴい", "すげ",
		"意外", "まさか", "うそ", "ウソ",
	},
}

// emotionKeywordsEN maps emotion types to English keyword triggers (all lowercase).
// Matching is performed on a lowercased copy of the input.
var emotionKeywordsEN = map[EmotionType][]string{
	EmotionJoy: {
		"happy", "glad", "excited", "grateful", "thankful",
		"wonderful", "great", "love", "awesome", "fantastic",
		"joy", "joyful", "delighted", "pleased", "cheerful", "lol",
	},
	EmotionSadness: {
		"sad", "unhappy", "depressed", "lonely", "miss",
		"cry", "crying", "tears", "heartbroken", "miserable",
		"down", "blue", "grief", "sorrow",
	},
	EmotionAnger: {
		"angry", "mad", "furious", "annoyed", "frustrated",
		"hate", "disgusting", "terrible", "awful", "horrible",
		"outrageous", "unacceptable", "rage",
	},
	EmotionFear: {
		"scared", "afraid", "worried", "anxious", "nervous",
		"fear", "terrified", "panic", "uneasy", "stressed", "dread",
	},
	EmotionSurprise: {
		"wow", "omg", "unbelievable", "amazing", "incredible",
		"shocking", "no way", "unexpected",
	},
}

// emotionValence maps each emotion to its base valence value.
var emotionValence = map[EmotionType]float64{
	EmotionJoy:      0.8,
	EmotionSadness:  -0.6,
	EmotionAnger:    -0.8,
	EmotionFear:     -0.5,
	EmotionSurprise: 0.3,
	EmotionNeutral:  0.0,
}

// emotionOrder fixes the evaluation order of emotions. Ranging a map has a
// randomized iteration order in Go, so iterating this slice instead guarantees
// a deterministic winner when two emotions match the same number of keywords
// (the earlier entry wins).
var emotionOrder = []EmotionType{
	EmotionJoy,
	EmotionSadness,
	EmotionAnger,
	EmotionFear,
	EmotionSurprise,
}

// AnalyzeEmotion detects emotion from a text message using keyword matching.
// lang must be LangJapanese or LangEnglish; English matching is case-insensitive.
// Returns EmotionNeutral with zero intensity if no emotional keywords are found.
func AnalyzeEmotion(msg string, lang Language) *EmotionalState {
	var keywords map[EmotionType][]string
	var target string
	var tokens map[string]struct{}

	switch lang {
	case LangEnglish:
		keywords = emotionKeywordsEN
		target = strings.ToLower(msg)
		tokens = tokenizeEN(target)
	default: // LangJapanese
		keywords = emotionKeywordsJA
		target = msg
	}

	bestEmotion := EmotionNeutral
	bestCount := 0

	// Iterate in a fixed order (not over the map) so ties resolve deterministically.
	for _, emotion := range emotionOrder {
		count := 0
		for _, kw := range keywords[emotion] {
			if matchKeyword(target, tokens, kw, lang) {
				count++
			}
		}
		if count > bestCount {
			bestCount = count
			bestEmotion = emotion
		}
	}

	if bestCount == 0 {
		return &EmotionalState{
			Primary:   EmotionNeutral,
			Intensity: 0,
			Valence:   0,
		}
	}

	// Intensity: 1 match=0.4, 2=0.6, 3+=0.8
	intensity := 0.4
	if bestCount >= 3 {
		intensity = 0.8
	} else if bestCount >= 2 {
		intensity = 0.6
	}

	// Exclamation density boost
	exclCount := strings.Count(msg, "!") + strings.Count(msg, "！")
	if exclCount > 0 {
		intensity += 0.1
	}
	if intensity > 1.0 {
		intensity = 1.0
	}

	valence := emotionValence[bestEmotion] * intensity

	return &EmotionalState{
		Primary:   bestEmotion,
		Intensity: intensity,
		Valence:   valence,
	}
}

// tokenizeEN splits an already-lowercased English string into the set of its
// word tokens (runs of letters/digits), used for whole-word keyword matching.
func tokenizeEN(lower string) map[string]struct{} {
	set := make(map[string]struct{})
	for _, tok := range strings.FieldsFunc(lower, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	}) {
		set[tok] = struct{}{}
	}
	return set
}

// matchKeyword reports whether kw matches target. For English, single-word
// keywords are matched against whole tokens (avoiding substring false positives
// such as "miss" in "mission" or "fear" in "fearless"); multi-word keywords
// (e.g. "no way") fall back to substring matching. Japanese has no word
// separators, so substring matching is used as before.
func matchKeyword(target string, tokens map[string]struct{}, kw string, lang Language) bool {
	if lang == LangEnglish {
		if strings.ContainsRune(kw, ' ') {
			return strings.Contains(target, kw)
		}
		_, ok := tokens[kw]
		return ok
	}
	return strings.Contains(target, kw)
}

// KeywordEmotionAnalyzer implements EmotionAnalyzer using keyword matching.
// Use NewKeywordEmotionAnalyzer to create an instance.
type KeywordEmotionAnalyzer struct {
	lang Language
}

// NewKeywordEmotionAnalyzer returns a keyword-based EmotionAnalyzer for the given language.
func NewKeywordEmotionAnalyzer(lang Language) *KeywordEmotionAnalyzer {
	return &KeywordEmotionAnalyzer{lang: lang}
}

// Analyze implements EmotionAnalyzer using keyword matching.
// The ctx argument is unused but satisfies the interface for drop-in LLM replacement.
func (a *KeywordEmotionAnalyzer) Analyze(_ context.Context, msg string) (*EmotionalState, error) {
	return AnalyzeEmotion(msg, a.lang), nil
}
