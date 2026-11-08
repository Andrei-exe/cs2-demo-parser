package models

type Round struct {
	Number        int    `json:"number"`
	Winner        string `json:"winner"`
	OpeningKiller uint64 `json:"opening_killer,omitempty"`
	OpeningVictim uint64 `json:"opening_victim,omitempty"`
}
