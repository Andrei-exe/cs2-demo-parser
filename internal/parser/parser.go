package parser

import (
	"cs2-parser/internal/models"

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

		return nil
	})

	if err != nil {
		return nil, err
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
