package models

type MatchStats struct {
	MapName   string                  `json:"map_name"`
	TeroScore int                     `json:"t_score"`
	CTScore   int                     `json:"ct_score"`
	Players   map[uint64]*PlayerStats `json:"players"`
}
