package memai

import "testing"

func TestDetectFeedback_Positive(t *testing.T) {
	cases := []string{"ありがとう！", "そうそう、それ", "さすが"}
	for _, msg := range cases {
		if DetectFeedback(msg) != FeedbackBoostPositive {
			t.Errorf("expected positive for %q", msg)
		}
	}
}

func TestDetectFeedback_Negative(t *testing.T) {
	cases := []string{"違うよ", "それじゃない", "忘れてるよ"}
	for _, msg := range cases {
		if DetectFeedback(msg) != FeedbackBoostNegative {
			t.Errorf("expected negative for %q", msg)
		}
	}
}

func TestDetectFeedback_Neutral(t *testing.T) {
	if DetectFeedback("明日の天気は？") != 0 {
		t.Error("expected neutral feedback")
	}
}
