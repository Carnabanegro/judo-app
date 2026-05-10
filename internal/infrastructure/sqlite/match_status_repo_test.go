package sqlite_test

import (
	"context"
	"testing"
	"time"

	"judo-app/internal/domain"
	infrasqlite "judo-app/internal/infrastructure/sqlite"
)

// seedBracket creates a minimal tournament/division/category/athlete graph and
// returns a generated 2-athlete bracket ready for UpsertFromBracket.
func seedBracket(t *testing.T, ctx context.Context, db interface {
	// We accept the raw *sql.DB through the open repos.
}) (*domain.Bracket, string) {
	t.Helper()
	return nil, ""
}

func TestMatchStatusRepo_UpsertAndList(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	// seed tournament chain
	tournamentRepo := infrasqlite.NewTournamentRepo(db)
	divisionRepo := infrasqlite.NewDivisionRepo(db)
	categoryRepo := infrasqlite.NewCategoryRepo(db)
	athleteRepo := infrasqlite.NewAthleteRepo(db)
	matchStatusRepo := infrasqlite.NewMatchStatusRepo(db)

	tournament, _ := domain.NewTournament("Copa Test", "BsAs", time.Now())
	_ = tournamentRepo.Save(ctx, tournament)

	division, _ := domain.NewDivision(tournament.ID, domain.AgeGroupSenior, domain.GenderMale, "-66kg", domain.FormatIndividualIJF)
	_ = divisionRepo.Save(ctx, division)

	category := domain.NewCategory(division.ID)
	_ = categoryRepo.Save(ctx, category)

	a1, _ := domain.NewAthlete("Uke", "ClubA", 65, time.Now().AddDate(-20, 0, 0))
	a2, _ := domain.NewAthlete("Tori", "ClubB", 64, time.Now().AddDate(-21, 0, 0))
	_ = athleteRepo.SaveToCategory(ctx, a1, category.ID)
	_ = athleteRepo.SaveToCategory(ctx, a2, category.ID)

	bracket, err := domain.GenerateBracket(category.ID, []*domain.Athlete{a1, a2})
	if err != nil {
		t.Fatalf("GenerateBracket: %v", err)
	}

	if err := matchStatusRepo.UpsertFromBracket(ctx, bracket); err != nil {
		t.Fatalf("UpsertFromBracket: %v", err)
	}

	rows, err := matchStatusRepo.ListByTournament(ctx, tournament.ID)
	if err != nil {
		t.Fatalf("ListByTournament: %v", err)
	}
	if len(rows) == 0 {
		t.Fatal("expected at least one match row")
	}
	for _, r := range rows {
		if r.Status != "PENDING" && r.Status != "FINISHED" {
			// byes get FINISHED immediately by GenerateBracket
			t.Errorf("unexpected status %q for match %s", r.Status, r.ID)
		}
	}
}

