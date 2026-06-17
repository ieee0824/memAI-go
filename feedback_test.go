package memai

import "testing"

func TestDetectFeedback_Positive(t *testing.T) {
	cases := []string{"ありがとう！", "そうそう、それ", "さすが"}
	for _, msg := range cases {
		if DetectFeedback(msg, LangJapanese) != FeedbackBoostPositive {
			t.Errorf("expected positive for %q", msg)
		}
	}
}

func TestDetectFeedback_Negative(t *testing.T) {
	cases := []string{"違うよ", "それじゃない", "忘れてるよ"}
	for _, msg := range cases {
		if DetectFeedback(msg, LangJapanese) != FeedbackBoostNegative {
			t.Errorf("expected negative for %q", msg)
		}
	}
}

func TestDetectFeedback_Neutral(t *testing.T) {
	if DetectFeedback("明日の天気は？", LangJapanese) != 0 {
		t.Error("expected neutral feedback")
	}
}

// Regression: a positive token embedded in a negation must not be scored
// positive (issue #1).
func TestDetectFeedback_NegatedPositive(t *testing.T) {
	cases := []string{
		"覚えてくれてないじゃん", // "覚えてくれ" + negation
		"正解じゃない",      // "正解" + negation
		"それそれじゃない",    // "それそれ" + negation
	}
	for _, msg := range cases {
		if got := DetectFeedback(msg, LangJapanese); got != FeedbackBoostNegative {
			t.Errorf("expected negative for negated phrase %q, got %v", msg, got)
		}
	}
}

// Regression: an explicit negative correction wins over a leading positive.
func TestDetectFeedback_NegativeWins(t *testing.T) {
	if got := DetectFeedback("ありがとう、でもそれじゃない", LangJapanese); got != FeedbackBoostNegative {
		t.Errorf("expected negative when correction follows thanks, got %v", got)
	}
}

func TestDetectFeedback_English(t *testing.T) {
	positives := []string{"Thanks!", "Exactly", "That's right", "you remembered"}
	for _, msg := range positives {
		if got := DetectFeedback(msg, LangEnglish); got != FeedbackBoostPositive {
			t.Errorf("expected positive for %q, got %v", msg, got)
		}
	}
	negatives := []string{"no that's wrong", "not right", "you forgot", "that's not it"}
	for _, msg := range negatives {
		if got := DetectFeedback(msg, LangEnglish); got != FeedbackBoostNegative {
			t.Errorf("expected negative for %q, got %v", msg, got)
		}
	}
	if got := DetectFeedback("what time is the meeting?", LangEnglish); got != 0 {
		t.Errorf("expected neutral, got %v", got)
	}
}
