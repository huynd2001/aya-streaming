package discord_source

import (
	"aya-backend/service"
	"fmt"
	dg "github.com/bwmarrin/discordgo"
	"os"
)

type DiscordEmitter struct {
	service.ChatEmitter
	updateEmitter chan service.MessageUpdate
	discordClient *dg.Session
}

func (discordEmitter *DiscordEmitter) UpdateEmitter() chan service.MessageUpdate {
	return discordEmitter.updateEmitter
}

func (discordEmitter *DiscordEmitter) CloseEmitter() error {
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

			color := client.State.UserColor(m.Author.ID, m.ChannelID)

			user, err := client.User(m.Author.ID)
			if err != nil {
				fmt.Println(err.Error())
			}
			userPerm, err := client.State.UserChannelPermissions(m.Author.ID, m.ChannelID)
			if err != nil {
				fmt.Println(err.Error())
			}

			messageUpdates <- service.MessageUpdate{
				Update: service.New,
				Message: service.Message{
					Source: service.Discord,
					Id:     m.ID,
					Author: service.Author{
						Username: m.Author.Username,
						IsAdmin:  (userPerm & dg.PermissionAdministrator) != 0,
						IsBot:    user.Bot,
						Color:    fmt.Sprintf("#%06x", color),
					},
					MessageParts: discordMsgParser.ParseMessage(m.Message),
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
