package parser

import (
	"cs2-parser/internal/models"
	"math"

	dem "github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs"
	com "github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs/common"
	msg "github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs/msg"
)

type MatchAnalyzer struct {
	MatchData *models.MatchStats
	parser    dem.Parser
	isLive    bool
	round     roundState
}

type roundState struct {
	openingDone   bool
	ended         bool
	openingKiller uint64
	openingVictim uint64
}

func (m *MatchAnalyzer) resetRound() {
	m.round = roundState{
		openingDone: false,
		ended:       false,
	}
}

func (m *MatchAnalyzer) commitRound() {
	if p, ok := m.MatchData.Players[m.round.openingKiller]; ok {
		p.OpeningKills++
	}
	if p, ok := m.MatchData.Players[m.round.openingVictim]; ok {
		p.OpeningDeaths++
	}
}

func sideName(team com.Team) string {
	switch team {
	case com.TeamCounterTerrorists:
		return "CT"
	case com.TeamTerrorists:
		return "T"
	default:
		return ""
	}
}

func percentage(total, part int) int {
	if total == 0 {
		return 0
	}
	perc := (float64(part) / float64(total)) * 100
	return int(math.Round(perc))
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

		// PlayerData-related handlers
		p.RegisterEventHandler(analyzer.InitializePlayersOnMatchStart)

		// Combat-related handlers
		p.RegisterEventHandler(analyzer.OnKill)
		p.RegisterEventHandler(analyzer.TrackOpeningDuel)

		// RoundState-related handlers
		p.RegisterEventHandler(analyzer.OnRoundStart)
		p.RegisterEventHandler(analyzer.OnRoundEnd)

		// Other-related handlers
		p.RegisterEventHandler(analyzer.RegisterChatMessages)

		return nil
	})

	if err != nil {
		return nil, err
	}

	// In case demo end abruptly after last round
	analyzer.SyncScore()
	analyzer.SyncUtilityScore()

	for _, p := range analyzer.MatchData.Players {
		p.HeadshotPercentage = percentage(p.Kills, p.Headshots)
		p.OpeningSuccess = percentage(p.OpeningDeaths+p.OpeningKills, p.OpeningKills)
	}

	gameState := analyzer.parser.GameState()
	analyzer.MatchData.CT = models.TeamStats{
		Name:  gameState.TeamCounterTerrorists().ClanName(),
		Score: gameState.TeamCounterTerrorists().Score(),
	}
	analyzer.MatchData.T = models.TeamStats{
		Name:  gameState.TeamTerrorists().ClanName(),
		Score: gameState.TeamTerrorists().Score(),
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
