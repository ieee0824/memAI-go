package memai

import (
	"sort"
	"strings"
	"sync"
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
//
// All exported methods are safe for concurrent use.
type STM struct {
	mu     sync.Mutex
	config STMConfig
	items  []*WorkingMemoryItem
}

// NewSTM creates a new short-term memory manager. Non-positive MaxItems and
// negative rate/boost/threshold fields are replaced with the DefaultSTMConfig
// values so that a zero-value or partially-filled config cannot silently
// wipe working memory or invert decay.
func NewSTM(config STMConfig) *STM {
	d := DefaultSTMConfig()
	if config.MaxItems <= 0 {
		config.MaxItems = d.MaxItems
	}
	if config.ActivationThreshold < 0 {
		config.ActivationThreshold = d.ActivationThreshold
	}
	if config.NormalDecayRate < 0 {
		config.NormalDecayRate = d.NormalDecayRate
	}
	if config.EmotionalDecayRate < 0 {
		config.EmotionalDecayRate = d.EmotionalDecayRate
	}
	if config.RefreshBoost < 0 {
		config.RefreshBoost = d.RefreshBoost
	}
	return &STM{config: config}
}

// Items returns a snapshot of the current working memory items. The returned
// slice is a copy, so appending to it does not affect internal state (the
// pointed-to items are shared).
func (s *STM) Items() []*WorkingMemoryItem {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]*WorkingMemoryItem, len(s.items))
	copy(out, s.items)
	return out
}

// SetItems replaces the working memory contents.
func (s *STM) SetItems(items []*WorkingMemoryItem) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items = items
}

// Update performs a full STM cycle: decay, emotional marking, refresh, eviction.
func (s *STM) Update(turn int, message string, emotion *EmotionalState) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.decay(turn)
	s.markEmotional(message, emotion)
	s.refresh(message)
	s.evict()
}

// Add inserts a new item into working memory, evicting the lowest-activation
// item if capacity is exceeded.
func (s *STM) Add(item *WorkingMemoryItem) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items = append(s.items, item)
	s.evict()
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

// markEmotional flags items as emotional when the current message carries
// emotion (intensity > 0.3). Only items topically relevant to the message
// (their keywords appear in it) are marked, so an emotional turn does not
// retroactively tag unrelated items in working memory. The flag is sticky:
// once set it persists, giving the item the slower EmotionalDecayRate.
func (s *STM) markEmotional(message string, emotion *EmotionalState) {
	if emotion == nil || emotion.Intensity <= 0.3 {
		return
	}
	lower := strings.ToLower(message)
	for _, item := range s.items {
		if itemMatchesMessage(item, lower) {
			item.Emotional = true
		}
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

	// Enforce capacity (MaxItems <= 0 means no capacity limit).
	if s.config.MaxItems > 0 && len(s.items) > s.config.MaxItems {
		sort.SliceStable(s.items, func(i, j int) bool {
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
