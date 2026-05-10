package application

import (
	"context"
	"errors"
	"sync"
	"time"

	"judo-app/internal/application/ports"
	"judo-app/internal/domain"
)

const tickInterval = 100 * time.Millisecond

// CombatService manages the lifecycle of an active match.
type CombatService struct {
	broadcaster ports.EventBroadcaster
	mu          sync.Mutex
	active      map[domain.MatchID]*combatSession
}

type combatSession struct {
	combat  *domain.Combat
	labelA  string
	labelB  string
	cancel  context.CancelFunc
}

// NewCombatService creates a new CombatService.
func NewCombatService(broadcaster ports.EventBroadcaster) *CombatService {
	return &CombatService{
		broadcaster: broadcaster,
		active:      make(map[domain.MatchID]*combatSession),
	}
}

// StartMatch begins a new match from an existing bracket match.
func (s *CombatService) StartMatch(ctx context.Context, match *domain.Match, labelA, labelB string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.active[match.ID]; exists {
		return errors.New("match already active")
	}

	combat := domain.NewCombat(match.ID)
	if err := combat.Start(); err != nil {
		return err
	}

	tickCtx, cancel := context.WithCancel(ctx)
	session := &combatSession{combat: combat, labelA: labelA, labelB: labelB, cancel: cancel}
	s.active[match.ID] = session

	go s.runTicker(tickCtx, match.ID)
	s.broadcastUpdate(match.ID, session)
	return nil
}

// Pause pauses the active match timer.
func (s *CombatService) Pause(matchID domain.MatchID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	sess, err := s.getSession(matchID)
	if err != nil {
		return err
	}
	if err := sess.combat.Pause(); err != nil {
		return err
	}
	s.broadcastUpdate(matchID, sess)
	return nil
}

// Resume resumes a paused match.
func (s *CombatService) Resume(matchID domain.MatchID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	sess, err := s.getSession(matchID)
	if err != nil {
		return err
	}
	if err := sess.combat.Start(); err != nil {
		return err
	}
	s.broadcastUpdate(matchID, sess)
	return nil
}

// Ippon records an ippon for the given athlete.
func (s *CombatService) Ippon(matchID domain.MatchID, athleteIdx int) error {
	return s.applyScore(matchID, func(c *domain.Combat) error {
		return c.ApplyIppon(athleteIdx)
	})
}

// WazaAri records a waza-ari for the given athlete.
func (s *CombatService) WazaAri(matchID domain.MatchID, athleteIdx int) error {
	return s.applyScore(matchID, func(c *domain.Combat) error {
		return c.ApplyWazaAri(athleteIdx)
	})
}

// Yuko records a yuko for the given athlete.
func (s *CombatService) Yuko(matchID domain.MatchID, athleteIdx int) error {
	return s.applyScore(matchID, func(c *domain.Combat) error {
		return c.ApplyYuko(athleteIdx)
	})
}

// Shido records a shido (penalty) for the given athlete.
func (s *CombatService) Shido(matchID domain.MatchID, athleteIdx int) error {
	return s.applyScore(matchID, func(c *domain.Combat) error {
		return c.ApplyShido(athleteIdx)
	})
}

// StartOsaekomi begins the osaekomi (hold-down) clock.
func (s *CombatService) StartOsaekomi(matchID domain.MatchID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	sess, err := s.getSession(matchID)
	if err != nil {
		return err
	}
	return sess.combat.StartOsaekomi(time.Now())
}

// StopOsaekomi ends the osaekomi clock and applies the appropriate score.
func (s *CombatService) StopOsaekomi(matchID domain.MatchID, athleteIdx int) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sess, err := s.getSession(matchID)
	if err != nil {
		return "", err
	}
	elapsed, err := sess.combat.StopOsaekomi(time.Now())
	if err != nil {
		return "", err
	}
	scored, err := sess.combat.ApplyOsaekomiScore(elapsed, athleteIdx)
	if err != nil {
		return "", err
	}
	s.checkAndFinish(matchID, sess)
	s.broadcastUpdate(matchID, sess)
	return scored, nil
}

