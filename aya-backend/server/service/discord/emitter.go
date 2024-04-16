package discord_source

import (
	"aya-backend/server/service"
	"fmt"
	dg "github.com/bwmarrin/discordgo"
)

type DiscordInfo struct {
	DiscordGuildId   string
	DiscordChannelId string
}

type DiscordEmitter struct {
	service.ChatEmitter
	updateEmitter chan service.MessageUpdate
	discordClient *dg.Session
	register      *discordRegister
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

	discordEmitter := DiscordEmitter{
		discordClient: client,
		updateEmitter: messageUpdates,
		register:      newDiscordRegister(),
	}

	discordMsgParser := NewParser(client)

	client.Identify.Intents = dg.IntentsAll

	client.AddHandler(func(s *dg.Session, m *dg.MessageCreate) {
		if discordEmitter.register.Check(m.GuildID, m.ChannelID) {

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
				ExtraFields: DiscordInfo{
					DiscordGuildId:   m.GuildID,
					DiscordChannelId: m.ChannelID,
				},
			}
		}
	})

	client.AddHandler(func(s *dg.Session, m *dg.MessageDelete) {
		if discordEmitter.register.Check(m.GuildID, m.ChannelID) {

			messageUpdates <- service.MessageUpdate{
				UpdateTime: m.Timestamp,
				Update:     service.Delete,
				Message: service.Message{
					Source: service.Discord,
					Id:     m.ID,
				},
				ExtraFields: DiscordInfo{
					DiscordGuildId:   m.GuildID,
					DiscordChannelId: m.ChannelID,
				},
			}
		}
	})

	client.AddHandler(func(s *dg.Session, m *dg.MessageUpdate) {
		if discordEmitter.register.Check(m.GuildID, m.ChannelID) {

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
				ExtraFields: DiscordInfo{
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
	return &discordEmitter, nil
}
