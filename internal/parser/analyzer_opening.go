package parser

import (
	"cs2-parser/internal/models"

	dem "github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs"
	"github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs/events"
)

type openingAnalyzer struct {
	m      *match
	done   bool
	killer uint64
	victim uint64
}

func newOpeningAnalyzer(m *match) *openingAnalyzer {
	return &openingAnalyzer{
		m: m,
	}
}

// Compile-time interface checks
var _ Analyzer = (*openingAnalyzer)(nil)
var _ MatchStartHook = (*openingAnalyzer)(nil)
var _ RoundStartHook = (*openingAnalyzer)(nil)
var _ RoundEndHook = (*openingAnalyzer)(nil)

// Resets the opening state for the new round
func (a *openingAnalyzer) OnRoundStart(_ *models.Round) {
	a.done = false
	a.killer = 0
	a.victim = 0
}

func (a *openingAnalyzer) onKill(e events.Kill) {
	if !a.m.isInRound() || a.done {
		return
	}

	if e.Killer == nil || e.Victim == nil || e.Killer.Team == e.Victim.Team {
		return
	}

	a.done = true
	a.killer = e.Killer.SteamID64
	a.victim = e.Victim.SteamID64
	a.m.player(e.Killer.SteamID64, e.Killer.Name)
	a.m.player(e.Victim.SteamID64, e.Victim.Name)
}

// Commits the opening duel only for valid rounds
func (a *openingAnalyzer) OnRoundEnd(r *models.Round) {
	r.OpeningKiller = a.killer
	r.OpeningVictim = a.victim

	if p, ok := a.m.data.Players[a.killer]; ok {
		p.OpeningKills++
	}

	if p, ok := a.m.data.Players[a.victim]; ok {
		p.OpeningDeaths++
	}
}

// Resets eventual opening stats from warmup / knife round
func (a *openingAnalyzer) OnMatchStart() {
	for _, p := range a.m.data.Players {
		p.OpeningKills = 0
		p.OpeningDeaths = 0
	}
}

func (a *openingAnalyzer) Register(p dem.Parser) {
	p.RegisterEventHandler(a.onKill)
}

func (a *openingAnalyzer) Finalize(ms *models.MatchStats) {
	for _, p := range ms.Players {
		p.OpeningSuccess = percentage(p.OpeningDeaths+p.OpeningKills, p.OpeningKills)
	}
}
