package parser

import (
	events "github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs/events"
)

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
	}
}

func (m *MatchAnalyzer) OnKill(e events.Kill) {
	m.SyncScore()

	if e.Victim != nil && e.Victim.SteamID64 != 0 {
		m.MatchData.Players[e.Victim.SteamID64].Deaths++
	}

	if e.Killer != nil && e.Killer.SteamID64 != 0 && e.Killer != e.Victim {
		m.MatchData.Players[e.Killer.SteamID64].Kills++
	}

	if e.Assister != nil && e.Assister.SteamID64 != 0 && e.AssistedFlash == false {
		m.MatchData.Players[e.Assister.SteamID64].Assists++
	}
}
