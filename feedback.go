package memai

import "strings"

const (
	FeedbackBoostPositive = 0.05
	FeedbackBoostNegative = -0.05
)

var positivePatterns = []string{
	"ありがとう", "そうそう", "正解", "それそれ", "そうだよ",
	"そうだね", "合ってる", "当たり", "さすが", "よく覚えてる",
	"覚えてくれ",
}

var negativePatterns = []string{
	"違うよ", "違う!", "違う！", "それじゃない", "間違い",
	"間違えてる", "ハズレ", "覚えてない", "忘れてる", "そうじゃなく",
}

// DetectFeedback analyzes a message for positive/negative feedback signals
// about memory accuracy. Returns a boost delta:
//
//	positive → FeedbackBoostPositive (+0.05)
//	negative → FeedbackBoostNegative (-0.05)
//	neutral  → 0
func DetectFeedback(message string) float64 {
	for _, p := range positivePatterns {
		if strings.Contains(message, p) {
			return FeedbackBoostPositive
		}
	}
	for _, p := range negativePatterns {
		if strings.Contains(message, p) {
			return FeedbackBoostNegative
		}
	}
	return 0
}
