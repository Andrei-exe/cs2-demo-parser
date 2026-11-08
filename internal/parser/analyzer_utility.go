package parser

import (
	"cs2-parser/internal/models"

	dem "github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs"
	"github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs/events"
)

type utilityAnalyzer struct {
	m *match
}

func newUtilityAnalyzer(m *match) *utilityAnalyzer {
	return &utilityAnalyzer{
		m: m,
	}
}

// Compile-time interface check
var _ Analyzer = (*utilityAnalyzer)(nil)

// Extracts utility stats from the scoreboard
func (a *utilityAnalyzer) syncUtilityScoreboard() {
	if !a.m.isActive() {
		return
	}

	for _, p := range a.m.parser.GameState().Participants().All() {
		if p == nil || p.SteamID64 == 0 || p.Entity == nil || sideName(p.Team) == "" {
			continue
		}

		player := a.m.player(p.SteamID64, p.Name)
		if player == nil {
			continue
		}

		player.UtilityDamage = p.UtilityDamage()

		flashProp := p.Entity.Property("m_pActionTrackingServices.m_iEnemiesFlashed")
		if flashProp != nil {
			player.EnemiesFlashed = int(flashProp.Value().Int())
		}
	}
}

func (a *utilityAnalyzer) onKill(_ events.Kill) {
	a.syncUtilityScoreboard()
}

func (a *utilityAnalyzer) onRoundEnd(_ events.RoundEnd) {
	a.syncUtilityScoreboard()
}

func (a *utilityAnalyzer) Register(p dem.Parser) {
	p.RegisterEventHandler(a.onKill)
	p.RegisterEventHandler(a.onRoundEnd)
}

// In case demo end abruptly after last round
func (a *utilityAnalyzer) Finalize(_ *models.MatchStats) {
	a.syncUtilityScoreboard()
}
