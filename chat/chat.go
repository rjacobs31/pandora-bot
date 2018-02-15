package chat

import (
	"github.com/bwmarrin/discordgo"

	pbot "github.com/rjacobs31/pandora-bot/chat/handlers"
)

type ChatClient struct {
	session *discordgo.Session

	messageHandlers []pbot.MessageHandler
}

func New(token string) (client *ChatClient, err error) {
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		return
	}

	client = &ChatClient{session: dg}
	return
}

func (c *ChatClient) Start() {
	c.session.AddHandler(c.handleIncomingMessage)
}

func (c *ChatClient) Close() {
	c.session.Close()
}

// AddHandler appends a handler to the list of handlers in the
// chain of responsibility.
func (c *ChatClient) AddHandler(newHandler pbot.MessageHandler) {
	oldLen := len(c.messageHandlers)
	if oldLen > 0 {
		c.messageHandlers[oldLen-1].SetNext(newHandler)
	}

	c.messageHandlers = append(c.messageHandlers, newHandler)
}

// messageCreate will be called (by AddHandler) every time a new
// message is created on any channel that the autenticated bot has access to.
func (c *ChatClient) handleIncomingMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}

	if len(c.messageHandlers) > 0 {
		c.messageHandlers[0].Handle(s, m)
	}
}
