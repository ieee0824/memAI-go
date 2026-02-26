package memai

import "testing"

func TestAnalyzeEmotion_Joy(t *testing.T) {
	es := AnalyzeEmotion("嬉しい！ありがとう！")
	if es.Primary != EmotionJoy {
		t.Errorf("expected joy, got %s", es.Primary)
	}
	if es.Intensity < 0.5 {
		t.Errorf("expected intensity >= 0.5, got %f", es.Intensity)
	}
	if es.Valence <= 0 {
		t.Errorf("expected positive valence, got %f", es.Valence)
	}
}

func TestAnalyzeEmotion_Sadness(t *testing.T) {
	es := AnalyzeEmotion("悲しいなあ、辛い")
	if es.Primary != EmotionSadness {
		t.Errorf("expected sadness, got %s", es.Primary)
	}
	if es.Valence >= 0 {
		t.Errorf("expected negative valence, got %f", es.Valence)
	}
}

func TestAnalyzeEmotion_Neutral(t *testing.T) {
	es := AnalyzeEmotion("明日の会議は何時から？")
	if es.Primary != EmotionNeutral {
		t.Errorf("expected neutral, got %s", es.Primary)
	}
	if es.Intensity != 0 {
		t.Errorf("expected 0 intensity, got %f", es.Intensity)
	}
}

func TestAnalyzeEmotion_ExclamationBoost(t *testing.T) {
	without := AnalyzeEmotion("嬉しい")
	with := AnalyzeEmotion("嬉しい！！！")
	if with.Intensity <= without.Intensity {
		t.Errorf("exclamation should boost intensity: %f <= %f", with.Intensity, without.Intensity)
	}
}

func TestAnalyzeEmotion_IntensityCap(t *testing.T) {
	es := AnalyzeEmotion("嬉しい！楽しい！ありがとう！最高！素晴らしい！")
	if es.Intensity > 1.0 {
		t.Errorf("intensity should not exceed 1.0, got %f", es.Intensity)
	}
}
