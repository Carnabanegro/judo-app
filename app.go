package main

import (
	"context"
	"fmt"
	"io/fs"
	"time"

	"judo-app/internal/application"
	"judo-app/internal/application/ports"
	"judo-app/internal/infrastructure/display"
	infrasqlite "judo-app/internal/infrastructure/sqlite"

	"github.com/google/uuid"
)

const displayAddr = ":8080"
const dbPath = "judo.db"

// App is the Wails application struct. All exported methods become Go bindings
// callable from the Angular frontend via window.go.main.App.*
type App struct {
	ctx context.Context
	spa fs.FS // embedded Angular SPA for the display server

	tournaments *application.TournamentService
	brackets    *application.BracketService
	combat      *application.CombatService
	practice    *application.PracticeService
	tatami      *application.TatamiService
}

// Ensure display.Server satisfies the broadcaster port at compile time.
var _ ports.EventBroadcaster = (*display.Server)(nil)

// NewApp creates a new App application struct.
func NewApp(spa fs.FS) *App {
	return &App{spa: spa}
}

// startup is called by the Wails runtime after the window is ready.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	db, err := infrasqlite.Open(dbPath)
	if err != nil {
		fmt.Println("ERROR: could not open database:", err)
		return
	}

	tournamentRepo := infrasqlite.NewTournamentRepo(db)
	divisionRepo := infrasqlite.NewDivisionRepo(db)
	categoryRepo := infrasqlite.NewCategoryRepo(db)
	athleteRepo := infrasqlite.NewAthleteRepo(db)
	bracketRepo := infrasqlite.NewBracketRepo(db)
	matchRepo := infrasqlite.NewMatchRepo(bracketRepo)
	matchStatusRepo := infrasqlite.NewMatchStatusRepo(db)

	displayServer := display.NewServer(displayAddr)

	a.tournaments = application.NewTournamentService(tournamentRepo, divisionRepo, categoryRepo, athleteRepo)
	a.brackets = application.NewBracketService(bracketRepo, matchRepo, matchStatusRepo, athleteRepo, categoryRepo)
	a.combat = application.NewCombatService(displayServer)
	a.practice = application.NewPracticeService(a.combat)
  a.tatami = application.NewTatamiService(matchStatusRepo, a.brackets, a.combat, displayServer)

  displayServer.SetTatamiService(a.tatami)
  displayServer.SetTournamentService(a.tournaments)
  if a.spa != nil {
		displayServer.SetSPA(a.spa)
	}
	displayServer.Start(ctx)
}

func (a *App) shutdown(_ context.Context) {}

// ── Tournament setup bindings ─────────────────────────────────────────────────

// CreateTournament creates a new tournament.
func (a *App) CreateTournament(name, location, dateISO string) (*TournamentDTO, error) {
	date, err := time.Parse(time.DateOnly, dateISO)
	if err != nil {
		return nil, fmt.Errorf("invalid date %q: use YYYY-MM-DD", dateISO)
	}
	t, err := a.tournaments.CreateTournament(a.ctx, name, location, date)
	if err != nil {
		return nil, err
	}
	return tournamentToDTO(t), nil
}

// ListTournaments returns all tournaments.
func (a *App) ListTournaments() ([]*TournamentDTO, error) {
	list, err := a.tournaments.ListTournaments(a.ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*TournamentDTO, len(list))
	for i, t := range list {
		out[i] = tournamentToDTO(t)
	}
	return out, nil
}

// CreateDivision creates a division inside a tournament.
func (a *App) CreateDivision(tournamentID, ageGroup, gender, weightClass, format string) (*DivisionDTO, error) {
	tID, err := uuid.Parse(tournamentID)
	if err != nil {
		return nil, fmt.Errorf("invalid tournamentID: %w", err)
	}
	d, err := a.tournaments.CreateDivision(a.ctx, tID,
		domain_AgeGroup(ageGroup), domain_Gender(gender), weightClass, domain_Format(format))
	if err != nil {
		return nil, err
	}
	return divisionToDTO(d), nil
}

