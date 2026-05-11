package application

import (
	"context"
	"errors"
	"time"

	"judo-app/internal/application/ports"
	"judo-app/internal/domain"
)

// TournamentService handles all tournament management use cases.
type TournamentService struct {
    tournaments ports.TournamentRepo
    divisions   ports.DivisionRepo
    categories  ports.CategoryRepo
    athletes    ports.AthleteRepo
    activeTournamentID *domain.TournamentID
}

// NewTournamentService creates a new TournamentService.
func NewTournamentService(
	tournaments ports.TournamentRepo,
	divisions ports.DivisionRepo,
	categories ports.CategoryRepo,
	athletes ports.AthleteRepo,
) *TournamentService {
	return &TournamentService{
		tournaments: tournaments,
		divisions:   divisions,
		categories:  categories,
		athletes:    athletes,
	}
}

// CreateTournament creates and persists a new tournament.
func (s *TournamentService) CreateTournament(ctx context.Context, name, location string, date time.Time) (*domain.Tournament, error) {
	t, err := domain.NewTournament(name, location, date)
	if err != nil {
		return nil, err
	}
	if err := s.tournaments.Save(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

// ListTournaments returns all tournaments.
func (s *TournamentService) ListTournaments(ctx context.Context) ([]*domain.Tournament, error) {
	return s.tournaments.List(ctx)
}

// GetTournament returns a tournament by ID.
func (s *TournamentService) GetTournament(ctx context.Context, id domain.TournamentID) (*domain.Tournament, error) {
	return s.tournaments.FindByID(ctx, id)
}

// DeleteTournament removes a tournament by ID.
func (s *TournamentService) DeleteTournament(ctx context.Context, id domain.TournamentID) error {
	return s.tournaments.Delete(ctx, id)
}

// CreateDivision creates a division within a tournament.
func (s *TournamentService) CreateDivision(ctx context.Context, tournamentID domain.TournamentID, ageGroup domain.AgeGroup, gender domain.Gender, weightClass string, format domain.Format) (*domain.Division, error) {
	// Verify tournament exists.
	if _, err := s.tournaments.FindByID(ctx, tournamentID); err != nil {
		return nil, errors.New("tournament not found")
	}
	d, err := domain.NewDivision(tournamentID, ageGroup, gender, weightClass, format)
	if err != nil {
		return nil, err
	}
	if err := s.divisions.Save(ctx, d); err != nil {
		return nil, err
	}
	return d, nil
}

// ListDivisions returns all divisions for a tournament.
func (s *TournamentService) ListDivisions(ctx context.Context, tournamentID domain.TournamentID) ([]*domain.Division, error) {
	return s.divisions.ListByTournament(ctx, tournamentID)
}

// CreateCategory creates a category within a division.
func (s *TournamentService) CreateCategory(ctx context.Context, divisionID domain.DivisionID) (*domain.Category, error) {
	cat := domain.NewCategory(divisionID)
	if err := s.categories.Save(ctx, cat); err != nil {
		return nil, err
	}
	return cat, nil
}

// ListCategories returns all categories for a division.
func (s *TournamentService) ListCategories(ctx context.Context, divisionID domain.DivisionID) ([]*domain.Category, error) {
	return s.categories.ListByDivision(ctx, divisionID)
}

// RegisterAthlete registers an athlete to a category.
func (s *TournamentService) RegisterAthlete(ctx context.Context, categoryID domain.CategoryID, name, club string, weight float64, birthDate time.Time) (*domain.Athlete, error) {
	a, err := domain.NewAthlete(name, club, weight, birthDate)
	if err != nil {
		return nil, err
	}
	if err := s.athletes.SaveToCategory(ctx, a, categoryID); err != nil {
		return nil, err
	}
	return a, nil
}

// ListAthletes returns athletes registered in a category.
func (s *TournamentService) ListAthletes(ctx context.Context, categoryID domain.CategoryID) ([]*domain.Athlete, error) {
    return s.athletes.ListByCategory(ctx, categoryID)
}

// SetActiveTournament marks a tournament as the current active one.
func (s *TournamentService) SetActiveTournament(id domain.TournamentID) {
    s.activeTournamentID = &id
}

// GetActiveTournament returns the active tournament, or nil if none is set.
func (s *TournamentService) GetActiveTournament(ctx context.Context) (*domain.Tournament, error) {
    if s.activeTournamentID == nil {
        return nil, nil
    }
    return s.tournaments.FindByID(ctx, *s.activeTournamentID)
}
