package application

import (
	"context"

	"judo-app/internal/domain"
)

// PracticeService handles standalone practice matches.
type PracticeService struct {
	combat *CombatService
}

// NewPracticeService creates a new PracticeService.
func NewPracticeService(combat *CombatService) *PracticeService {
	return &PracticeService{combat: combat}
}

// StartPractice begins a quick standalone match with ad-hoc athlete labels.
func (s *PracticeService) StartPractice(ctx context.Context, labelA, labelB string) (*domain.PracticeMatch, error) {
	pm := domain.NewPracticeMatch(labelA, labelB)
	// Reuse CombatService with a synthetic match.
	syntheticMatch := &domain.Match{
		ID: pm.ID,
	}
	if err := s.combat.StartMatch(ctx, syntheticMatch, labelA, labelB); err != nil {
		return nil, err
	}
	return pm, nil
}

// Ippon records an ippon in the practice match.
func (s *PracticeService) Ippon(matchID domain.PracticeMatchID, athleteIdx int) error {
	return s.combat.Ippon(matchID, athleteIdx)
}

// WazaAri records a waza-ari in the practice match.
func (s *PracticeService) WazaAri(matchID domain.PracticeMatchID, athleteIdx int) error {
	return s.combat.WazaAri(matchID, athleteIdx)
}

// Yuko records a yuko in the practice match.
func (s *PracticeService) Yuko(matchID domain.PracticeMatchID, athleteIdx int) error {
	return s.combat.Yuko(matchID, athleteIdx)
}

// Shido records a shido in the practice match.
func (s *PracticeService) Shido(matchID domain.PracticeMatchID, athleteIdx int) error {
	return s.combat.Shido(matchID, athleteIdx)
}

// StartOsaekomi begins the osaekomi clock.
func (s *PracticeService) StartOsaekomi(matchID domain.PracticeMatchID) error {
	return s.combat.StartOsaekomi(matchID)
}

// StopOsaekomi ends the osaekomi clock and applies the score.
func (s *PracticeService) StopOsaekomi(matchID domain.PracticeMatchID, athleteIdx int) (string, error) {
	return s.combat.StopOsaekomi(matchID, athleteIdx)
}

// Pause pauses the practice match.
func (s *PracticeService) Pause(matchID domain.PracticeMatchID) error {
	return s.combat.Pause(matchID)
}

// Resume resumes the practice match.
func (s *PracticeService) Resume(matchID domain.PracticeMatchID) error {
	return s.combat.Resume(matchID)
}

// Finish ends the practice match manually.
func (s *PracticeService) Finish(matchID domain.PracticeMatchID) error {
	return s.combat.Finish(matchID)
}