// ListDivisions returns all divisions for a tournament.
func (a *App) ListDivisions(tournamentID string) ([]*DivisionDTO, error) {
	tID, err := uuid.Parse(tournamentID)
	if err != nil {
		return nil, fmt.Errorf("invalid tournamentID: %w", err)
	}
	list, err := a.tournaments.ListDivisions(a.ctx, tID)
	if err != nil {
		return nil, err
	}
	out := make([]*DivisionDTO, len(list))
	for i, d := range list {
		out[i] = divisionToDTO(d)
	}
	return out, nil
}

// RegisterAthlete registers an athlete to a division (category auto-created per division for V1).
func (a *App) RegisterAthlete(divisionID, name, club string, weight float64, birthDateISO string) (*AthleteDTO, error) {
	dID, err := uuid.Parse(divisionID)
	if err != nil {
		return nil, fmt.Errorf("invalid divisionID: %w", err)
	}
	birth, err := time.Parse(time.DateOnly, birthDateISO)
	if err != nil {
		return nil, fmt.Errorf("invalid birthDate %q: use YYYY-MM-DD", birthDateISO)
	}
	// Ensure the division has a category.
	cats, err := a.tournaments.ListCategories(a.ctx, dID)
	if err != nil {
		return nil, err
	}
	var catID uuid.UUID
	if len(cats) == 0 {
		cat, err := a.tournaments.CreateCategory(a.ctx, dID)
		if err != nil {
			return nil, err
		}
		catID = cat.ID
	} else {
		catID = cats[0].ID
	}
	athlete, err := a.tournaments.RegisterAthlete(a.ctx, catID, name, club, weight, birth)
	if err != nil {
		return nil, err
	}
	return athleteToDTO(athlete, catID.String()), nil
}

// ListAthletes returns all athletes for a division.
func (a *App) ListAthletes(divisionID string) ([]*AthleteDTO, error) {
	dID, err := uuid.Parse(divisionID)
	if err != nil {
		return nil, fmt.Errorf("invalid divisionID: %w", err)
	}
	cats, err := a.tournaments.ListCategories(a.ctx, dID)
	if err != nil {
		return nil, err
	}
	if len(cats) == 0 {
		return []*AthleteDTO{}, nil
	}
	list, err := a.tournaments.ListAthletes(a.ctx, cats[0].ID)
	if err != nil {
		return nil, err
	}
	out := make([]*AthleteDTO, len(list))
	for i, ath := range list {
		out[i] = athleteToDTO(ath, cats[0].ID.String())
	}
	return out, nil
}

// GenerateBracket generates the bracket for a division.
func (a *App) GenerateBracket(divisionID string) error {
	dID, err := uuid.Parse(divisionID)
	if err != nil {
		return fmt.Errorf("invalid divisionID: %w", err)
	}
	cats, err := a.tournaments.ListCategories(a.ctx, dID)
	if err != nil {
		return err
	}
	if len(cats) == 0 {
		return fmt.Errorf("no category found for division %s", divisionID)
	}
	_, err = a.brackets.GenerateBracket(a.ctx, cats[0].ID)
	return err
}

// GetBracket returns the bracket for a division.
func (a *App) GetBracket(divisionID string) (*BracketDTO, error) {
	dID, err := uuid.Parse(divisionID)
	if err != nil {
		return nil, fmt.Errorf("invalid divisionID: %w", err)
	}
	cats, err := a.tournaments.ListCategories(a.ctx, dID)
	if err != nil {
		return nil, err
	}
	if len(cats) == 0 {
		return nil, fmt.Errorf("no category found for division %s", divisionID)
	}
	b, err := a.brackets.GetBracket(a.ctx, cats[0].ID)
	if err != nil {
		return nil, err
	}
	return bracketToDTO(b), nil
}

// ── Tatami bindings ───────────────────────────────────────────────────────────

// ListMatches returns all matches for a tournament with live status/tatami info.
func (a *App) ListMatches(tournamentID string) ([]ports.MatchRow, error) {
    tID, err := uuid.Parse(tournamentID)
    if err != nil {
        return nil, fmt.Errorf("invalid tournamentID: %w", err)
    }
    return a.tatami.ListMatches(a.ctx, tID)
}

