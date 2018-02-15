package handlers

import (
	"github.com/bwmarrin/discordgo"
)

// MessageHandler handles incoming messages using chain of responsibility.
type MessageHandler interface {
	Handle(s *discordgo.Session, m *discordgo.MessageCreate)
	SetNext(n *MessageHandler)
}
