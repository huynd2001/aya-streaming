package discord_source

import (
	. "aya-backend/service"
	"fmt"
	dg "github.com/bwmarrin/discordgo"
	"regexp"
	"slices"
	"strings"
)

const (
	EMOJI_REGEX   = `<a?:[A-z0-9_~]+:[0-9]+>`
	CHANNEL_REGEX = `<#[0-9]+>`
	MENTION_REGEX = `<@!?[0-9]+>`
	ROLE_REGEX    = `<@&[0-9]+>`
	EVERYONE      = "@everyone"
	HERE          = "@here"

	EMOJI_SPLIT_REGEX   = `<(a?)(:[A-z0-9_~]+:)([0-9]+)>`
	CHANNEL_SPLIT_REGEX = `<#([0-9]+)>`
	MENTION_SPLIT_REGEX = `<@!?([0-9]+)>`
	ROLE_SPLIT_REGEX    = `<@&([0-9]+)>`
)

type RegexIndex int

const (
	EmojiRegex RegexIndex = iota
	ChannelRegex
	MentionRegex
	RoleRegex
	Everyone
	Here
)

var (
	// Do not change this line since it corresponds to the upper defined "enums".
	regexes       = []string{EMOJI_REGEX, CHANNEL_REGEX, MENTION_REGEX, ROLE_REGEX, EVERYONE, HERE}
	splitRegexes  = []string{EMOJI_SPLIT_REGEX, CHANNEL_SPLIT_REGEX, MENTION_SPLIT_REGEX, ROLE_SPLIT_REGEX}
	ultimateRegex = "(" + strings.Join(regexes, ")|(") + ")"
)

type DiscordMessageParser struct {
	client           *dg.Session
	compiledUltRegex *regexp.Regexp
}

func getEmojiInfo(emojiStr string) Emoji {

	items := regexp.MustCompile(EMOJI_SPLIT_REGEX).FindStringSubmatch(emojiStr)
	if items == nil {
		return Emoji{}
	}

	id := items[3]
	alt := items[2]

	if items[1] == "" {
		id = fmt.Sprintf("https://cdn.discordapp.com/emojis/%s.gif?v=1", id)
	} else {
		id = fmt.Sprintf("https://cdn.discordapp.com/emojis/%s.png", id)

	}

	return Emoji{
		Id:  id,
		Alt: alt,
	}

}

func getIdWithRegex(content string, r *regexp.Regexp) string {
	items := r.FindStringSubmatch(content)
	if items == nil {
		return ""
	}
	return items[1]
}

func (parser *DiscordMessageParser) parsingColoredContent(message *dg.Message, content string, matchedRegex RegexIndex) MessagePart {
	switch matchedRegex {
	case EmojiRegex:
		emoji := getEmojiInfo(content)
		return MessagePart{
			Emoji: emoji,
		}
	case ChannelRegex:
		channelId := getIdWithRegex(content, regexp.MustCompile(splitRegexes[matchedRegex]))
		defaultResult := MessagePart{
			Content: "#unknown-channel",
			Format: Format{
				Color: "#ffffff",
			},
		}
		channel, err := parser.client.Channel(channelId)
		if err != nil {
			fmt.Println(err.Error())
			return defaultResult
		}
		return MessagePart{
			Content: fmt.Sprintf("#%s", channel.Name),
			Format: Format{
				Color: "#ffffff",
			},
		}
	case MentionRegex:
		userId := getIdWithRegex(content, regexp.MustCompile(splitRegexes[matchedRegex]))
		defaultResult := MessagePart{
			Content: "@unknown-user",
			Format: Format{
				Color: "#ffffff",
			},
		}

		member, err := parser.client.State.Member(message.GuildID, userId)
		if err != nil {
			fmt.Println(err.Error())
			return defaultResult
		}

		color := parser.client.State.UserColor(userId, message.ChannelID)
		if err != nil {
			fmt.Println(err.Error())
			return defaultResult
		}

		var username string
		if member.Nick == "" {
			username = member.User.Username
		} else {
			username = member.Nick
		}

		return MessagePart{
			Content: fmt.Sprintf("@%s", username),
			Format: Format{
				Color: fmt.Sprintf("#%06x", color),
			},
		}
	case RoleRegex:
		roleId := getIdWithRegex(content, regexp.MustCompile(splitRegexes[matchedRegex]))
		defaultResult := MessagePart{
			Content: "@unknown-role",
			Format: Format{
				Color: "#ffffff",
			},
		}
		guildRoles, err := parser.client.GuildRoles(message.GuildID)
		if err != nil {
			fmt.Println(err.Error())
			return defaultResult
		}
		idx := slices.IndexFunc(guildRoles, func(role *dg.Role) bool {
			return role.ID == roleId
		})
		if idx == -1 {
			fmt.Println("No role found!")
			return defaultResult
		}
		return MessagePart{
			Content: fmt.Sprintf("@%s", guildRoles[idx].Name),
			Format: Format{
				Color: fmt.Sprintf("#%06x", guildRoles[idx].Color),
			},
		}
	case Everyone:
	case Here:
		return MessagePart{
			Content: content,
			Format: Format{
				Color: "#ffffff",
			},
		}
	default:
	}
	return MessagePart{}
}

func NewParser(client *dg.Session) DiscordMessageParser {

	compiledUltRegex := regexp.MustCompile(ultimateRegex)

	return DiscordMessageParser{
		client,
		compiledUltRegex,
	}
}

func (parser *DiscordMessageParser) ParseMessage(message *dg.Message) []MessagePart {

	msgContent := message.Content
	delimiters := parser.compiledUltRegex.FindAllStringSubmatch(msgContent, -1)

	contents := parser.compiledUltRegex.Split(msgContent, -1)
	var messageParts []MessagePart
	for i := range contents {
		if i > 0 && len(delimiters) > i-1 {

			content := delimiters[i-1][0]
			matchedRegexIdx := slices.IndexFunc(delimiters[i-1][1:], func(match string) bool {
				return match != ""
			})

			messageParts = append(
				messageParts,
				parser.parsingColoredContent(message, content, RegexIndex(matchedRegexIdx)),
			)
		}
		if contents[i] != "" {
			messageParts = append(messageParts, MessagePart{
				Content: contents[i],
			})
		}
	}
	return messageParts
}
