package parser

import (
	"cs2-parser/internal/models"

	dem "github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs"
)

// Computes one group of stats from a demo; every analyzer must be registered in ParseDemo to be used
type Analyzer interface {
	// The analyzer subscribes here to the events it needs and does nothing else
	Register(p dem.Parser)

	// Used for end-of-match calculations (percentages, final syncs)
	Finalize(m *models.MatchStats)
}

// Implemented by analyzers that need to reset their own player stats when the match (re)starts (can be called multiple times per demo)
type MatchStartHook interface {
	OnMatchStart()
}

// Implemented by analyzers that keep per-round state
type RoundStartHook interface {
	OnRoundStart(r *models.Round)
}

// Implemented by analyzers that commit per-round data
type RoundEndHook interface {
	OnRoundEnd(r *models.Round)
}
