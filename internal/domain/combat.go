package domain

import (
	"errors"
	"time"
)

// CombatState represents the lifecycle state of a match.
type CombatState string

const (
	StatePending     CombatState = "PENDING"
	StateActive      CombatState = "ACTIVE"
	StatePaused      CombatState = "PAUSED"
	StateGoldenScore CombatState = "GOLDEN_SCORE"
	StateFinished    CombatState = "FINISHED"
)

// MatchDuration is the regulation combat time per IJF rules.
const MatchDuration = 4 * time.Minute

// OsaekomiYuko is the minimum osaekomi time for a yuko score.
const OsaekomiYuko = 5 * time.Second

// OsaekomiWazaAri is the minimum osaekomi time for a waza-ari score.
const OsaekomiWazaAri = 10 * time.Second

// OsaekomiIppon is the osaekomi time for an ippon score.
const OsaekomiIppon = 20 * time.Second

// MaxShido is the number of shidos that result in hansoku-make.
const MaxShido = 3

// Score holds the technical scores for one athlete in a match.
type Score struct {
	Ippon   int
	WazaAri int
	Yuko    int
	Shido   int
	Hansoku bool
}

// HasIppon returns true if this score constitutes an ippon win.
func (s Score) HasIppon() bool {
	return s.Ippon >= 1 || s.WazaAri >= 2 || s.Hansoku
}

// CombatTimer tracks match time and osaekomi for a single bout.
type CombatTimer struct {
	State       CombatState
	Remaining   time.Duration // counts down from MatchDuration; in GoldenScore counts up from 0
	OsaekomiAt  *time.Time   // non-nil when osaekomi is active
	GoldenScore bool
}

// NewCombatTimer creates a timer ready to start.
func NewCombatTimer() *CombatTimer {
	return &CombatTimer{
		State:     StatePending,
		Remaining: MatchDuration,
	}
}

// OsaekomiElapsed returns elapsed osaekomi duration, or 0 if not active.
func (t *CombatTimer) OsaekomiElapsed(now time.Time) time.Duration {
	if t.OsaekomiAt == nil {
		return 0
	}
	return now.Sub(*t.OsaekomiAt)
}

// Combat holds the full state of an active bout.
type Combat struct {
	MatchID  MatchID
	ScoreA   Score // AthleteA score
	ScoreB   Score // AthleteB score
	Timer    *CombatTimer
}

// NewCombat initialises a combat for the given match.
func NewCombat(matchID MatchID) *Combat {
	return &Combat{
		MatchID: matchID,
		Timer:   NewCombatTimer(),
	}
}

// Start transitions the combat from PENDING or PAUSED to ACTIVE.
func (c *Combat) Start() error {
	switch c.Timer.State {
	case StatePending, StatePaused:
		c.Timer.State = StateActive
		return nil
	default:
		return errors.New("combat cannot be started in state: " + string(c.Timer.State))
	}
}

// Pause transitions from ACTIVE or GOLDEN_SCORE to PAUSED.
func (c *Combat) Pause() error {
	switch c.Timer.State {
	case StateActive, StateGoldenScore:
		c.Timer.State = StatePaused
		c.Timer.OsaekomiAt = nil // stop osaekomi on pause
		return nil
	default:
		return errors.New("combat cannot be paused in state: " + string(c.Timer.State))
	}
}

// Tick advances the timer by delta. Returns true if the match ended (time expired and tied).
// Caller is responsible for invoking this on each tick interval.
func (c *Combat) Tick(delta time.Duration) (transitionedToGoldenScore bool) {
	if c.Timer.State != StateActive {
		return false
	}
	if c.Timer.GoldenScore {
		c.Timer.Remaining += delta
		return false
	}
	c.Timer.Remaining -= delta
	if c.Timer.Remaining <= 0 {
		c.Timer.Remaining = 0
		if !c.isTechnicallyDecided() {
			c.Timer.State = StateGoldenScore
			c.Timer.GoldenScore = true
			return true
		}
	}
	return false
}

// StartOsaekomi begins the osaekomi (hold-down) clock at the given time.
func (c *Combat) StartOsaekomi(now time.Time) error {
	if c.Timer.State != StateActive && c.Timer.State != StateGoldenScore {
		return errors.New("osaekomi can only be started when combat is active")
	}
	if c.Timer.OsaekomiAt != nil {
		return errors.New("osaekomi already active")
	}
	c.Timer.OsaekomiAt = &now
	return nil
}

