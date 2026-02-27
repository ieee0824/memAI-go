package memai

import "strings"

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
		"心配", "不安", "怖い", "こわい", "大丈夫",
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
		"shocking", "really", "seriously", "no way", "unexpected",
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

// AnalyzeEmotion detects emotion from a text message using keyword matching.
// lang must be LangJapanese or LangEnglish; English matching is case-insensitive.
// Returns EmotionNeutral with zero intensity if no emotional keywords are found.
func AnalyzeEmotion(msg string, lang Language) *EmotionalState {
	var keywords map[EmotionType][]string
	var target string

	switch lang {
	case LangEnglish:
		keywords = emotionKeywordsEN
		target = strings.ToLower(msg)
	default: // LangJapanese
		keywords = emotionKeywordsJA
		target = msg
	}

	bestEmotion := EmotionNeutral
	bestCount := 0

	for emotion, kws := range keywords {
		count := 0
		for _, kw := range kws {
			if strings.Contains(target, kw) {
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
