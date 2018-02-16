package handlers

import (
	"github.com/bwmarrin/discordgo"

	"github.com/rjacobs31/pandora-bot/database"
)

// RetortHandler attempts to find a retort for a remark made
// by the user.
type RetortHandler struct {
	fm   *database.FactoidManager
	next MessageHandler
}

func NewRetortHandler(fm *database.FactoidManager) (handler *RetortHandler) {
	return &RetortHandler{fm: fm}
}

// SetNext sets the next handler in the chain, after this RetortHandler.
func (h *RetortHandler) SetNext(newHandler MessageHandler) {
	h.next = newHandler
}

// Handle attempts to lookup a retort for the user's message.
//
// After the message is cleaned, the database is queried to find
// a retort registered for this remark.
//
// If a retort is found, it is sent to the user and the chain of
// responsibility terminates. If no retort is found, the next
// handler in the chain is tried.
func (h *RetortHandler) Handle(s *discordgo.Session, m *discordgo.MessageCreate) {
	retort, err := h.fm.Select(m.Content)
	if err != nil || retort == nil {
		if h.next != nil {
			h.next.Handle(s, m)
		}
		return
	}

	s.ChannelMessageSend(m.ChannelID, retort.Text)
}
