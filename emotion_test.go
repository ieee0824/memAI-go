package memai

import (
	"context"
	"testing"
)

func TestAnalyzeEmotion_Joy(t *testing.T) {
	es := AnalyzeEmotion("嬉しい！ありがとう！", LangJapanese)
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
	es := AnalyzeEmotion("悲しいなあ、辛い", LangJapanese)
	if es.Primary != EmotionSadness {
		t.Errorf("expected sadness, got %s", es.Primary)
	}
	if es.Valence >= 0 {
		t.Errorf("expected negative valence, got %f", es.Valence)
	}
}

func TestAnalyzeEmotion_Neutral(t *testing.T) {
	es := AnalyzeEmotion("明日の会議は何時から？", LangJapanese)
	if es.Primary != EmotionNeutral {
		t.Errorf("expected neutral, got %s", es.Primary)
	}
	if es.Intensity != 0 {
		t.Errorf("expected 0 intensity, got %f", es.Intensity)
	}
}

func TestAnalyzeEmotion_ExclamationBoost(t *testing.T) {
	without := AnalyzeEmotion("嬉しい", LangJapanese)
	with := AnalyzeEmotion("嬉しい！！！", LangJapanese)
	if with.Intensity <= without.Intensity {
		t.Errorf("exclamation should boost intensity: %f <= %f", with.Intensity, without.Intensity)
	}
}

func TestAnalyzeEmotion_IntensityCap(t *testing.T) {
	es := AnalyzeEmotion("嬉しい！楽しい！ありがとう！最高！素晴らしい！", LangJapanese)
	if es.Intensity > 1.0 {
		t.Errorf("intensity should not exceed 1.0, got %f", es.Intensity)
	}
}

func TestAnalyzeEmotion_EnglishJoy(t *testing.T) {
	es := AnalyzeEmotion("I'm so happy and excited! This is awesome!", LangEnglish)
	if es.Primary != EmotionJoy {
		t.Errorf("expected joy, got %s", es.Primary)
	}
	if es.Valence <= 0 {
		t.Errorf("expected positive valence, got %f", es.Valence)
	}
}

func TestAnalyzeEmotion_EnglishJoyCaseInsensitive(t *testing.T) {
	es := AnalyzeEmotion("I'm HAPPY and GRATEFUL!", LangEnglish)
	if es.Primary != EmotionJoy {
		t.Errorf("expected joy, got %s", es.Primary)
	}
}

func TestAnalyzeEmotion_EnglishSadness(t *testing.T) {
	es := AnalyzeEmotion("I feel so sad and lonely", LangEnglish)
	if es.Primary != EmotionSadness {
		t.Errorf("expected sadness, got %s", es.Primary)
	}
	if es.Valence >= 0 {
		t.Errorf("expected negative valence, got %f", es.Valence)
	}
}

func TestAnalyzeEmotion_EnglishNeutral(t *testing.T) {
	es := AnalyzeEmotion("What time is the meeting tomorrow?", LangEnglish)
	if es.Primary != EmotionNeutral {
		t.Errorf("expected neutral, got %s", es.Primary)
	}
	if es.Intensity != 0 {
		t.Errorf("expected 0 intensity, got %f", es.Intensity)
	}
}

func TestAnalyzeEmotion_EnglishAnger(t *testing.T) {
	es := AnalyzeEmotion("This is awful, I'm so angry and frustrated!", LangEnglish)
	if es.Primary != EmotionAnger {
		t.Errorf("expected anger, got %s", es.Primary)
	}
	if es.Valence >= 0 {
		t.Errorf("expected negative valence, got %f", es.Valence)
	}
}

func TestKeywordEmotionAnalyzer_ImplementsInterface(t *testing.T) {
	var _ EmotionAnalyzer = NewKeywordEmotionAnalyzer(LangJapanese)
}

func TestKeywordEmotionAnalyzer_Japanese(t *testing.T) {
	a := NewKeywordEmotionAnalyzer(LangJapanese)
	es, err := a.Analyze(context.Background(), "嬉しい！ありがとう！")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if es.Primary != EmotionJoy {
		t.Errorf("expected joy, got %s", es.Primary)
	}
}

func TestKeywordEmotionAnalyzer_English(t *testing.T) {
	a := NewKeywordEmotionAnalyzer(LangEnglish)
	es, err := a.Analyze(context.Background(), "I feel so sad and lonely")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if es.Primary != EmotionSadness {
		t.Errorf("expected sadness, got %s", es.Primary)
	}
}

// Regression (#7): a keyword-count tie must resolve to the same emotion on
// every run. "happy" (joy) vs "scared" (fear) is a 1-1 tie; joy must win
// deterministically via emotionOrder.
func TestAnalyzeEmotion_DeterministicTie(t *testing.T) {
	for i := 0; i < 1000; i++ {
		es := AnalyzeEmotion("happy scared", LangEnglish)
		if es.Primary != EmotionJoy {
			t.Fatalf("nondeterministic tie-break: got %s on iteration %d", es.Primary, i)
		}
	}
}

// Regression (#8): English keywords must match whole words, not substrings.
func TestAnalyzeEmotion_EnglishNoSubstringFalsePositive(t *testing.T) {
	cases := []string{
		"I made dinner",        // "mad" must not match "made"
		"the mission deadline", // "miss" must not match "mission"
		"I am fearless",        // "fear" must not match "fearless"
		"download the file",    // "down" must not match "download"
		"check the storage",    // "rage" must not match "storage"
	}
	for _, msg := range cases {
		es := AnalyzeEmotion(msg, LangEnglish)
		if es.Primary != EmotionNeutral {
			t.Errorf("expected neutral for %q, got %s (intensity %f)", msg, es.Primary, es.Intensity)
		}
	}
}

// Regression (#9): "really" is no longer a surprise keyword.
func TestAnalyzeEmotion_ReallyNotSurprise(t *testing.T) {
	es := AnalyzeEmotion("I really need this done.", LangEnglish)
	if es.Primary != EmotionNeutral {
		t.Errorf("expected neutral, got %s", es.Primary)
	}
}

// Regression (#10): 大丈夫 (reassurance) must not be classified as fear.
func TestAnalyzeEmotion_DaijobuNotFear(t *testing.T) {
	es := AnalyzeEmotion("もう大丈夫だよ", LangJapanese)
	if es.Primary == EmotionFear {
		t.Errorf("大丈夫 should not be fear, got %s with valence %f", es.Primary, es.Valence)
	}
}
