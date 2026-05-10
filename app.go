package main

import (
	"context"
	"fmt"

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

	tournaments *application.TournamentService
	brackets    *application.BracketService
	combat      *application.CombatService
	practice    *application.PracticeService
}

// Ensure display.Server satisfies the broadcaster port at compile time.
var _ ports.EventBroadcaster = (*display.Server)(nil)

// NewApp creates a new App application struct.
func NewApp() *App {
	return &App{}
}

// startup is called by the Wails runtime after the window is ready.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Infrastructure — SQLite
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

	// Infrastructure — Display WebSocket server (localhost:8080/ws)
	displayServer := display.NewServer(displayAddr)
	displayServer.Start(ctx)

	// Application layer
	a.tournaments = application.NewTournamentService(tournamentRepo, divisionRepo, categoryRepo, athleteRepo)
	a.brackets = application.NewBracketService(bracketRepo, matchRepo, athleteRepo, categoryRepo)
	a.combat = application.NewCombatService(displayServer)
	a.practice = application.NewPracticeService(a.combat)
}

// shutdown is called by the Wails runtime when the window closes.
func (a *App) shutdown(_ context.Context) {}

// ── Combat bindings ───────────────────────────────────────────────────────────

// StartMatch begins an existing bracket match identified by its UUID string.
func (a *App) StartMatch(matchID, labelA, labelB string) error {
	id, err := uuid.Parse(matchID)
	if err != nil {
		return fmt.Errorf("invalid matchId: %w", err)
	}
	return a.combat.StartMatchByID(a.ctx, id, labelA, labelB)
}

// Pause pauses the active combat timer.
func (a *App) Pause(matchID string) error {
	id, err := uuid.Parse(matchID)
	if err != nil {
		return err
	}
	return a.combat.Pause(id)
}

// Resume resumes a paused combat.
func (a *App) Resume(matchID string) error {
	id, err := uuid.Parse(matchID)
	if err != nil {
		return err
	}
	return a.combat.Resume(id)
}

// Ippon records an ippon for the given athlete (0=A, 1=B).
func (a *App) Ippon(matchID string, athleteIdx int) error {
	id, err := uuid.Parse(matchID)
	if err != nil {
		return err
	}
	return a.combat.Ippon(id, athleteIdx)
}

// WazaAri records a waza-ari.
func (a *App) WazaAri(matchID string, athleteIdx int) error {
	id, err := uuid.Parse(matchID)
	if err != nil {
		return err
	}
	return a.combat.WazaAri(id, athleteIdx)
}

// Yuko records a yuko.
func (a *App) Yuko(matchID string, athleteIdx int) error {
	id, err := uuid.Parse(matchID)
	if err != nil {
		return err
	}
	return a.combat.Yuko(id, athleteIdx)
}

// Shido records a shido (penalty).
func (a *App) Shido(matchID string, athleteIdx int) error {
	id, err := uuid.Parse(matchID)
	if err != nil {
		return err
	}
	return a.combat.Shido(id, athleteIdx)
}

// StartOsaekomi begins the osaekomi (hold-down) clock.
func (a *App) StartOsaekomi(matchID string) error {
	id, err := uuid.Parse(matchID)
	if err != nil {
		return err
	}
	return a.combat.StartOsaekomi(id)
}

// StopOsaekomi ends the osaekomi and returns the score applied ("IPPON", "WAZA_ARI", "YUKO", "").
func (a *App) StopOsaekomi(matchID string, athleteIdx int) (string, error) {
	id, err := uuid.Parse(matchID)
	if err != nil {
		return "", err
	}
	return a.combat.StopOsaekomi(id, athleteIdx)
}

// Finish forcefully ends a combat (kiken-gachi / fusen-gachi).
func (a *App) Finish(matchID string) error {
	id, err := uuid.Parse(matchID)
	if err != nil {
		return err
	}
	return a.combat.Finish(id)
}

// ── Practice bindings ─────────────────────────────────────────────────────────

// StartPractice begins a standalone practice match. Returns the match UUID string.
func (a *App) StartPractice(labelA, labelB string) (string, error) {
	pm, err := a.practice.StartPractice(a.ctx, labelA, labelB)
	if err != nil {
		return "", err
	}
	return pm.ID.String(), nil
}
