package ports

import (
	"context"

	"judo-app/internal/domain"
)

// TournamentRepo defines persistence operations for tournaments.
type TournamentRepo interface {
	Save(ctx context.Context, t *domain.Tournament) error
	FindByID(ctx context.Context, id domain.TournamentID) (*domain.Tournament, error)
	List(ctx context.Context) ([]*domain.Tournament, error)
	Delete(ctx context.Context, id domain.TournamentID) error
}

// DivisionRepo defines persistence operations for divisions.
type DivisionRepo interface {
	Save(ctx context.Context, d *domain.Division) error
	FindByID(ctx context.Context, id domain.DivisionID) (*domain.Division, error)
	ListByTournament(ctx context.Context, tournamentID domain.TournamentID) ([]*domain.Division, error)
	Delete(ctx context.Context, id domain.DivisionID) error
}

// CategoryRepo defines persistence operations for categories.
type CategoryRepo interface {
	Save(ctx context.Context, c *domain.Category) error
	FindByID(ctx context.Context, id domain.CategoryID) (*domain.Category, error)
	ListByDivision(ctx context.Context, divisionID domain.DivisionID) ([]*domain.Category, error)
}

// AthleteRepo defines persistence operations for athletes.
type AthleteRepo interface {
	Save(ctx context.Context, a *domain.Athlete) error
	SaveToCategory(ctx context.Context, a *domain.Athlete, categoryID domain.CategoryID) error
	FindByID(ctx context.Context, id domain.AthleteID) (*domain.Athlete, error)
	ListByCategory(ctx context.Context, categoryID domain.CategoryID) ([]*domain.Athlete, error)
	Delete(ctx context.Context, id domain.AthleteID) error
}

// MatchRepo defines persistence operations for matches.
type MatchRepo interface {
	Save(ctx context.Context, m *domain.Match) error
	FindByID(ctx context.Context, id domain.MatchID) (*domain.Match, error)
	ListByCategory(ctx context.Context, categoryID domain.CategoryID) ([]*domain.Match, error)
}

// BracketRepo defines persistence operations for brackets.
type BracketRepo interface {
	Save(ctx context.Context, b *domain.Bracket) error
	FindByCategory(ctx context.Context, categoryID domain.CategoryID) (*domain.Bracket, error)
}
