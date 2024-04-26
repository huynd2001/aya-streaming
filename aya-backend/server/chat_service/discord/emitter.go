package discord_source

import (
	"aya-backend/server/chat_service"
	"fmt"
	dg "github.com/bwmarrin/discordgo"
	"github.com/fatih/color"
	"sync"
)

type DiscordEmitter struct {
	chat_service.ChatEmitter
	mutex         sync.Mutex
	updateEmitter chan chat_service.MessageUpdate
	discordClient *dg.Session
	register      *discordRegister

	resource2Subscriber map[string]map[string]bool
}

func (emitter *DiscordEmitter) Register(subscriber string, resourceInfo any) {
	discordInfo, ok := resourceInfo.(DiscordInfo)
	if !ok {
		return
	}
	emitter.mutex.Lock()
	defer emitter.mutex.Unlock()
	guildId := discordInfo.DiscordGuildId
	channelId := discordInfo.DiscordChannelId
	guildChannel := fmt.Sprintf("%s/%s", guildId, channelId)
	if emitter.resource2Subscriber[guildChannel] == nil {
		emitter.resource2Subscriber[guildChannel] = make(map[string]bool)
		emitter.resource2Subscriber[guildChannel][subscriber] = true
		emitter.register.register(guildId, channelId)
	} else {
		emitter.resource2Subscriber[guildChannel][subscriber] = true
	}
}

func (emitter *DiscordEmitter) Deregister(subscriber string, resourceInfo any) {
	discordInfo, ok := resourceInfo.(DiscordInfo)
	if !ok {
		return
	}
	emitter.mutex.Lock()
	defer emitter.mutex.Unlock()
	guildId := discordInfo.DiscordGuildId
	channelId := discordInfo.DiscordChannelId
	guildChannel := fmt.Sprintf("%s/%s", guildId, channelId)
	if emitter.resource2Subscriber[guildChannel] == nil {
		// ignore since there is no resource to deregister
		return
	}
	delete(emitter.resource2Subscriber[guildChannel], subscriber)
	if len(emitter.resource2Subscriber[guildChannel]) == 0 {
		delete(emitter.resource2Subscriber, guildChannel)
		emitter.register.deregister(guildId, channelId)
	}
}

func (emitter *DiscordEmitter) UpdateEmitter() chan chat_service.MessageUpdate {
	return emitter.updateEmitter
}

func (emitter *DiscordEmitter) CloseEmitter() error {
	close(emitter.updateEmitter)
	return emitter.discordClient.Close()
}

func NewEmitter(token string) (*DiscordEmitter, error) {
	messageUpdates := make(chan chat_service.MessageUpdate)

	client, err := dg.New("Bot " + token)
	if err != nil {
		return nil, err
	}

	discordEmitter := DiscordEmitter{
		discordClient:       client,
		updateEmitter:       messageUpdates,
		register:            newDiscordRegister(),
		resource2Subscriber: make(map[string]map[string]bool),
	}

	discordMsgParser := NewParser(client)

	client.Identify.Intents = dg.IntentsAll

	client.AddHandler(func(s *dg.Session, m *dg.MessageCreate) {
		if discordEmitter.register.check(m.GuildID, m.ChannelID) {

			messageUpdates <- chat_service.MessageUpdate{
				UpdateTime: m.Timestamp,
				Update:     chat_service.New,
				Message: chat_service.Message{
					Source:       chat_service.Discord,
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
		if discordEmitter.register.check(m.GuildID, m.ChannelID) {

			messageUpdates <- chat_service.MessageUpdate{
				UpdateTime: m.Timestamp,
				Update:     chat_service.Delete,
				Message: chat_service.Message{
					Source: chat_service.Discord,
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
		if discordEmitter.register.check(m.GuildID, m.ChannelID) {

			messageUpdates <- chat_service.MessageUpdate{
				UpdateTime: m.Timestamp,
				Update:     chat_service.Edit,
				Message: chat_service.Message{
					Source:       chat_service.Discord,
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

	color.Green("New Discord Emitter created!\n")
	return &discordEmitter, nil
}
