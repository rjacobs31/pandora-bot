package handlers

import (
	"github.com/bwmarrin/discordgo"
)

// MessageHandler handles incoming messages using chain of responsibility.
//
// Every handler may decide to handle the request itself and terminate
// the chain, or may pass that responsibility to the next handler
// in the chain.
type MessageHandler interface {
	// Handle takes an incoming message and either handles it or
	// passes responsibility to the next handler in the chain.
	Handle(s *discordgo.Session, m *discordgo.MessageCreate)

	// SetNext specifies  a different handler which the current handler
	// may call if it decides to delegate responsibility.
	SetNext(n MessageHandler)
}
