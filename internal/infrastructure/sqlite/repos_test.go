package sqlite_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"judo-app/internal/domain"
	infrasqlite "judo-app/internal/infrastructure/sqlite"
)

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := infrasqlite.Open(":memory:")
	if err != nil {
		t.Fatalf("open in-memory db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func TestTournamentRepo_SaveAndFind(t *testing.T) {
	db := openTestDB(t)
	repo := infrasqlite.NewTournamentRepo(db)
	ctx := context.Background()

	want, err := domain.NewTournament("Copa IJF 2025", "Buenos Aires", time.Now().Truncate(time.Second))
	if err != nil {
		t.Fatal(err)
	}

	if err := repo.Save(ctx, want); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := repo.FindByID(ctx, want.ID)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}

	if got.Name != want.Name {
		t.Errorf("Name: got %q want %q", got.Name, want.Name)
	}
	if got.Location != want.Location {
		t.Errorf("Location: got %q want %q", got.Location, want.Location)
	}
	if !got.Date.Equal(want.Date) {
		t.Errorf("Date: got %v want %v", got.Date, want.Date)
	}
}

func TestTournamentRepo_List(t *testing.T) {
	db := openTestDB(t)
	repo := infrasqlite.NewTournamentRepo(db)
	ctx := context.Background()

	for _, name := range []string{"T1", "T2", "T3"} {
		t, err := domain.NewTournament(name, "Somewhere", time.Now())
		if err != nil {
			panic(err)
		}
		_ = repo.Save(ctx, t)
	}

	list, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 3 {
		t.Errorf("expected 3 tournaments, got %d", len(list))
	}
}

func TestTournamentRepo_Delete(t *testing.T) {
	db := openTestDB(t)
	repo := infrasqlite.NewTournamentRepo(db)
	ctx := context.Background()

	tournament, _ := domain.NewTournament("Deletable", "X", time.Now())
	_ = repo.Save(ctx, tournament)
	_ = repo.Delete(ctx, tournament.ID)

	list, _ := repo.List(ctx)
	if len(list) != 0 {
		t.Errorf("expected 0 after delete, got %d", len(list))
	}
}

func TestAthleteRepo_SaveAndList(t *testing.T) {
	db := openTestDB(t)
	athleteRepo := infrasqlite.NewAthleteRepo(db)
	categoryRepo := infrasqlite.NewCategoryRepo(db)
	divisionRepo := infrasqlite.NewDivisionRepo(db)
	tournamentRepo := infrasqlite.NewTournamentRepo(db)
	ctx := context.Background()

	// Build the required hierarchy.
	tour, _ := domain.NewTournament("Tour", "BA", time.Now())
	_ = tournamentRepo.Save(ctx, tour)

	div, _ := domain.NewDivision(tour.ID, domain.AgeGroupSenior, domain.GenderMale, "-66kg", domain.FormatIndividualIJF)
	_ = divisionRepo.Save(ctx, div)

	cat := domain.NewCategory(div.ID)
	_ = categoryRepo.Save(ctx, cat)

	a1, _ := domain.NewAthlete("Keiji Suzuki", "Club A", 65.4, time.Now().AddDate(-25, 0, 0))
	a2, _ := domain.NewAthlete("Ilias Iliadis", "Club B", 65.8, time.Now().AddDate(-28, 0, 0))

	_ = athleteRepo.SaveToCategory(ctx, a1, cat.ID)
	_ = athleteRepo.SaveToCategory(ctx, a2, cat.ID)

	athletes, err := athleteRepo.ListByCategory(ctx, cat.ID)
	if err != nil {
		t.Fatalf("ListByCategory: %v", err)
	}
	if len(athletes) != 2 {
		t.Errorf("expected 2 athletes, got %d", len(athletes))
	}
}

func TestBracketRepo_SaveAndFind(t *testing.T) {
	db := openTestDB(t)
	bracketRepo := infrasqlite.NewBracketRepo(db)
	categoryRepo := infrasqlite.NewCategoryRepo(db)
	divisionRepo := infrasqlite.NewDivisionRepo(db)
	tournamentRepo := infrasqlite.NewTournamentRepo(db)
	ctx := context.Background()

	tour, _ := domain.NewTournament("Tour", "BA", time.Now())
	_ = tournamentRepo.Save(ctx, tour)

	div, _ := domain.NewDivision(tour.ID, domain.AgeGroupSenior, domain.GenderMale, "-73kg", domain.FormatIndividualIJF)
	_ = divisionRepo.Save(ctx, div)

	cat := domain.NewCategory(div.ID)
	_ = categoryRepo.Save(ctx, cat)

	athletes := make([]*domain.Athlete, 4)
	for i := range athletes {
		a, _ := domain.NewAthlete("Athlete", "Club", 72.0, time.Now().AddDate(-20, 0, 0))
		athletes[i] = a
	}

	bracket, err := domain.GenerateBracket(cat.ID, athletes)
	if err != nil {
		t.Fatalf("GenerateBracket: %v", err)
	}

	if err := bracketRepo.Save(ctx, bracket); err != nil {
		t.Fatalf("Save bracket: %v", err)
	}

	got, err := bracketRepo.FindByCategory(ctx, cat.ID)
	if err != nil {
		t.Fatalf("FindByCategory: %v", err)
	}

	if len(got.Rounds) != len(bracket.Rounds) {
		t.Errorf("rounds: got %d want %d", len(got.Rounds), len(bracket.Rounds))
	}
}