// Finish forcefully ends a match (kiken-gachi / fusen-gachi).
func (s *CombatService) Finish(matchID domain.MatchID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	sess, err := s.getSession(matchID)
	if err != nil {
		return err
	}
	return s.finishSession(matchID, sess)
}

// --- internal ---

func (s *CombatService) runTicker(ctx context.Context, matchID domain.MatchID) {
	ticker := time.NewTicker(tickInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.mu.Lock()
			sess, ok := s.active[matchID]
			if !ok {
				s.mu.Unlock()
				return
			}
			toGS := sess.combat.Tick(tickInterval)
			s.broadcastTick(matchID, sess)
			if toGS {
				s.broadcastUpdate(matchID, sess)
			}
			s.checkAndFinish(matchID, sess)
			s.mu.Unlock()
		}
	}
}

func (s *CombatService) applyScore(matchID domain.MatchID, fn func(*domain.Combat) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	sess, err := s.getSession(matchID)
	if err != nil {
		return err
	}
	if err := fn(sess.combat); err != nil {
		return err
	}
	s.checkAndFinish(matchID, sess)
	s.broadcastUpdate(matchID, sess)
	return nil
}

func (s *CombatService) checkAndFinish(matchID domain.MatchID, sess *combatSession) {
	winnerIdx, method, decided := sess.combat.Winner()
	if !decided {
		return
	}
	_ = sess.combat.Finish()
	sess.cancel()
	delete(s.active, matchID)

	s.broadcaster.Broadcast(ports.DisplayEvent{
		Type: ports.EventCombatFinish,
		Payload: ports.CombatFinishPayload{
			MatchID:   matchID.String(),
			WinnerIdx: winnerIdx,
			Method:    string(method),
		},
	})
}

func (s *CombatService) finishSession(matchID domain.MatchID, sess *combatSession) error {
	if err := sess.combat.Finish(); err != nil {
		return err
	}
	sess.cancel()
	delete(s.active, matchID)
	return nil
}

func (s *CombatService) broadcastUpdate(matchID domain.MatchID, sess *combatSession) {
	s.broadcaster.Broadcast(ports.DisplayEvent{
		Type: ports.EventCombatUpdate,
		Payload: ports.CombatUpdatePayload{
			MatchID:  matchID.String(),
			ScoreA:   scoreToDTO(sess.combat.ScoreA),
			ScoreB:   scoreToDTO(sess.combat.ScoreB),
			State:    string(sess.combat.Timer.State),
			LabelA:   sess.labelA,
			LabelB:   sess.labelB,
		},
	})
}

func (s *CombatService) broadcastTick(matchID domain.MatchID, sess *combatSession) {
	var osaeMs int64
	if sess.combat.Timer.OsaekomiAt != nil {
		osaeMs = time.Since(*sess.combat.Timer.OsaekomiAt).Milliseconds()
	}
	s.broadcaster.Broadcast(ports.DisplayEvent{
		Type: ports.EventTimerTick,
		Payload: ports.TimerTickPayload{
			MatchID:     matchID.String(),
			RemainingMs: sess.combat.Timer.Remaining.Milliseconds(),
			OsaekomiMs:  osaeMs,
			State:       string(sess.combat.Timer.State),
			GoldenScore: sess.combat.Timer.GoldenScore,
		},
	})
}

func (s *CombatService) getSession(matchID domain.MatchID) (*combatSession, error) {
	sess, ok := s.active[matchID]
	if !ok {
		return nil, errors.New("no active match with id: " + matchID.String())
	}
	return sess, nil
}

func scoreToDTO(s domain.Score) ports.ScoreDTO {
	return ports.ScoreDTO{
		Ippon:   s.Ippon,
		WazaAri: s.WazaAri,
		Yuko:    s.Yuko,
		Shido:   s.Shido,
		Hansoku: s.Hansoku,
	}
}
