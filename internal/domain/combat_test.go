package domain_test

import (
	"testing"
	"time"

	"judo-app/internal/domain"
)

func newCombat() *domain.Combat {
	return domain.NewCombat(domain.MatchID{})
}

func TestCombat_StartFromPending(t *testing.T) {
	c := newCombat()
	if err := c.Start(); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if c.Timer.State != domain.StateActive {
		t.Errorf("want ACTIVE, got %s", c.Timer.State)
	}
}

func TestCombat_CannotStartFromFinished(t *testing.T) {
	c := newCombat()
	_ = c.Start()
	_ = c.ApplyIppon(0)
	_ = c.Finish()
	if err := c.Start(); err == nil {
		t.Fatal("expected error starting finished combat")
	}
}

func TestCombat_ThirdShidoIsHansokuMake(t *testing.T) {
	c := newCombat()
	_ = c.Start()
	for i := 0; i < 3; i++ {
		_ = c.ApplyShido(0)
	}
	if !c.ScoreA.Hansoku {
		t.Error("expected Hansoku=true after 3 shidos")
	}
	idx, method, decided := c.Winner()
	if !decided {
		t.Fatal("expected match decided after hansoku-make")
	}
	if idx != 1 {
		t.Errorf("expected athlete B (idx 1) to win, got %d", idx)
	}
	if method != domain.FinishHansokuMake {
		t.Errorf("expected HANSOKU_MAKE, got %s", method)
	}
}

func TestCombat_TwoWazaAriIsIppon(t *testing.T) {
	c := newCombat()
	_ = c.Start()
	_ = c.ApplyWazaAri(1)
	_ = c.ApplyWazaAri(1)
	idx, method, decided := c.Winner()
	if !decided {
		t.Fatal("expected match decided after waza-ari-awasete-ippon")
	}
	if idx != 1 {
		t.Errorf("expected athlete B (idx 1) to win")
	}
	if method != domain.FinishWazaAriAwasete {
		t.Errorf("expected WAZA_ARI_AWASETE_IPPON, got %s", method)
	}
}

func TestCombat_GoldenScoreOnTimerExpiry(t *testing.T) {
	c := newCombat()
	_ = c.Start()
	// Tick past match duration with no score.
	transitioned := c.Tick(domain.MatchDuration + time.Second)
	if !transitioned {
		t.Error("expected transition to golden score")
	}
	if c.Timer.State != domain.StateGoldenScore {
		t.Errorf("want GOLDEN_SCORE, got %s", c.Timer.State)
	}
}

func TestCombat_OsaekomiThresholds(t *testing.T) {
	cases := []struct {
		elapsed    time.Duration
		wantScore  string
	}{
		{4 * time.Second, ""},                       // below yuko
		{5 * time.Second, "YUKO"},                   // exact yuko
		{9 * time.Second, "YUKO"},                   // just below waza-ari
		{10 * time.Second, "WAZA_ARI"},              // exact waza-ari
		{19 * time.Second, "WAZA_ARI"},              // just below ippon
		{20 * time.Second, "IPPON"},                 // exact ippon
		{30 * time.Second, "IPPON"},                 // above ippon
	}
	for _, tc := range cases {
		c := newCombat()
		_ = c.Start()
		now := time.Now()
		_ = c.StartOsaekomi(now)
		fakeNow := now.Add(tc.elapsed)
		elapsed, _ := c.StopOsaekomi(fakeNow)
		got, _ := c.ApplyOsaekomiScore(elapsed, 0)
		if got != tc.wantScore {
			t.Errorf("elapsed %v: want %q, got %q", tc.elapsed, tc.wantScore, got)
		}
	}
}

func TestCombat_FinishAlreadyFinished(t *testing.T) {
	c := newCombat()
	_ = c.Start()
	_ = c.ApplyIppon(0)
	_ = c.Finish()
	if err := c.Finish(); err == nil {
		t.Fatal("expected error finishing already-finished combat")
	}
}
