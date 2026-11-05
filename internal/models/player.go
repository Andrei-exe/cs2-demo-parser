package models

type PlayerStats struct {
	SteamID uint64 `json:"steam_id"`
	Name    string `json:"name"`
	Team    string `json:"team"`

	Kills              int `json:"kills"`
	Deaths             int `json:"deaths"`
	Assists            int `json:"assists"`
	TotalDamage        int `json:"total_damage"`
	Headshots          int `json:"headshots"`
	HeadshotPercentage int `json:"headshot_percentage"`

	UtilityDamage  int `json:"utility_damage"`
	EnemiesFlashed int `json:"enemies_flashed"`

	OpeningKills   int `json:"opening_kills"`
	OpeningDeaths  int `json:"opening_deaths"`
	OpeningSuccess int `json:"opening_success"`

	ChatMessages []string `json:"chat_messages"`
}

func (p *PlayerStats) ResetEventsStats() {
	p.Headshots = 0
	p.OpeningKills = 0
	p.OpeningDeaths = 0
}
