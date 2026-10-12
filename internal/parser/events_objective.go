package parser

import (
	events "github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs/events"
)

func (m *MatchAnalyzer) InitializePlayersOnMatchStart(e events.AnnouncementMatchStarted) {
	m.isLive = true

	for _, p := range m.parser.GameState().Participants().Playing() {

		if p.SteamID64 == 0 {
			continue
		}

		m.EnsurePlayerIsRegistered(p.SteamID64, p.Name)
	}
}
