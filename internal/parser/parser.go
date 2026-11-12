package parser

import (
	"context"
	"cs2-parser/internal/models"
	"errors"
	"fmt"
	"math"

	dem "github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs"
	com "github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs/common"
	"github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs/events"
)

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

// Parses a demo file and returns the stats computed by all analyzers
func ParseDemo(ctx context.Context, demoFilePath string, onProgress func(float64)) (*models.MatchStats, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	m := newMatch()

	m.analyzers = []Analyzer{
		// Combat-related analyzers
		newScoreAnalyzer(m),
		newOpeningAnalyzer(m),

		// Utility-related analyzers
		newUtilityAnalyzer(m),

		// Other-related analyzers
		newChatAnalyzer(m),
	}

	err := dem.ParseFile(demoFilePath, func(p dem.Parser) error {
		m.parser = p
		registerControl(ctx, p, onProgress)

		// Core must be registered first so it decides round validity before analyzers react
		m.register(p)

		for _, a := range m.analyzers {
			a.Register(p)
		}

		return nil
	})

	if err != nil {
		if errors.Is(err, dem.ErrCancelled) && ctx.Err() != nil {
			return nil, fmt.Errorf("parsing aborted: %w", ctx.Err())
		}
		return nil, err
	}

	// Final syncs and percentages
	for _, a := range m.analyzers {
		a.Finalize(m.data)
	}

	// Set team names and score
	gameState := m.parser.GameState()
	m.data.CT = models.TeamStats{
		Name:  gameState.TeamCounterTerrorists().ClanName(),
		Score: gameState.TeamCounterTerrorists().Score(),
	}
	m.data.T = models.TeamStats{
		Name:  gameState.TeamTerrorists().ClanName(),
		Score: gameState.TeamTerrorists().Score(),
	}

	return m.data, nil
}

func registerControl(ctx context.Context, p dem.Parser, onProgress func(float64)) {
	var last float32

	p.RegisterEventHandler(func(events.FrameDone) {
		if ctx.Err() != nil {
			p.Cancel()
			return
		}

		if onProgress == nil {
			return
		}

		current := p.Progress()
		if current-last >= 0.01 {
			last = current
			onProgress(float64(current))
		}
	})
}
