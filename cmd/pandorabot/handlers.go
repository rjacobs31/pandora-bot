package main

import (
	"fmt"
	"log"
	"math/rand"
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"

	"../../bolt"
	"../../interpolate"
)

// CommandType Type of command given to Pandora
type CommandType int

// Enum of command types
const (
	InsertCommand CommandType = iota
)

// InsertCommandType Type of insert command given to Pandora
type InsertCommandType int

// Enum of insert command types
const (
	ExplicitIs InsertCommandType = iota
	ImplicitIs
	IsReply
)

var (
	// Regexes
	addressRegex    regexp.Regexp = *regexp.MustCompile("^pan(dora)?: *")
	isReplyRegex    regexp.Regexp = *regexp.MustCompile("( +is)? *<reply> *")
	explicitIsRegex regexp.Regexp = *regexp.MustCompile(" *<is> *")
	impliedIsRegex  regexp.Regexp = *regexp.MustCompile(" +is +")
)

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

	// Pandora likes any post about bananas
	if strings.Contains(strings.ToLower(m.Content), "banana") {
		s.MessageReactionAdd(m.ChannelID, m.ID, "😍")
	}

	if addressRegex.MatchString(m.Content) {
		str := addressRegex.ReplaceAllLiteralString(m.Content, "")

		if idxs := isReplyRegex.FindIndex([]byte(str)); idxs != nil {
			insertHandler(s, IsReply, m, idxs, str)
		} else if idxs := explicitIsRegex.FindIndex([]byte(str)); idxs != nil {
			insertHandler(s, ExplicitIs, m, idxs, str)
		} else if idxs := impliedIsRegex.FindIndex([]byte(str)); idxs != nil {
			insertHandler(s, ImplicitIs, m, idxs, str)
		}
	} else {
		responseHandler(s, m)
	}
}

func insertHandler(s *discordgo.Session, commandType InsertCommandType, m *discordgo.MessageCreate, idxs []int, str string) {
	var (
		err      error
		response string
		trigger  string
	)

	switch commandType {
	case IsReply:
		trigger = str[:idxs[0]]
		response = strings.TrimSpace(str[idxs[1]:])
		err = Client.FactoidService().PutResponse(trigger, response)
	case ExplicitIs:
		trigger = str[:idxs[0]]
		response = strings.TrimSpace(str[:idxs[0]] + " is " + str[idxs[1]:])
		err = Client.FactoidService().PutResponse(trigger, response)
	case ImplicitIs:
		trigger = str[:idxs[0]]
		response = strings.TrimSpace(str[:])
		err = Client.FactoidService().PutResponse(trigger, response)
	default:
		return
	}

	switch err.(type) {
	case nil:
		s.ChannelMessageSend(m.ChannelID, "Okay, "+m.Author.Mention()+". Remembering that \""+trigger+"\" is \""+response+"\" 🙂")
	case bolt.FactoidAlreadyExistsError:
		s.ChannelMessageSend(m.ChannelID, "But \""+trigger+"\" is already \""+response+"\" 🤔")
	default:
		s.ChannelMessageSend(m.ChannelID, "Sorry. Error putting into DB. 😞\n\nTrigger: \""+trigger+"\"\n\nResponse: \""+response+"\"")
	}
}

func responseHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	trigger := m.Content[:]
	response, err := Client.FactoidService().GetRandomResponse(trigger)
	var someone string
	interpolations := map[string]interface{}{
		"who": func() string {
			return m.Author.Mention()
		},
		"someone": func() string {
			if someone != "" {
				return someone
			}
			if channel, errC := s.Channel(m.ChannelID); errC == nil {
				if members, errG := s.GuildMembers(channel.GuildID, "", 50); errG == nil {
					idx := rand.Intn(len(members))
					someone = members[idx].User.Mention()
				} else {
					log.Println("Error fetching guild members: ", errG)
				}
			} else {
				log.Println("Error fetching channel: ", errC)
			}

			if someone == "" {
				someone = "Someone"
			}
			return someone
		},
	}

	if err == nil && response != "" {
		if response != "" {
			interp := &interpolate.Interpolator{}
			interp.SetMap(interpolations)
			result, _ := interp.Interpolate(response)
			if err != nil {
				fmt.Println(err)
				return
			}
			s.ChannelMessageSend(m.ChannelID, result)
		}
	} else if err != nil {
		fmt.Println(err)
	}
}
