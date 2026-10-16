package parser

import (
	"cs2-parser/internal/models"
	"math"

	dem "github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs"
	msg "github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs/msg"
)

type MatchAnalyzer struct {
	MatchData *models.MatchStats
	parser    dem.Parser
	isLive    bool
}

func ParseDemo(demoFilePath string) (*models.MatchStats, error) {
	analyzer := MatchAnalyzer{
		MatchData: &models.MatchStats{
			Players: make(map[uint64]*models.PlayerStats),
		},
		isLive: false,
	}

	err := dem.ParseFile(demoFilePath, func(p dem.Parser) error {
		analyzer.parser = p

		p.RegisterNetMessageHandler(func(m *msg.CSVCMsg_ServerInfo) {
			analyzer.MatchData.MapName = m.GetMapName()
		})

		p.RegisterEventHandler(analyzer.InitializePlayersOnMatchStart)
		p.RegisterEventHandler(analyzer.OnKill)
		p.RegisterEventHandler(analyzer.RegisterChatMessages)
		p.RegisterEventHandler(analyzer.OnRoundEnd)

		return nil
	})

	if err != nil {
		return nil, err
	}

	// In case demo end abruptly after last round
	analyzer.SyncScore()

	for _, p := range analyzer.MatchData.Players {
		if p.Kills > 0 && p.Headshots > 0 {
			percentage := (float64(p.Headshots) / float64(p.Kills)) * 100
			p.HeadshotPercentage = int(math.Round(percentage))
		}
	}

	return analyzer.MatchData, nil
}

func (m *MatchAnalyzer) EnsurePlayerIsRegistered(steamID uint64, name string) {
	if steamID == 0 {
		return
	}

	if m.MatchData.Players[steamID] == nil {
		m.MatchData.Players[steamID] = &models.PlayerStats{
			SteamID: steamID,
			Name:    name,
		}
	}
}
