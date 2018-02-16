package handlers

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"

	"github.com/rjacobs31/pandora-bot/database"
)

// FactoidRegisterHandler checks messages for a new remark-retort pair
// to teach to the bot.
type FactoidRegisterHandler struct {
	fm   *database.FactoidManager
	next MessageHandler
}

func NewFactoidRegisterHandler(fm *database.FactoidManager) (handler *FactoidRegisterHandler) {
	return &FactoidRegisterHandler{fm: fm}
}

// SetNext sets the next handler in the chain, after this
// FactoidRegisterHandler.
func (h *FactoidRegisterHandler) SetNext(newHandler MessageHandler) {
	h.next = newHandler
}

// Handle attempts to extract a remark and a retort to teach the bot.
//
// For the user to teach the bot a new remark-retort pair, the user
// must be addressing the bot. This is true if the message begins
// with "pan:" or "pandora:". This prefix and following whitespace
// is stripped from the message.
//
// If the user is determined to be addressing the bot, the message
// is checked for an "is". Any text before that point is taken
// as a remark, and the whole string is taken as the retort.
func (h *FactoidRegisterHandler) Handle(s *discordgo.Session, m *discordgo.MessageCreate) {
	remark, retort := extractRetortFromMessage(m.Content)

	if (remark == "" || retort == "") && h.next != nil {
		h.next.Handle(s, m)
		return
	}

	h.fm.Add(remark, retort)
	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Okay, remembering that %q is %q.", remark, retort))
}

var (
	// botAddressRegex represents the regex for determining
	// when the user is trying to teach the bot a retort.
	botAddressRegex, _ = regexp.Compile("^pan(dora)?:\\s*")

	// remarkRegexIs represents whether message content
	// contains an "is" for the purposes of delimiting
	// a remark and a retort.
	remarkRegexIs, _ = regexp.Compile("\\s+is")
)

// extractRetortFromMessage checks whether the user is addressing
// the bot and whether a remark and a retort can be extracted.
func extractRetortFromMessage(content string) (remark, retort string) {
	if content == "" {
		return
	}

	contentBytes := []byte(content)
	botAddressIndices := botAddressRegex.FindIndex(contentBytes)
	if botAddressIndices == nil {
		return
	}

	strippedContent := contentBytes[botAddressIndices[1]:]
	indicesIs := remarkRegexIs.FindIndex(strippedContent)
	if indicesIs == nil {
		return
	}

	remark = string(strippedContent[:indicesIs[0]])
	retort = strings.TrimSpace(string(strippedContent))
	return
}
