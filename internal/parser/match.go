package parser

import (
	"cs2-parser/internal/models"

	dem "github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs"
	events "github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs/events"
	msg "github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs/msg"
)

// Holds the shared match state and drives the round lifecycle for all analyzers
type match struct {
	data         *models.MatchStats
	parser       dem.Parser
	isLive       bool
	currentRound *models.Round
	roundEnded   bool
	analyzers    []Analyzer
}

// Creates an empty match with the players map initialized
func newMatch() *match {
	return &match{
		data: &models.MatchStats{
			Players: make(map[uint64]*models.PlayerStats),
			Rounds:  []*models.Round{},
		},
	}
}

// Register player if missing (or return him otherwise)
func (m *match) player(id uint64, name string) *models.PlayerStats {
	if id == 0 {
		return nil
	}

	if p, ok := m.data.Players[id]; ok {
		return p
	}

	m.data.Players[id] = &models.PlayerStats{
		SteamID: id,
		Name:    name,
	}
	return m.data.Players[id]
}

// Check match is live
func (m *match) isActive() bool {
	return m.isLive && !m.parser.GameState().IsWarmupPeriod()
}

// Check event occurred during a live round
func (m *match) isInRound() bool {
	return m.isActive() && m.currentRound != nil && !m.roundEnded
}

// Opens a new round and lets every analyzer reset its per-round state
func (m *match) startRound() {
	m.currentRound = &models.Round{}
	m.roundEnded = false

	for _, a := range m.analyzers {
		if h, ok := a.(RoundStartHook); ok {
			h.OnRoundStart(m.currentRound)
		}
	}
}

// Marks the match as live, clears stats from before the start and opens the first round
func (m *match) onMatchStart(_ events.AnnouncementMatchStarted) {
	m.isLive = true

	// Clear unnecessary stats from knife rounds / warmup / pre-restart events
	m.data.Rounds = []*models.Round{}

	for _, p := range m.parser.GameState().Participants().Playing() {
		m.player(p.SteamID64, p.Name)
	}

	for _, a := range m.analyzers {
		if h, ok := a.(MatchStartHook); ok {
			h.OnMatchStart()
		}
	}

	m.startRound()
}

func (m *match) onRoundStart(_ events.RoundStart) {
	m.startRound()
}

// Commits the round only if it's valid
func (m *match) onRoundEnd(e events.RoundEnd) {
	defer func() { m.roundEnded = true }()

	if !m.isInRound() {
		return
	}

	winner := sideName(e.Winner)
	if winner == "" {
		return
	}

	m.currentRound.Number = len(m.data.Rounds) + 1
	m.currentRound.Winner = winner

	for _, a := range m.analyzers {
		if h, ok := a.(RoundEndHook); ok {
			h.OnRoundEnd(m.currentRound)
		}
	}
	m.data.Rounds = append(m.data.Rounds, m.currentRound)
}

// Subscribes the core handlers; must be called before any analyzer's Register
func (m *match) register(p dem.Parser) {
	p.RegisterEventHandler(m.onMatchStart)
	p.RegisterEventHandler(m.onRoundStart)
	p.RegisterEventHandler(m.onRoundEnd)
	p.RegisterNetMessageHandler(func(info *msg.CSVCMsg_ServerInfo) { m.data.MapName = info.GetMapName() })
}
