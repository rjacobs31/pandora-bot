package handlers

import (
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

	strippedContent := content[botAddressIndices[1]:]
	indicesIs := remarkRegexIs.FindIndex(content)
	if indicesIs == nil {
		return
	}

	remark = string(strippedContent[:indicesIs[0]])
	retort = strings.TrimSpace(string(strippedContent))
	return
}

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
		h.next.Handle(s, m)
		return
	}

	if h.next != nil {
		h.next.Handle(s, m)
	}
}
