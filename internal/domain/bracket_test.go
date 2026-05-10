package domain_test

import (
	"testing"
	"time"

	"judo-app/internal/domain"
)

func makeAthletes(n int) []*domain.Athlete {
	athletes := make([]*domain.Athlete, n)
	for i := range athletes {
		a, _ := domain.NewAthlete(
			"Athlete "+string(rune('A'+i)),
			"Club",
			70.0,
			time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
		)
		athletes[i] = a
	}
	return athletes
}

func TestGenerateBracket_EmptyAthletes(t *testing.T) {
	_, err := domain.GenerateBracket(domain.CategoryID{}, nil)
	if err == nil {
		t.Fatal("expected error for empty athletes, got nil")
	}
}

func TestGenerateBracket_PowerOfTwo(t *testing.T) {
	cases := []struct {
		n         int
		wantRounds int
	}{
		{2, 1},
		{4, 2},
		{8, 3},
		{16, 4},
	}
	for _, tc := range cases {
		athletes := makeAthletes(tc.n)
		b, err := domain.GenerateBracket(domain.CategoryID{}, athletes)
		if err != nil {
			t.Fatalf("n=%d: unexpected error: %v", tc.n, err)
		}
		if got := len(b.Rounds); got != tc.wantRounds {
			t.Errorf("n=%d: want %d rounds, got %d", tc.n, tc.wantRounds, got)
		}
		// First round must have n/2 matches (no byes needed).
		wantMatches := tc.n / 2
		if got := len(b.Rounds[0].Matches); got != wantMatches {
			t.Errorf("n=%d: want %d matches in R1, got %d", tc.n, wantMatches, got)
		}
	}
}

func TestGenerateBracket_OddCount(t *testing.T) {
	// 5 athletes → padded to 8 → 3 rounds, 4 matches in R1 (3 real + 1 bye)
	athletes := makeAthletes(5)
	b, err := domain.GenerateBracket(domain.CategoryID{}, athletes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(b.Rounds) != 3 {
		t.Errorf("want 3 rounds, got %d", len(b.Rounds))
	}
	if len(b.Rounds[0].Matches) != 4 {
		t.Errorf("want 4 matches in R1, got %d", len(b.Rounds[0].Matches))
	}
}

func TestGenerateBracket_RepechagePools(t *testing.T) {
	// 8 athletes → 3 rounds → semifinal is round 2 (index 1) → 2 repechage pools.
	athletes := makeAthletes(8)
	b, err := domain.GenerateBracket(domain.CategoryID{}, athletes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := len(b.Repechage); got != 2 {
		t.Errorf("want 2 repechage pools, got %d", got)
	}
}

func TestBracket_AdvanceMatchNotFound(t *testing.T) {
	athletes := makeAthletes(4)
	b, _ := domain.GenerateBracket(domain.CategoryID{}, athletes)
	err := b.Advance(domain.MatchID{}, athletes[0], athletes[1], domain.FinishIppon, false)
	if err == nil {
		t.Fatal("expected error for unknown match ID")
	}
}

func TestBracket_AdvanceDuplicateResult(t *testing.T) {
	athletes := makeAthletes(4)
	b, _ := domain.GenerateBracket(domain.CategoryID{}, athletes)
	m := b.Rounds[0].Matches[0]
	// First advance — should succeed.
	_ = b.Advance(m.ID, athletes[0], athletes[1], domain.FinishIppon, false)
	// Second advance on same match — should fail.
	err := b.Advance(m.ID, athletes[0], athletes[1], domain.FinishIppon, false)
	if err == nil {
		t.Fatal("expected error when advancing already-decided match")
	}
}
