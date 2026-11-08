package parser

import (
	"cs2-parser/internal/models"
	"fmt"

	dem "github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs"
	"github.com/markus-wa/demoinfocs-golang/v5/pkg/demoinfocs/events"
)

type chatAnalyzer struct {
	m *match
}

func newChatAnalyzer(m *match) *chatAnalyzer {
	return &chatAnalyzer{
		m: m,
	}
}

// Compile-time interface check
var _ Analyzer = (*chatAnalyzer)(nil)

// Extracts all messages (sent by players) from the chat
func (a *chatAnalyzer) onChat(e events.ChatMessage) {
	if e.Sender == nil {
		return
	}
	typeChat := "[Team]"
	if e.IsChatAll {
		typeChat = "[All]"
	}

	chatMessage := fmt.Sprintf("%s %s", typeChat, e.Text)

	player := a.m.player(e.Sender.SteamID64, e.Sender.Name)
	if player == nil {
		return
	}
	player.ChatMessages = append(player.ChatMessages, chatMessage)
}

func (a *chatAnalyzer) Register(p dem.Parser) {
	p.RegisterEventHandler(a.onChat)
}

// Nothing to process at the end
func (a *chatAnalyzer) Finalize(_ *models.MatchStats) {}
