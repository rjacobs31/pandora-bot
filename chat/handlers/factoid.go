package handlers

import (
	"fmt"
	"regexp"

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

	// remarkRegexReply represents a message attempting to teach
	// the bot a complex query.
	remarkRegexReply, _ = regexp.Compile("\\s*<reply>")

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
	if indices := remarkRegexReply.FindIndex(strippedContent); indices != nil {
		return extractReplyRetort(strippedContent, indices)
	} else if indices := remarkRegexIs.FindIndex(strippedContent); indices != nil {
		return extractIsRetort(strippedContent, indices)
	}
	return
}

// extractReplyRetort extracts a remark-retort pair when the content is known
// to be of the complex reply type.
//
// Everything before the matched expression is the remark and everything
// after the matched expression is the retort.
func extractReplyRetort(content []byte, indices []int) (remark, retort string) {
	return string(content[:indices[0]]), string(content[indices[1]:])
}

// extractIsRetort extracts a remark-retort pair when the content is known
// to be of the simple "is" type.
//
// Everything before the matched expression is the remark and the entire
// string (including the beginning) is the retort.
func extractIsRetort(content []byte, indices []int) (remark, retort string) {
	return string(content[:indices[0]]), string(content[:])
}