// StopOsaekomi ends the osaekomi clock and returns the elapsed duration.
func (c *Combat) StopOsaekomi(now time.Time) (time.Duration, error) {
	if c.Timer.OsaekomiAt == nil {
		return 0, errors.New("osaekomi is not active")
	}
	elapsed := now.Sub(*c.Timer.OsaekomiAt)
	c.Timer.OsaekomiAt = nil
	return elapsed, nil
}

// ApplyOsaekomiScore applies the appropriate score for the given osaekomi duration.
// Returns the score type applied, or empty string if below yuko threshold.
func (c *Combat) ApplyOsaekomiScore(elapsed time.Duration, athleteIdx int) (string, error) {
	score, err := c.scoreFor(athleteIdx)
	if err != nil {
		return "", err
	}
	switch {
	case elapsed >= OsaekomiIppon:
		score.Ippon++
		c.setScore(athleteIdx, *score)
		return "IPPON", nil
	case elapsed >= OsaekomiWazaAri:
		score.WazaAri++
		c.setScore(athleteIdx, *score)
		return "WAZA_ARI", nil
	case elapsed >= OsaekomiYuko:
		score.Yuko++
		c.setScore(athleteIdx, *score)
		return "YUKO", nil
	default:
		return "", nil
	}
}

// ApplyIppon records an ippon for the given athlete (0=A, 1=B).
func (c *Combat) ApplyIppon(athleteIdx int) error {
	score, err := c.scoreFor(athleteIdx)
	if err != nil {
		return err
	}
	score.Ippon++
	c.setScore(athleteIdx, *score)
	return nil
}

// ApplyWazaAri records a waza-ari for the given athlete.
func (c *Combat) ApplyWazaAri(athleteIdx int) error {
	score, err := c.scoreFor(athleteIdx)
	if err != nil {
		return err
	}
	score.WazaAri++
	c.setScore(athleteIdx, *score)
	return nil
}

// ApplyYuko records a yuko for the given athlete.
func (c *Combat) ApplyYuko(athleteIdx int) error {
	score, err := c.scoreFor(athleteIdx)
	if err != nil {
		return err
	}
	score.Yuko++
	c.setScore(athleteIdx, *score)
	return nil
}

// ApplyShido records a shido (penalty) for the given athlete.
// If this is the 3rd shido, sets Hansoku = true (hansoku-make).
func (c *Combat) ApplyShido(athleteIdx int) error {
	score, err := c.scoreFor(athleteIdx)
	if err != nil {
		return err
	}
	score.Shido++
	if score.Shido >= MaxShido {
		score.Hansoku = true
	}
	c.setScore(athleteIdx, *score)
	return nil
}

// Winner returns the index of the winning athlete (0=A, 1=B) and the finish method,
// or -1 if the match is not yet decided.
func (c *Combat) Winner() (athleteIdx int, method FinishMethod, decided bool) {
	aWins := c.ScoreA.HasIppon()
	bWins := c.ScoreB.HasIppon()

	switch {
	case aWins && c.ScoreA.Hansoku:
		// A's hansoku means B wins.
		return 1, FinishHansokuMake, true
	case bWins && c.ScoreB.Hansoku:
		return 0, FinishHansokuMake, true
	case c.ScoreA.Ippon >= 1:
		return 0, FinishIppon, true
	case c.ScoreB.Ippon >= 1:
		return 1, FinishIppon, true
	case c.ScoreA.WazaAri >= 2:
		return 0, FinishWazaAriAwasete, true
	case c.ScoreB.WazaAri >= 2:
		return 1, FinishWazaAriAwasete, true
	}
	return -1, "", false
}

// Finish marks the combat as FINISHED. Must be called after Winner() returns decided=true.
func (c *Combat) Finish() error {
	if c.Timer.State == StateFinished {
		return errors.New("combat is already finished")
	}
	c.Timer.State = StateFinished
	c.Timer.OsaekomiAt = nil
	return nil
}

func (c *Combat) isTechnicallyDecided() bool {
	_, _, decided := c.Winner()
	return decided
}

func (c *Combat) scoreFor(idx int) (*Score, error) {
	switch idx {
	case 0:
		s := c.ScoreA
		return &s, nil
	case 1:
		s := c.ScoreB
		return &s, nil
	default:
		return nil, errors.New("athlete index must be 0 or 1")
	}
}

func (c *Combat) setScore(idx int, s Score) {
	switch idx {
	case 0:
		c.ScoreA = s
	case 1:
		c.ScoreB = s
	}
}
