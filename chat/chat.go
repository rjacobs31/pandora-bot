package chat

import (
	"github.com/bwmarrin/discordgo"

	pbot "github.com/rjacobs31/pandora-bot/chat/handlers"
)

// ChatClient manages a Discord session and ensures that registered
// message handlers are called in a chain.
//
// When a message is received, it is passed to the first handler
// in the chain. Each handler may handle the message itself, or
// delegate responsibility for the message to the next handler.
type ChatClient struct {
	// session represents the Discord session used by the bot.
	session *discordgo.Session

	// messageHandlers contains the chain of message handlers
	// which will attempt to handle incoming messages.
	messageHandlers []pbot.MessageHandler
}

// New generates a new ChatClient, which will use the given token
// to open a websocket to the Discord servers.
func New(token string) (client *ChatClient, err error) {
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		return
	}

	client = &ChatClient{
		session: dg,
	}

	return
}

// Start initialises the ChatClient, registering itself to listen
// to incoming messages and opening a connection to the Discord
// servers.
func (c *ChatClient) Start() (err error) {
	c.session.AddHandler(c.handleIncomingMessage)
	return c.session.Open()
}

// Close ends the ChatClient's session with Discord.
func (c *ChatClient) Close() {
	c.session.Close()
}

// AddHandler appends a handler to the list of handlers in the
// chain of responsibility.
func (c *ChatClient) AddHandler(newHandler pbot.MessageHandler) {
	// Set previous message handler to use the new handler
	// as its next.
	if oldLen := len(c.messageHandlers); oldLen > 0 {
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
