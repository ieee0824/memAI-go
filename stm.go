package memai

import (
	"fmt"
	"sort"
	"strings"
)

// STMConfig configures short-term memory behavior.
type STMConfig struct {
	MaxItems            int     // Maximum working memory capacity (default: 7)
	ActivationThreshold float64 // Below this, items are evicted (default: 0.1)
	NormalDecayRate     float64 // Activation decay per turn (default: 0.15)
	EmotionalDecayRate  float64 // Decay for emotional items (default: 0.07)
	RefreshBoost        float64 // Activation boost on keyword match (default: 0.3)
}

// DefaultSTMConfig returns the default STM configuration based on
// cognitive science models of working memory.
func DefaultSTMConfig() STMConfig {
	return STMConfig{
		MaxItems:            7,
		ActivationThreshold: 0.1,
		NormalDecayRate:     0.15,
		EmotionalDecayRate:  0.07,
		RefreshBoost:        0.3,
	}
}

// STM manages short-term / working memory with activation-based decay.
// Emotional items decay at a slower rate, modeling the amygdala's influence
// on memory consolidation.
type STM struct {
	config STMConfig
	items  []*WorkingMemoryItem
}

// NewSTM creates a new short-term memory manager.
func NewSTM(config STMConfig) *STM {
	return &STM{config: config}
}

// Items returns the current working memory items.
func (s *STM) Items() []*WorkingMemoryItem {
	return s.items
}

// SetItems replaces the working memory contents.
func (s *STM) SetItems(items []*WorkingMemoryItem) {
	s.items = items
}

// Update performs a full STM cycle: decay, emotional marking, refresh, eviction.
func (s *STM) Update(turn int, message string, emotion *EmotionalState) {
	s.decay(turn)
	s.markEmotional(emotion)
	s.refresh(message)
	s.evict()
}

// Add inserts a new item into working memory, evicting the lowest-activation
// item if capacity is exceeded.
func (s *STM) Add(item *WorkingMemoryItem) {
	s.items = append(s.items, item)
	s.evict()
}

// Format returns a human-readable representation of the working memory state.
func (s *STM) Format() string {
	if len(s.items) == 0 {
		return ""
	}

	var lines []string
	for _, item := range s.items {
		level := activationLevel(item.Activation)
		lines = append(lines, fmt.Sprintf("- [%s] %s", level, item.Topic))
	}
	return strings.Join(lines, "\n")
}

// decay reduces activation of all items based on elapsed turns.
func (s *STM) decay(currentTurn int) {
	for _, item := range s.items {
		elapsed := currentTurn - item.TurnAccessed
		if elapsed <= 0 {
			continue
		}

		rate := s.config.NormalDecayRate
		if item.Emotional {
			rate = s.config.EmotionalDecayRate
		}

		item.Activation -= rate * float64(elapsed)
		if item.Activation < 0 {
			item.Activation = 0
		}
		item.TurnAccessed = currentTurn
	}
}

// markEmotional flags items as emotional when emotion intensity is high.
func (s *STM) markEmotional(emotion *EmotionalState) {
	if emotion == nil || emotion.Intensity <= 0.3 {
		return
	}
	for _, item := range s.items {
		item.Emotional = true
	}
}

// refresh boosts activation of items whose keywords match the message.
func (s *STM) refresh(message string) {
	lower := strings.ToLower(message)
	for _, item := range s.items {
		if itemMatchesMessage(item, lower) {
			item.Activation += s.config.RefreshBoost
			if item.Activation > 1.0 {
				item.Activation = 1.0
			}
		}
	}
}

// evict removes low-activation items and enforces capacity.
func (s *STM) evict() {
	// Remove below threshold
	alive := s.items[:0]
	for _, item := range s.items {
		if item.Activation >= s.config.ActivationThreshold {
			alive = append(alive, item)
		}
	}
	s.items = alive

	// Enforce capacity
	if len(s.items) > s.config.MaxItems {
		sort.Slice(s.items, func(i, j int) bool {
			return s.items[i].Activation > s.items[j].Activation
		})
		s.items = s.items[:s.config.MaxItems]
	}
}

// itemMatchesMessage checks if any of the item's keywords appear in the message.
func itemMatchesMessage(item *WorkingMemoryItem, lowerMsg string) bool {
	for _, kw := range item.Keywords {
		if strings.Contains(lowerMsg, strings.ToLower(kw)) {
			return true
		}
	}
	return false
}

// activationLevel returns a label for the activation level.
func activationLevel(activation float64) string {
	switch {
	case activation >= 0.7:
		return "高"
	case activation >= 0.4:
		return "中"
	default:
		return "低"
	}
}
