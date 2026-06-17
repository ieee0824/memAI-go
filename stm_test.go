package memai

import (
	"sync"
	"testing"
)

func TestSTM_Decay(t *testing.T) {
	stm := NewSTM(DefaultSTMConfig())
	stm.SetItems([]*WorkingMemoryItem{
		{Topic: "A", Activation: 0.8, TurnCreated: 0, TurnAccessed: 0, Keywords: []string{"特定キーワード"}},
	})

	stm.Update(2, "zzz", nil)

	items := stm.Items()
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	// 2 turns * 0.15 decay = 0.3 decay → 0.8 - 0.3 = 0.5
	if items[0].Activation < 0.45 || items[0].Activation > 0.55 {
		t.Errorf("expected activation ~0.5, got %f", items[0].Activation)
	}
}

func TestSTM_EmotionalDecaySlower(t *testing.T) {
	stm := NewSTM(DefaultSTMConfig())
	stm.SetItems([]*WorkingMemoryItem{
		{Topic: "emotional", Activation: 0.8, TurnCreated: 0, TurnAccessed: 0, Emotional: true, Keywords: []string{"x"}},
		{Topic: "normal", Activation: 0.8, TurnCreated: 0, TurnAccessed: 0, Emotional: false, Keywords: []string{"y"}},
	})

	stm.Update(2, "unrelated", nil)

	items := stm.Items()
	var emotional, normal *WorkingMemoryItem
	for _, item := range items {
		if item.Topic == "emotional" {
			emotional = item
		}
		if item.Topic == "normal" {
			normal = item
		}
	}

	if emotional == nil || normal == nil {
		t.Fatal("expected both items to survive")
	}
	if emotional.Activation <= normal.Activation {
		t.Errorf("emotional item should decay slower: %f <= %f", emotional.Activation, normal.Activation)
	}
}

func TestSTM_Refresh(t *testing.T) {
	stm := NewSTM(DefaultSTMConfig())
	stm.SetItems([]*WorkingMemoryItem{
		{Topic: "weather", Activation: 0.3, TurnCreated: 0, TurnAccessed: 0, Keywords: []string{"天気"}},
	})

	stm.Update(0, "今日の天気は？", nil)

	items := stm.Items()
	if items[0].Activation < 0.5 {
		t.Errorf("keyword match should boost activation, got %f", items[0].Activation)
	}
}

func TestSTM_Eviction(t *testing.T) {
	stm := NewSTM(DefaultSTMConfig())
	stm.SetItems([]*WorkingMemoryItem{
		{Topic: "dead", Activation: 0.05, TurnCreated: 0, TurnAccessed: 0, Keywords: []string{"x"}},
		{Topic: "alive", Activation: 0.5, TurnCreated: 0, TurnAccessed: 0, Keywords: []string{"y"}},
	})

	stm.Update(0, "unrelated", nil)

	items := stm.Items()
	if len(items) != 1 {
		t.Fatalf("expected 1 item after eviction, got %d", len(items))
	}
	if items[0].Topic != "alive" {
		t.Errorf("wrong item survived: %s", items[0].Topic)
	}
}

func TestSTM_CapacityLimit(t *testing.T) {
	config := DefaultSTMConfig()
	config.MaxItems = 3
	stm := NewSTM(config)

	for i := 0; i < 5; i++ {
		stm.Add(&WorkingMemoryItem{
			Topic:      string(rune('A' + i)),
			Activation: float64(i+1) * 0.2,
			Keywords:   []string{string(rune('a' + i))},
		})
	}

	if len(stm.Items()) != 3 {
		t.Errorf("expected 3 items, got %d", len(stm.Items()))
	}
}

// Regression (#3): an emotional message must only flag topically relevant
// items, not every item in working memory.
func TestSTM_MarkEmotionalOnlyRelevant(t *testing.T) {
	stm := NewSTM(DefaultSTMConfig())
	stm.SetItems([]*WorkingMemoryItem{
		{Topic: "trip", Activation: 0.8, Keywords: []string{"旅行"}},
		{Topic: "logistics", Activation: 0.8, Keywords: []string{"会議"}},
	})

	// Emotional message that only relates to the "trip" item.
	stm.Update(0, "旅行が楽しみ", &EmotionalState{Primary: EmotionJoy, Intensity: 0.6})

	items := stm.Items()
	var trip, logistics *WorkingMemoryItem
	for _, it := range items {
		switch it.Topic {
		case "trip":
			trip = it
		case "logistics":
			logistics = it
		}
	}
	if trip == nil || logistics == nil {
		t.Fatal("expected both items to survive")
	}
	if !trip.Emotional {
		t.Error("relevant item should be marked emotional")
	}
	if logistics.Emotional {
		t.Error("unrelated item must NOT be marked emotional")
	}
}

// Regression (#5): a zero-value config must not silently wipe working memory.
func TestSTM_ZeroConfigDoesNotWipe(t *testing.T) {
	stm := NewSTM(STMConfig{})
	stm.Add(&WorkingMemoryItem{Topic: "A", Activation: 1.0, Keywords: []string{"a"}})
	if len(stm.Items()) != 1 {
		t.Fatalf("zero-value config wiped item; got %d items", len(stm.Items()))
	}
}

// Regression (#2): concurrent Update/Add must not race or panic.
func TestSTM_ConcurrentAccess(t *testing.T) {
	stm := NewSTM(DefaultSTMConfig())
	var wg sync.WaitGroup
	for g := 0; g < 8; g++ {
		wg.Add(1)
		go func(g int) {
			defer wg.Done()
			for i := 0; i < 200; i++ {
				stm.Add(&WorkingMemoryItem{Topic: "x", Activation: 0.9, Keywords: []string{"x"}})
				stm.Update(i, "x", &EmotionalState{Primary: EmotionJoy, Intensity: 0.6})
				_ = stm.Items()
			}
		}(g)
	}
	wg.Wait()
}
