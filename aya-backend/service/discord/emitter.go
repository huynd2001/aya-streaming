package discord_source

import (
	"aya-backend/service"
	"fmt"
	dg "github.com/bwmarrin/discordgo"
	"os"
)

type DiscordEmitter struct {
	updateEmitter chan service.MessageUpdate
	DiscordClient *dg.Session
}

func (discordEmitter *DiscordEmitter) UpdateEmitter() chan service.MessageUpdate {
	return discordEmitter.updateEmitter
}

func NewEmitter(token string) *DiscordEmitter {
	messageUpdates := make(chan service.MessageUpdate)

	client, err := dg.New("Bot " + token)
	if err != nil {
		panic(err)
	}

	guildId := os.Getenv("TEST_GUILD_ID")
	if guildId == "" {
		panic(fmt.Errorf("err"))
	}

	channelId := os.Getenv("TEST_CHANNEL_ID")
	if channelId == "" {
		panic(fmt.Errorf("err"))
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
				Source: service.Discord,
				Update: service.New,
				Message: service.Message{
					Id: m.ID,
					Author: service.Author{
						Username: m.Author.Username,
						IsAdmin:  (userPerm & dg.PermissionAdministrator) != 0,
						IsBot:    user.Bot,
						Color:    fmt.Sprintf("#%06x", color),
					},
					Content: discordMsgParser.ParseMessage(m.Message),
				},
			}
		}
	})

	err = client.Open()
	if err != nil {
		panic(err)
	}

	return &DiscordEmitter{
		DiscordClient: client,
		updateEmitter: messageUpdates,
	}
}
