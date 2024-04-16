package discord_source

import (
	"aya-backend/server/service"
	"fmt"
	dg "github.com/bwmarrin/discordgo"
	"os"
)

type DiscordSpecificInfo struct {
	DiscordGuildId   string
	DiscordChannelId string
}

type DiscordEmitter struct {
	service.ChatEmitter
	updateEmitter chan service.MessageUpdate
	discordClient *dg.Session
}

func (discordEmitter *DiscordEmitter) UpdateEmitter() chan service.MessageUpdate {
	return discordEmitter.updateEmitter
}

func (discordEmitter *DiscordEmitter) CloseEmitter() error {
	close(discordEmitter.updateEmitter)
	return discordEmitter.discordClient.Close()
}

func NewEmitter(token string) (*DiscordEmitter, error) {
	messageUpdates := make(chan service.MessageUpdate)

	client, err := dg.New("Bot " + token)
	if err != nil {
		return nil, err
	}

	// TODO: work with database to retrieve the guild id for system

	guildId := os.Getenv("TEST_GUILD_ID")
	if guildId == "" {
		return nil, fmt.Errorf("no Guild specified")
	}

	channelId := os.Getenv("TEST_CHANNEL_ID")
	if channelId == "" {
		return nil, fmt.Errorf("no Channel specified")
	}

	discordMsgParser := NewParser(client)

	client.Identify.Intents = dg.IntentsAll

	client.AddHandler(func(s *dg.Session, m *dg.MessageCreate) {
		if m.GuildID == guildId && m.ChannelID == channelId {

			messageUpdates <- service.MessageUpdate{
				UpdateTime: m.Timestamp,
				Update:     service.New,
				Message: service.Message{
					Source:       service.Discord,
					Id:           m.ID,
					Author:       discordMsgParser.ParseAuthor(m.Author, m.ChannelID),
					MessageParts: discordMsgParser.ParseMessage(m.Message),
					Attachments:  discordMsgParser.ParseAttachment(m.Message),
				},
				ExtraFields: DiscordSpecificInfo{
					DiscordGuildId:   m.GuildID,
					DiscordChannelId: m.ChannelID,
				},
			}
		}
	})

	client.AddHandler(func(s *dg.Session, m *dg.MessageDelete) {
		if m.GuildID == guildId && m.ChannelID == channelId {

			messageUpdates <- service.MessageUpdate{
				UpdateTime: m.Timestamp,
				Update:     service.Delete,
				Message: service.Message{
					Source: service.Discord,
					Id:     m.ID,
				},
				ExtraFields: DiscordSpecificInfo{
					DiscordGuildId:   m.GuildID,
					DiscordChannelId: m.ChannelID,
				},
			}
		}
	})

	client.AddHandler(func(s *dg.Session, m *dg.MessageUpdate) {
		if m.GuildID == guildId && m.ChannelID == channelId {

			messageUpdates <- service.MessageUpdate{
				UpdateTime: m.Timestamp,
				Update:     service.Edit,
				Message: service.Message{
					Source:       service.Discord,
					Id:           m.ID,
					Author:       discordMsgParser.ParseAuthor(m.Author, m.ChannelID),
					MessageParts: discordMsgParser.ParseMessage(m.Message),
					Attachments:  discordMsgParser.ParseAttachment(m.Message),
				},
				ExtraFields: DiscordSpecificInfo{
					DiscordGuildId:   m.GuildID,
					DiscordChannelId: m.ChannelID,
				},
			}
		}
	})

	err = client.Open()
	if err != nil {
		return nil, err
	}

	fmt.Printf("New Discord Emitter created!\n")

	return &DiscordEmitter{
		discordClient: client,
		updateEmitter: messageUpdates,
	}, nil
}
