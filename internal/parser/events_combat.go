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

		playerSide := sideName(p.Team)
		if playerSide == "" {
			continue
		}

		m.EnsurePlayerIsRegistered(p.SteamID64, p.Name)

		m.MatchData.Players[p.SteamID64].Team = playerSide
		m.MatchData.Players[p.SteamID64].Kills = p.Kills()
		m.MatchData.Players[p.SteamID64].Deaths = p.Deaths()
		m.MatchData.Players[p.SteamID64].Assists = p.Assists()
		m.MatchData.Players[p.SteamID64].TotalDamage = p.TotalDamage()
	}
}

// Calls for scoreboard update and listens for exit-frags after RoundEnd is triggered
func (m *MatchAnalyzer) OnKill(e events.Kill) {
	if m.parser.GameState().IsWarmupPeriod() || !m.isLive {
		return
	}

	m.SyncScore()
	m.SyncUtilityScore()

	if e.Killer != nil && e.Victim != nil && e.Killer.SteamID64 != 0 {
		if e.Killer.Team != e.Victim.Team && e.IsHeadshot {
			m.EnsurePlayerIsRegistered(e.Killer.SteamID64, e.Killer.Name)
			m.MatchData.Players[e.Killer.SteamID64].Headshots++
		}
	}
}

func (m *MatchAnalyzer) TrackOpeningDuel(e events.Kill) {
	if m.parser.GameState().IsWarmupPeriod() || !m.isLive || m.round.openingDone || m.round.ended {
		return
	}

	if e.Killer != nil && e.Victim != nil {
		if e.Killer.Team != e.Victim.Team {
			m.round.openingDone = true
			if e.Killer.SteamID64 != 0 {
				m.EnsurePlayerIsRegistered(e.Killer.SteamID64, e.Killer.Name)
				m.round.openingKiller = e.Killer.SteamID64
			}
			if e.Victim.SteamID64 != 0 {
				m.EnsurePlayerIsRegistered(e.Victim.SteamID64, e.Victim.Name)
				m.round.openingVictim = e.Victim.SteamID64
			}
		}
	}
}
