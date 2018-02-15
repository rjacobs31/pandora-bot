package handlers

import (
	"github.com/bwmarrin/discordgo"
)

// PingHandler verifies that the bot is active, by responding
// with "Pong!" when a message matching "!ping" is encountered.
type PingHandler struct {
	next MessageHandler
}

func (h *PingHandler) SetNext(newHandler MessageHandler) {
	h.next = newHandler
}

func (h *PingHandler) Handle(s *discordgo.Session, m *discordgo.MessageCreate) {
	// If the message is "ping" reply with "Pong!"
	if m.Content == "!ping" {
		s.ChannelMessageSend(m.ChannelID, "Pong!")
		return
	}

	if h.next != nil {
		h.next.Handle(s, m)
	}
}
