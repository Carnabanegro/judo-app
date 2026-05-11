package ports

// DisplayEvent is the payload broadcast to all display clients (operator + projector).
type DisplayEvent struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// EventBroadcaster sends real-time events to all connected display clients.
type EventBroadcaster interface {
	Broadcast(event DisplayEvent)
}

// Known event types.
const (
	EventTimerTick    = "combat:tick"
	EventCombatUpdate = "combat:update"
	EventCombatFinish = "combat:finished"
)

// TimerTickPayload is the payload for combat:tick events.
type TimerTickPayload struct {
	MatchID         string  `json:"matchId"`
	RemainingMs     int64   `json:"remainingMs"`
	OsaekomiMs      int64   `json:"osaekoiMs"`
	State           string  `json:"state"`
	GoldenScore     bool    `json:"goldenScore"`
}

// CombatUpdatePayload is the payload for combat:update events.
type CombatUpdatePayload struct {
	MatchID  string      `json:"matchId"`
	ScoreA   ScoreDTO    `json:"scoreA"`
	ScoreB   ScoreDTO    `json:"scoreB"`
	State    string      `json:"state"`
	LabelA   string      `json:"labelA"`
	LabelB   string      `json:"labelB"`
}

// ScoreDTO is a serialisable score snapshot.
type ScoreDTO struct {
	Ippon   int  `json:"ippon"`
	WazaAri int  `json:"wazaAri"`
	Yuko    int  `json:"yuko"`
	Shido   int  `json:"shido"`
	Hansoku bool `json:"hansoku"`
}

// CombatFinishPayload is the payload for combat:finished events.
type CombatFinishPayload struct {
	MatchID    string `json:"matchId"`
	WinnerIdx  int    `json:"winnerIdx"`
	Method     string `json:"method"`
}
