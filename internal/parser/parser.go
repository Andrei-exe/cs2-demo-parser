package parser

import (
	"cs2-parser/internal/models"

	dem "github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs"
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
		p.RegisterEventHandler(analyzer.OnMatchStart)
		p.RegisterEventHandler(analyzer.OnKill)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return analyzer.MatchData, nil
}
