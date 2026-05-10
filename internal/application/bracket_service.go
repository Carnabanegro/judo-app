package application

import (
	"context"
	"errors"

	"judo-app/internal/application/ports"
	"judo-app/internal/domain"
)

// BracketService handles bracket generation and match advancement.
type BracketService struct {
	brackets   ports.BracketRepo
	matches    ports.MatchRepo
	athletes   ports.AthleteRepo
	categories ports.CategoryRepo
}

// NewBracketService creates a new BracketService.
func NewBracketService(
	brackets ports.BracketRepo,
	matches ports.MatchRepo,
	athletes ports.AthleteRepo,
	categories ports.CategoryRepo,
) *BracketService {
	return &BracketService{
		brackets:   brackets,
		matches:    matches,
		athletes:   athletes,
		categories: categories,
	}
}

// GenerateBracket generates a direct-elimination + repechage bracket for a category.
func (s *BracketService) GenerateBracket(ctx context.Context, categoryID domain.CategoryID) (*domain.Bracket, error) {
	athletes, err := s.athletes.ListByCategory(ctx, categoryID)
	if err != nil {
		return nil, err
	}
	if len(athletes) < 2 {
		return nil, errors.New("at least 2 athletes required to generate a bracket")
	}
	b, err := domain.GenerateBracket(categoryID, athletes)
	if err != nil {
		return nil, err
	}
	if err := s.brackets.Save(ctx, b); err != nil {
		return nil, err
	}
	return b, nil
}

// GetBracket retrieves the bracket for a category.
func (s *BracketService) GetBracket(ctx context.Context, categoryID domain.CategoryID) (*domain.Bracket, error) {
	return s.brackets.FindByCategory(ctx, categoryID)
}

// RecordResult records the result of a match and advances the bracket.
func (s *BracketService) RecordResult(ctx context.Context, categoryID domain.CategoryID, matchID domain.MatchID, winnerIdx int, method domain.FinishMethod) error {
	b, err := s.brackets.FindByCategory(ctx, categoryID)
	if err != nil {
		return err
	}

	match := s.findMatch(b, matchID)
	if match == nil {
		return errors.New("match not found in bracket")
	}

	winner, loser, err := s.athletesFromMatch(match, winnerIdx)
	if err != nil {
		return err
	}

	// Losers from semifinal round enter repechage.
	totalRounds := len(b.Rounds)
	isSemifinalOrLater := totalRounds >= 2 && match.Round >= totalRounds-1
	if err := b.Advance(matchID, winner, loser, method, isSemifinalOrLater); err != nil {
		return err
	}

	return s.brackets.Save(ctx, b)
}

// Classification returns the final IJF classification for a category.
func (s *BracketService) Classification(ctx context.Context, categoryID domain.CategoryID) ([]*domain.Athlete, error) {
	b, err := s.brackets.FindByCategory(ctx, categoryID)
	if err != nil {
		return nil, err
	}
	return b.Classification(), nil
}

func (s *BracketService) findMatch(b *domain.Bracket, id domain.MatchID) *domain.Match {
	for i := range b.Rounds {
		for j := range b.Rounds[i].Matches {
			if b.Rounds[i].Matches[j].ID == id {
				return b.Rounds[i].Matches[j]
			}
		}
	}
	return nil
}

func (s *BracketService) athletesFromMatch(m *domain.Match, winnerIdx int) (winner, loser *domain.Athlete, err error) {
	if m.AthleteA == nil || m.AthleteB == nil {
		return nil, nil, errors.New("match has a bye slot — cannot record result")
	}
	if winnerIdx == 0 {
		return m.AthleteA, m.AthleteB, nil
	} else if winnerIdx == 1 {
		return m.AthleteB, m.AthleteA, nil
	}
	return nil, nil, errors.New("winnerIdx must be 0 or 1")
}
