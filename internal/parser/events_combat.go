package parser

import (
	"cs2-parser/internal/models"

	events "github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs/events"
)

func (m *MatchAnalyzer) UpdateScoreboardOnRoundEnd(_ events.FrameDone) {
	if m.parser.GameState().IsWarmupPeriod() || !m.isLive {
		return
	}

	for _, p := range m.parser.GameState().Participants().All() {
		if p == nil || p.SteamID64 == 0 || p.Entity == nil {
			continue
		}

		if m.MatchData.Players[p.SteamID64] == nil {
			m.MatchData.Players[p.SteamID64] = &models.PlayerStats{
				SteamID: p.SteamID64,
				Name:    p.Name,
			}
		}

		m.MatchData.Players[p.SteamID64].Kills = p.Kills()
		m.MatchData.Players[p.SteamID64].Deaths = p.Deaths()
		m.MatchData.Players[p.SteamID64].Assists = p.Assists()
	}
}
