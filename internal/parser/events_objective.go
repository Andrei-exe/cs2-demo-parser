package parser

import (
	"cs2-parser/internal/models"

	events "github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs/events"
)

func (m *MatchAnalyzer) OnMatchStart(e events.AnnouncementMatchStarted) {
	m.isLive = true

	for _, p := range m.parser.GameState().Participants().Playing() {

		if p.SteamID64 == 0 {
			continue
		}

		m.MatchData.Players[p.SteamID64] = &models.PlayerStats{
			SteamID: p.SteamID64,
			Name:    p.Name,
		}
	}
}
