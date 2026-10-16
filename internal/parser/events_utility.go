package parser

func (m *MatchAnalyzer) SyncUtilityScore() {
	if m.parser.GameState().IsWarmupPeriod() || !m.isLive {
		return
	}

	for _, p := range m.parser.GameState().Participants().All() {
		if p == nil || p.SteamID64 == 0 || p.Entity == nil {
			continue
		}

		m.EnsurePlayerIsRegistered(p.SteamID64, p.Name)

		m.MatchData.Players[p.SteamID64].UtilityDamage = p.UtilityDamage()

		flash_prop := p.Entity.Property("m_pActionTrackingServices.m_iEnemiesFlashed")
		if flash_prop != nil {
			m.MatchData.Players[p.SteamID64].EnemiesFlashed = int(flash_prop.Value().Int())
		}
	}
}
