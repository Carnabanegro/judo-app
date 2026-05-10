package domain

import (
	"errors"
	"math"
	"time"

	"github.com/google/uuid"
)

// MatchID is a type-safe match identifier.
type MatchID = uuid.UUID

// MatchResult holds the outcome of a completed match.
type MatchResult struct {
	WinnerID AthleteID
	LoserID  AthleteID
	Method   FinishMethod
}

// MatchStatus represents the lifecycle state of a match.
type MatchStatus string

const (
	MatchPending    MatchStatus = "PENDING"
	MatchInProgress MatchStatus = "IN_PROGRESS"
	MatchFinished   MatchStatus = "FINISHED"
)

// FinishMethod describes how a match ended.
type FinishMethod string

const (
	FinishIppon              FinishMethod = "IPPON"
	FinishWazaAriAwasete     FinishMethod = "WAZA_ARI_AWASETE_IPPON"
	FinishHansokuMake        FinishMethod = "HANSOKU_MAKE"
	FinishKikenGachi         FinishMethod = "KIKEN_GACHI"
	FinishFusenGachi         FinishMethod = "FUSEN_GACHI"
	FinishGoldenScore        FinishMethod = "GOLDEN_SCORE"
)

// Match represents a single bout between two athletes.
type Match struct {
	ID           MatchID
	CategoryID   CategoryID
	Round        int
	Position     int
	AthleteA     *Athlete
	AthleteB     *Athlete
	Result       *MatchResult
	NextMatchID  *MatchID
	IsRepechage  bool
	RepechagePool int
	Status       MatchStatus // PENDING | IN_PROGRESS | FINISHED
	TatamiID     string      // "" = unclaimed
}

// Bracket holds the full direct-elimination + repechage structure for a category.
type Bracket struct {
	CategoryID CategoryID
	Rounds     []Round
	Repechage  []RepechagePool
}

// Round groups matches by elimination round.
type Round struct {
	Number  int
	Matches []*Match
}

// RepechagePool holds the two bronze-medal repechage matches.
type RepechagePool struct {
	SemifinalMatchID MatchID   // the semifinal whose loser enters this pool
	Matches          []*Match
}

// GenerateBracket creates a direct-elimination bracket for the given athletes.
// Odd counts are padded with byes (nil athlete = bye).
func GenerateBracket(categoryID CategoryID, athletes []*Athlete) (*Bracket, error) {
	if len(athletes) == 0 {
		return nil, errors.New("cannot generate bracket: no athletes provided")
	}

	size := nextPowerOfTwo(len(athletes))
	padded := make([]*Athlete, size)
	copy(padded, athletes)
	// remaining slots are nil (bye)

	bracket := &Bracket{
		CategoryID: categoryID,
	}

	// Build rounds bottom-up.
	// Round 1 has size/2 matches; subsequent rounds halve until final.
	matchesByRound := make([][]*Match, 0)
	currentSlots := padded

	roundNum := 1
	for len(currentSlots) > 1 {
		round := Round{Number: roundNum}
		var nextSlots []*Athlete

		for i := 0; i < len(currentSlots); i += 2 {
			m := &Match{
				ID:         uuid.New(),
				CategoryID: categoryID,
				Round:      roundNum,
				Position:   i / 2,
				AthleteA:   currentSlots[i],
				AthleteB:   currentSlots[i+1],
				Status:     MatchPending,
			}
			// Advance byes immediately.
			if m.AthleteA == nil || m.AthleteB == nil {
				winner := m.AthleteA
				if winner == nil {
					winner = m.AthleteB
				}
				m.Result = &MatchResult{Method: FinishFusenGachi}
				if winner != nil {
					m.Result.WinnerID = winner.ID
				}
				nextSlots = append(nextSlots, winner)
			} else {
				nextSlots = append(nextSlots, nil) // placeholder for winner
			}
			round.Matches = append(round.Matches, m)
		}

		matchesByRound = append(matchesByRound, round.Matches)
		bracket.Rounds = append(bracket.Rounds, round)
		currentSlots = nextSlots
		roundNum++
	}

	// Wire NextMatchID: each match points to the match in the next round
	// that its winner will feed.
	totalRounds := len(matchesByRound)
	for r := 0; r < totalRounds-1; r++ {
		for i, m := range matchesByRound[r] {
			nextMatch := matchesByRound[r+1][i/2]
			nextID := nextMatch.ID
			m.NextMatchID = &nextID
		}
	}

	// Identify semifinal round (second-to-last) and create repechage pools.
	if totalRounds >= 2 {
		semifinalRound := matchesByRound[totalRounds-2]
		for _, sfMatch := range semifinalRound {
			pool := RepechagePool{
				SemifinalMatchID: sfMatch.ID,
				Matches:          make([]*Match, 0),
			}
			bracket.Repechage = append(bracket.Repechage, pool)
		}
	}

	return bracket, nil
}

