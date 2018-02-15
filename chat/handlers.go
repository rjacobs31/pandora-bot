package chat

import (
	"github.com/bwmarrin/discordgo"
)

// MessageHandler handles incoming messages using chain of responsibility.
type MessageHandler interface {
	Handle(s *discordgo.Session, m *discordgo.MessageCreate)
	SetNext(n *MessageHandler)
}

type ChatClient struct {
	session *discordgo.Session

	messageHandlers []MessageHandler
}

func New(token string) (client *ChatClient, err error) {
	dg, err := discordgo.New("Bot " + token)
	client = &ChatClient{dg}
	return
}

func (c *ChatClient) Start() {
	c.session.AddHandler(handleIncomingMessage)
}

func (c *ChatClient) Close() {
	c.session.Close()
}

// AddHandler appends a handler to the list of handlers in the
// chain of responsibility.
func (c *ChatClient) AddHandler(newHandler MessageHandler) {
	oldLen := len(c.messageHandlers)
	if oldLen > 0 {
		c.messageHandlers[oldLen-1].SetNext(newHandler)
	}

	append(c.messageHandlers, newHandler)
}

// messageCreate will be called (by AddHandler) every time a new
// message is created on any channel that the autenticated bot has access to.
func (c *ChatClient) handleIncomingMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}

	// If the message is "ping" reply with "Pong!"
	if m.Content == "ping" {
		s.ChannelMessageSend(m.ChannelID, "Pong!")
	}
}
