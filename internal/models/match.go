package models

type MatchStats struct {
	MapName string                  `json:"map_name"`
	CT      TeamStats               `json:"ct"`
	T       TeamStats               `json:"t"`
	Players map[uint64]*PlayerStats `json:"players"`
	Rounds  []*Round                `json:"rounds"`
}

type TeamStats struct {
	Name  string `json:"name"`
	Score int    `json:"score"`
}
