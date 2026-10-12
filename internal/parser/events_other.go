package parser

import (
	"cs2-parser/internal/models"
	"fmt"

	events "github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs/events"
)

func (m *MatchAnalyzer) RegisterChatMessages(e events.ChatMessage) {

	if e.Sender != nil && e.Sender.SteamID64 != 0 {
		typeChat := "[Team]"
		if e.IsChatAll {
			typeChat = "[All]"
		}
		chatMessage := fmt.Sprintf("%s %s", typeChat, e.Text)
		if m.MatchData.Players[e.Sender.SteamID64] == nil {
			m.MatchData.Players[e.Sender.SteamID64] = &models.PlayerStats{
				SteamID: e.Sender.SteamID64,
				Name:    e.Sender.Name,
			}
		}
		m.MatchData.Players[e.Sender.SteamID64].ChatMessages = append(m.MatchData.Players[e.Sender.SteamID64].ChatMessages, chatMessage)
	}
}
