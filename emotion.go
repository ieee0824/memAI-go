package memai

import "strings"

// emotionKeywords maps emotion types to their Japanese keyword triggers.
var emotionKeywords = map[EmotionType][]string{
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

// emotionValence maps each emotion to its base valence value.
var emotionValence = map[EmotionType]float64{
	EmotionJoy:      0.8,
	EmotionSadness:  -0.6,
	EmotionAnger:    -0.8,
	EmotionFear:     -0.5,
	EmotionSurprise: 0.3,
	EmotionNeutral:  0.0,
}

// AnalyzeEmotion detects emotion from a Japanese text message using keyword matching.
// Returns EmotionNeutral with zero intensity if no emotional keywords are found.
func AnalyzeEmotion(msg string) *EmotionalState {
	bestEmotion := EmotionNeutral
	bestCount := 0

	for emotion, keywords := range emotionKeywords {
		count := 0
		for _, kw := range keywords {
			if strings.Contains(msg, kw) {
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
