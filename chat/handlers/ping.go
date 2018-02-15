package handlers

import (
	"github.com/bwmarrin/discordgo"
)

type PingHandler struct {
	next MessageHandler
}

func (h *PingHandler) Handle(s *discordgo.Session, m *discordgo.MessageCreate) {
	// If the message is "ping" reply with "Pong!"
	if m.Content == "ping" {
		s.ChannelMessageSend(m.ChannelID, "Pong!")
		return
	}

	if h.next != nil {
		h.next.Handle(s, m)
	}
}
