package models

type PlayerStats struct {
	SteamID uint64 `json:"steam_id"`
	Name    string `json:"name"`

	Kills              int `json:"kills"`
	Deaths             int `json:"deaths"`
	Assists            int `json:"assists"`
	TotalDamage        int `json:"total_damage"`
	HeadshotPercentage int `json:"headshot_percentage"`
	UtilityDamage      int `json:"utility_damage"`
	EnemiesFlashed     int `json:"enemies_flashed"`

	ChatMessages []string `json:"chat_messages"`
}
