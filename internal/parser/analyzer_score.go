package parser

import (
	"cs2-parser/internal/models"

	dem "github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs"
	"github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs/events"
)

type scoreAnalyzer struct {
	m *match
}

func newScoreAnalyzer(m *match) *scoreAnalyzer {
	return &scoreAnalyzer{
		m: m,
	}
}

// Compile-time interface checks
var _ Analyzer = (*scoreAnalyzer)(nil)
var _ MatchStartHook = (*scoreAnalyzer)(nil)

// Extracts K/D/A, damage and team name from server memory (1 tick delay)
func (a *scoreAnalyzer) syncScoreboard() {
	if !a.m.isActive() {
		return
	}

	for _, p := range a.m.parser.GameState().Participants().All() {
		if p == nil || p.SteamID64 == 0 || p.Entity == nil || sideName(p.Team) == "" {
			continue
		}

		player := a.m.player(p.SteamID64, p.Name)
		if player == nil {
			continue
		}

		player.Team = sideName(p.Team)
		player.Kills = p.Kills()
		player.Deaths = p.Deaths()
		player.Assists = p.Assists()
		player.TotalDamage = p.TotalDamage()
	}
}

func (a *scoreAnalyzer) onKill(e events.Kill) {
	if !a.m.isActive() {
		return
	}

	a.syncScoreboard()

	if e.Killer != nil && e.Victim != nil && e.Killer.SteamID64 != 0 {
		if e.Killer.Team != e.Victim.Team && e.IsHeadshot {
			player := a.m.player(e.Killer.SteamID64, e.Killer.Name)
			player.Headshots++
		}
	}
}

// Fixes 1 tick delay and catches any last kills occurring in the round (not exit-frags)
func (a *scoreAnalyzer) onRoundEnd(_ events.RoundEnd) {
	a.syncScoreboard()
}

func (a *scoreAnalyzer) Register(p dem.Parser) {
	p.RegisterEventHandler(a.onKill)
	p.RegisterEventHandler(a.onRoundEnd)
}

// Resets eventual headshots occurring in warmup
func (a *scoreAnalyzer) OnMatchStart() {
	for _, p := range a.m.data.Players {
		p.Headshots = 0
	}
}

func (a *scoreAnalyzer) Finalize(ms *models.MatchStats) {
	// In case demo end abruptly after last round & calculates HS%
	a.syncScoreboard()

	for _, p := range ms.Players {
		p.HeadshotPercentage = percentage(p.Kills, p.Headshots)
	}
}
