package handlers

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// PingHandler verifies that the bot is active, by responding
// with "Pong!" when a message matching "!ping" is encountered.
type BananaLoveHandler struct {
	next MessageHandler
}

func (h *BananaLoveHandler) SetNext(newHandler MessageHandler) {
	h.next = newHandler
}

func (h *BananaLoveHandler) Handle(s *discordgo.Session, m *discordgo.MessageCreate) {
	// If the message is "ping" reply with "Pong!"
	fmt.Println("Checking banana")
	if strings.Contains(strings.ToLower(m.Content), "banana") {
		fmt.Println("Found banana")
		fmt.Println(s.MessageReactionAdd(m.ChannelID, m.ID, "\U0001F60D"))
	}

	if h.next != nil {
		h.next.Handle(s, m)
	}
}
