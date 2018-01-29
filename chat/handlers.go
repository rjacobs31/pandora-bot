package chat

import "github.com/bwmarrin/discordgo"

type ChatClient struct {
	discordgo.Session
}

func New(token string) (client *ChatClient, err error) {
	dg, err := discordgo.New(token)
	return &ChatClient{dg}
}

func (c *ChatClient) Start() {
	c.AddHandler(messageCreate)
}

func (c *ChatClient) Close() {
	c.Session.Close()
}

// messageCreate will be called (by AddHandler) every time a new
// message is created on any channel that the autenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
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
