package ports

import (
	"context"

	"judo-app/internal/domain"
)

// MatchStatusRepo manages the mutable match state for multi-tatami coordination.
type MatchStatusRepo interface {
	// UpsertFromBracket syncs all matches from a bracket into the status table.
	UpsertFromBracket(ctx context.Context, b *domain.Bracket) error
	// ClaimMatch atomically marks a PENDING match as IN_PROGRESS for a tatami.
	// Returns ErrMatchNotAvailable if already taken.
	ClaimMatch(ctx context.Context, matchID domain.MatchID, tatamiID string) error
	// FinishMatch marks a match as FINISHED with its result.
	FinishMatch(ctx context.Context, matchID domain.MatchID, result *domain.MatchResult) error
	// ListByTournament returns all match rows for a tournament (with athlete/division info).
	ListByTournament(ctx context.Context, tournamentID domain.TournamentID) ([]MatchRow, error)
	// GetMatch returns a single match row by ID.
	GetMatch(ctx context.Context, matchID domain.MatchID) (*MatchRow, error)
}

// MatchRow is a flat projection for the tatami view — no domain reconstruction needed.
type MatchRow struct {
	ID           string
	CategoryID   string
	DivisionID   string
	WeightClass  string
	Gender       string
	AgeGroup     string
	Round        int
	Position     int
	IsRepechage  bool
	AthleteAID   string
	AthleteBID   string
	AthleteAName string
	AthleteAClub string
	AthleteBName string
	AthleteBClub string
	Status       string // PENDING | IN_PROGRESS | FINISHED
	TatamiID     string
	ResultJSON   string
}