// SetActiveTournament sets the active tournament used by the display server.
func (a *App) SetActiveTournament(tournamentID string) error {
    id, err := uuid.Parse(tournamentID)
    if err != nil {
        return fmt.Errorf("invalid tournament ID: %w", err)
    }
    a.tournaments.SetActiveTournament(id)
    fmt.Println("[DEBUG] SetActiveTournament:", id)
    return nil
}

// GetActiveTournament returns the active tournament, or nil if none is set.
func (a *App) GetActiveTournament() (*TournamentDTO, error) {
    t, err := a.tournaments.GetActiveTournament(a.ctx)
    fmt.Printf("[DEBUG] GetActiveTournament: t=%v err=%v\n", t, err)
    if err != nil {
        return nil, err
    }
    if t == nil {
        return nil, nil
    }
    return tournamentToDTO(t), nil
}

// ClaimMatch atomically claims a PENDING match for a tatami and starts the combat.
func (a *App) ClaimMatch(matchID, tatamiID, labelA, labelB string) error {
	mID, err := uuid.Parse(matchID)
	if err != nil {
		return fmt.Errorf("invalid matchID: %w", err)
	}
	return a.tatami.ClaimMatch(a.ctx, mID, tatamiID, labelA, labelB)
}

// RecordMatchResult records the result of a match and advances the bracket.
func (a *App) RecordMatchResult(categoryID, matchID string, winnerIdx int, method string) error {
	cID, err := uuid.Parse(categoryID)
	if err != nil {
		return fmt.Errorf("invalid categoryID: %w", err)
	}
	mID, err := uuid.Parse(matchID)
	if err != nil {
		return fmt.Errorf("invalid matchID: %w", err)
	}
	return a.tatami.RecordResult(a.ctx, cID, mID, winnerIdx, domain_FinishMethod(method))
}

// ── Combat bindings (unchanged) ───────────────────────────────────────────────

func (a *App) StartMatch(matchID, labelA, labelB string) error {
	id, err := uuid.Parse(matchID)
	if err != nil {
		return fmt.Errorf("invalid matchId: %w", err)
	}
	return a.combat.StartMatchByID(a.ctx, id, labelA, labelB)
}

func (a *App) Pause(matchID string) error {
	id, err := uuid.Parse(matchID)
	if err != nil {
		return err
	}
	return a.combat.Pause(id)
}

func (a *App) Resume(matchID string) error {
	id, err := uuid.Parse(matchID)
	if err != nil {
		return err
	}
	return a.combat.Resume(id)
}

func (a *App) Ippon(matchID string, athleteIdx int) error {
	id, err := uuid.Parse(matchID)
	if err != nil {
		return err
	}
	return a.combat.Ippon(id, athleteIdx)
}

func (a *App) WazaAri(matchID string, athleteIdx int) error {
	id, err := uuid.Parse(matchID)
	if err != nil {
		return err
	}
	return a.combat.WazaAri(id, athleteIdx)
}

func (a *App) Yuko(matchID string, athleteIdx int) error {
	id, err := uuid.Parse(matchID)
	if err != nil {
		return err
	}
	return a.combat.Yuko(id, athleteIdx)
}

func (a *App) Shido(matchID string, athleteIdx int) error {
	id, err := uuid.Parse(matchID)
	if err != nil {
		return err
	}
	return a.combat.Shido(id, athleteIdx)
}

func (a *App) StartOsaekomi(matchID string) error {
	id, err := uuid.Parse(matchID)
	if err != nil {
		return err
	}
	return a.combat.StartOsaekomi(id)
}

func (a *App) StopOsaekomi(matchID string, athleteIdx int) (string, error) {
	id, err := uuid.Parse(matchID)
	if err != nil {
		return "", err
	}
	return a.combat.StopOsaekomi(id, athleteIdx)
}

func (a *App) Finish(matchID string) error {
	id, err := uuid.Parse(matchID)
	if err != nil {
		return err
	}
	return a.combat.Finish(id)
}

func (a *App) StartPractice(labelA, labelB string) (string, error) {
	pm, err := a.practice.StartPractice(a.ctx, labelA, labelB)
	if err != nil {
		return "", err
	}
	return pm.ID.String(), nil
}