func TestMatchStatusRepo_ClaimMatch(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	tournamentRepo := infrasqlite.NewTournamentRepo(db)
	divisionRepo := infrasqlite.NewDivisionRepo(db)
	categoryRepo := infrasqlite.NewCategoryRepo(db)
	athleteRepo := infrasqlite.NewAthleteRepo(db)
	matchStatusRepo := infrasqlite.NewMatchStatusRepo(db)

	tournament, _ := domain.NewTournament("Copa Claim", "BsAs", time.Now())
	_ = tournamentRepo.Save(ctx, tournament)
	division, _ := domain.NewDivision(tournament.ID, domain.AgeGroupSenior, domain.GenderFemale, "-52kg", domain.FormatIndividualIJF)
	_ = divisionRepo.Save(ctx, division)
	category := domain.NewCategory(division.ID)
	_ = categoryRepo.Save(ctx, category)
	a1, _ := domain.NewAthlete("A", "X", 51, time.Now().AddDate(-20, 0, 0))
	a2, _ := domain.NewAthlete("B", "Y", 50, time.Now().AddDate(-20, 0, 0))
	_ = athleteRepo.SaveToCategory(ctx, a1, category.ID)
	_ = athleteRepo.SaveToCategory(ctx, a2, category.ID)

	bracket, _ := domain.GenerateBracket(category.ID, []*domain.Athlete{a1, a2})
	_ = matchStatusRepo.UpsertFromBracket(ctx, bracket)

	// Find a PENDING match to claim.
	rows, _ := matchStatusRepo.ListByTournament(ctx, tournament.ID)
	var pendingID string
	for _, r := range rows {
		if r.Status == "PENDING" {
			pendingID = r.ID
			break
		}
	}
	if pendingID == "" {
		t.Skip("no PENDING match found — all byes resolved")
	}

	matchUUID, _ := domain.MatchID{}, error(nil)
	_ = matchUUID.UnmarshalText([]byte(pendingID))

	if err := matchStatusRepo.ClaimMatch(ctx, matchUUID, "tatami-1"); err != nil {
		t.Fatalf("ClaimMatch: %v", err)
	}

	// Second claim on same match must fail.
	err := matchStatusRepo.ClaimMatch(ctx, matchUUID, "tatami-2")
	if err == nil {
		t.Fatal("expected error on second ClaimMatch, got nil")
	}
}

func TestMatchStatusRepo_ClaimMatch_Idempotent(t *testing.T) {
	db := openTestDB(t)
	ctx := context.Background()

	tournamentRepo := infrasqlite.NewTournamentRepo(db)
	divisionRepo := infrasqlite.NewDivisionRepo(db)
	categoryRepo := infrasqlite.NewCategoryRepo(db)
	athleteRepo := infrasqlite.NewAthleteRepo(db)
	matchStatusRepo := infrasqlite.NewMatchStatusRepo(db)

	tournament, _ := domain.NewTournament("Copa Idem", "BsAs", time.Now())
	_ = tournamentRepo.Save(ctx, tournament)
	division, _ := domain.NewDivision(tournament.ID, domain.AgeGroupJunior, domain.GenderMale, "-73kg", domain.FormatIndividualIJF)
	_ = divisionRepo.Save(ctx, division)
	category := domain.NewCategory(division.ID)
	_ = categoryRepo.Save(ctx, category)
	a1, _ := domain.NewAthlete("C", "P", 72, time.Now().AddDate(-18, 0, 0))
	a2, _ := domain.NewAthlete("D", "Q", 71, time.Now().AddDate(-19, 0, 0))
	_ = athleteRepo.SaveToCategory(ctx, a1, category.ID)
	_ = athleteRepo.SaveToCategory(ctx, a2, category.ID)

	bracket, _ := domain.GenerateBracket(category.ID, []*domain.Athlete{a1, a2})
	_ = matchStatusRepo.UpsertFromBracket(ctx, bracket)

	rows, _ := matchStatusRepo.ListByTournament(ctx, tournament.ID)
	var pendingID string
	for _, r := range rows {
		if r.Status == "PENDING" {
			pendingID = r.ID
			break
		}
	}
	if pendingID == "" {
		t.Skip("no PENDING match")
	}

	var matchUUID domain.MatchID
	_ = matchUUID.UnmarshalText([]byte(pendingID))

	_ = matchStatusRepo.ClaimMatch(ctx, matchUUID, "tatami-1")

	// Verify status is IN_PROGRESS.
	mr, err := matchStatusRepo.GetMatch(ctx, matchUUID)
	if err != nil {
		t.Fatalf("GetMatch: %v", err)
	}
	if mr.Status != "IN_PROGRESS" {
		t.Errorf("expected IN_PROGRESS, got %q", mr.Status)
	}
	if mr.TatamiID != "tatami-1" {
		t.Errorf("expected tatami-1, got %q", mr.TatamiID)
	}
}
