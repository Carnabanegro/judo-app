package domain

import "github.com/google/uuid"

// PracticeMatchID is a type-safe practice match identifier.
type PracticeMatchID = uuid.UUID

// PracticeMatch represents a quick standalone bout outside any tournament.
// It reuses Combat for all timing and scoring logic.
type PracticeMatch struct {
	ID       PracticeMatchID
	LabelA   string  // athlete name or label (no registration required)
	LabelB   string
	Combat   *Combat
}

// NewPracticeMatch creates a new practice match with ad-hoc athlete labels.
func NewPracticeMatch(labelA, labelB string) *PracticeMatch {
	id := uuid.New()
	// Use a synthetic MatchID derived from the practice match ID.
	combat := NewCombat(id)
	return &PracticeMatch{
		ID:     id,
		LabelA: labelA,
		LabelB: labelB,
		Combat: combat,
	}
}
