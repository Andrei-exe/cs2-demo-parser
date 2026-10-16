package parser

import (
	events "github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs/events"
)

// Extracts K/D/A from server memory (1 tick delay)
func (m *MatchAnalyzer) SyncScore() {
	if m.parser.GameState().IsWarmupPeriod() || !m.isLive {
		return
	}

	for _, p := range m.parser.GameState().Participants().All() {
		if p == nil || p.SteamID64 == 0 || p.Entity == nil {
			continue
		}

		m.EnsurePlayerIsRegistered(p.SteamID64, p.Name)

		m.MatchData.Players[p.SteamID64].Kills = p.Kills()
		m.MatchData.Players[p.SteamID64].Deaths = p.Deaths()
		m.MatchData.Players[p.SteamID64].Assists = p.Assists()
		m.MatchData.Players[p.SteamID64].TotalDamage = p.TotalDamage()
	}
}

// Calls for scoreboard update and listens for exit-frags after RoundEnd is triggered
func (m *MatchAnalyzer) OnKill(e events.Kill) {
	m.SyncScore()
	m.SyncUtilityScore()

	if e.Killer != nil && e.Victim != nil && e.Killer.SteamID64 != 0 {
		if e.Killer.Team != e.Victim.Team && e.IsHeadshot {
			m.EnsurePlayerIsRegistered(e.Killer.SteamID64, e.Killer.Name)
			m.MatchData.Players[e.Killer.SteamID64].Headshots++
		}
	}
}
