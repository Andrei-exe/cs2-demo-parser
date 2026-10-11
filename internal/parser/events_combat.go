package parser

import (
	events "github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs/events"
)

func (m *MatchAnalyzer) OnKill(e events.Kill) {
	if m.isLive == false {
		return
	}

	if e.Victim != nil && e.Victim.SteamID64 != 0 {
		m.MatchData.Players[e.Victim.SteamID64].Deaths++
	}

	if e.Killer != nil && e.Killer.SteamID64 != 0 && e.Killer != e.Victim {
		m.MatchData.Players[e.Killer.SteamID64].Kills++
	}

	if e.Assister != nil && e.Assister.SteamID64 != 0 {
		m.MatchData.Players[e.Assister.SteamID64].Assists++
	}
}
