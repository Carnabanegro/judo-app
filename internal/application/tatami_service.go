package application

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"judo-app/internal/application/ports"
	"judo-app/internal/domain"
)

// TatamiService coordinates multi-tatami match claiming and result recording.
type TatamiService struct {
	matchStatus ports.MatchStatusRepo
	brackets    *BracketService
	combat      *CombatService
	broadcaster ports.EventBroadcaster
}

// NewTatamiService creates a new TatamiService.
func NewTatamiService(
	matchStatus ports.MatchStatusRepo,
	brackets *BracketService,
	combat *CombatService,
	broadcaster ports.EventBroadcaster,
) *TatamiService {
	return &TatamiService{
		matchStatus: matchStatus,
		brackets:    brackets,
		combat:      combat,
		broadcaster: broadcaster,
	}
}

// ListMatches returns all matches for a tournament with live status.
func (s *TatamiService) ListMatches(ctx context.Context, tournamentID domain.TournamentID) ([]ports.MatchRow, error) {
	return s.matchStatus.ListByTournament(ctx, tournamentID)
}

// ClaimMatch atomically claims a PENDING match for a tatami and starts the combat engine.
func (s *TatamiService) ClaimMatch(ctx context.Context, matchID domain.MatchID, tatamiID, labelA, labelB string) error {
	if tatamiID == "" {
		return errors.New("tatamiID must not be empty")
	}
	if err := s.matchStatus.ClaimMatch(ctx, matchID, tatamiID); err != nil {
		return err
	}
	if err := s.combat.StartMatchByID(ctx, matchID, labelA, labelB); err != nil {
		return err
	}
	s.broadcaster.Broadcast(ports.DisplayEvent{
		Type: EventBracketUpdate,
		Payload: BracketUpdatePayload{
			MatchID:  matchID.String(),
			Status:   string(domain.MatchInProgress),
			TatamiID: tatamiID,
		},
	})
	return nil
}

// RecordResult records the final result of a match, advances the bracket,
// and broadcasts the bracket update to all operators.
func (s *TatamiService) RecordResult(ctx context.Context, categoryID domain.CategoryID, matchID domain.MatchID, winnerIdx int, method domain.FinishMethod) error {
	if err := s.brackets.RecordResult(ctx, categoryID, matchID, winnerIdx, method); err != nil {
		return err
	}

	row, err := s.matchStatus.GetMatch(ctx, matchID)
	if err != nil {
		return err
	}

	winnerStrID := row.AthleteAID
	loserStrID := row.AthleteBID
	if winnerIdx == 1 {
		winnerStrID = row.AthleteBID
		loserStrID = row.AthleteAID
	}
	winnerID, err := uuid.Parse(winnerStrID)
	if err != nil {
		return err
	}
	loserID, err := uuid.Parse(loserStrID)
	if err != nil {
		return err
	}

	result := &domain.MatchResult{
		WinnerID: winnerID,
		LoserID:  loserID,
		Method:   method,
	}
	if err := s.matchStatus.FinishMatch(ctx, matchID, result); err != nil {
		return err
	}

	s.broadcaster.Broadcast(ports.DisplayEvent{
		Type: EventBracketUpdate,
		Payload: BracketUpdatePayload{
			MatchID:   matchID.String(),
			Status:    string(domain.MatchFinished),
			WinnerIdx: winnerIdx,
			Method:    string(method),
		},
	})
	return nil
}

// BracketUpdatePayload is the payload for bracket:update events.
type BracketUpdatePayload struct {
	MatchID   string `json:"matchId"`
	Status    string `json:"status"`
	TatamiID  string `json:"tatamiId,omitempty"`
	WinnerIdx int    `json:"winnerIdx,omitempty"`
	Method    string `json:"method,omitempty"`
}

// EventBracketUpdate is broadcast when any match status changes.
const EventBracketUpdate = "bracket:update"