// Advance records the result of a match and returns the updated bracket.
// loserEntersRepechage should be true for quarterfinal-or-later losers.
func (b *Bracket) Advance(matchID MatchID, winner, loser *Athlete, method FinishMethod, loserEntersRepechage bool) error {
	m := b.findMatch(matchID)
	if m == nil {
		return errors.New("match not found in bracket")
	}
	if m.Result != nil {
		return errors.New("match already has a result")
	}

	m.Result = &MatchResult{
		WinnerID: winner.ID,
		LoserID:  loser.ID,
		Method:   method,
	}

	// Feed winner into next match.
	if m.NextMatchID != nil {
		next := b.findMatch(*m.NextMatchID)
		if next != nil {
			if next.AthleteA == nil {
				next.AthleteA = winner
			} else {
				next.AthleteB = winner
			}
		}
	}

	// Feed loser into repechage pool if applicable.
	if loserEntersRepechage {
		b.addToRepechage(matchID, loser)
	}

	return nil
}

// Classification returns athletes ordered 1st through last based on bracket results.
// Returns [gold, silver, bronze1, bronze2, ...] using IJF 1,2,3,3,5,5,7,7 scheme.
func (b *Bracket) Classification() []*Athlete {
	if len(b.Rounds) == 0 {
		return nil
	}
	// Final is the last match of the last round.
	finalRound := b.Rounds[len(b.Rounds)-1]
	if len(finalRound.Matches) == 0 {
		return nil
	}
	final := finalRound.Matches[0]
	if final.Result == nil {
		return nil
	}

	var result []*Athlete
	gold := b.athleteByID(final.Result.WinnerID)
	silver := b.athleteByID(final.Result.LoserID)
	result = append(result, gold, silver)

	// Bronze medal winners from repechage pools.
	for _, pool := range b.Repechage {
		for _, m := range pool.Matches {
			if m.Result != nil {
				bronze := b.athleteByID(m.Result.WinnerID)
				result = append(result, bronze)
			}
		}
	}

	return result
}

func (b *Bracket) findMatch(id MatchID) *Match {
	for i := range b.Rounds {
		for j := range b.Rounds[i].Matches {
			if b.Rounds[i].Matches[j].ID == id {
				return b.Rounds[i].Matches[j]
			}
		}
	}
	for i := range b.Repechage {
		for j := range b.Repechage[i].Matches {
			if b.Repechage[i].Matches[j].ID == id {
				return b.Repechage[i].Matches[j]
			}
		}
	}
	return nil
}

func (b *Bracket) addToRepechage(semifinalMatchID MatchID, loser *Athlete) {
	for i := range b.Repechage {
		if b.Repechage[i].SemifinalMatchID == semifinalMatchID {
			m := &Match{
				ID:            uuid.New(),
				CategoryID:    b.CategoryID,
				IsRepechage:   true,
				RepechagePool: i,
				AthleteA:      loser,
			}
			b.Repechage[i].Matches = append(b.Repechage[i].Matches, m)
			return
		}
	}
}

func (b *Bracket) athleteByID(id AthleteID) *Athlete {
	for _, r := range b.Rounds {
		for _, m := range r.Matches {
			if m.AthleteA != nil && m.AthleteA.ID == id {
				return m.AthleteA
			}
			if m.AthleteB != nil && m.AthleteB.ID == id {
				return m.AthleteB
			}
		}
	}
	return nil
}

func nextPowerOfTwo(n int) int {
	if n <= 1 {
		return 1
	}
	return int(math.Pow(2, math.Ceil(math.Log2(float64(n)))))
}

// NewRepechageBronzeMatch creates a bronze-medal match between two repechage survivors.
func NewRepechageBronzeMatch(categoryID CategoryID, poolIdx int, athleteA, athleteB *Athlete) *Match {
	return &Match{
		ID:            uuid.New(),
		CategoryID:    categoryID,
		IsRepechage:   true,
		RepechagePool: poolIdx,
		AthleteA:      athleteA,
		AthleteB:      athleteB,
		Round:         0,
		Status:        MatchPending,
	}
}

// ensure time is imported (used in future result timestamps)
var _ = time.Now
