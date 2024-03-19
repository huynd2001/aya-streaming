package discord_source

import (
	"fmt"
	dg "github.com/bwmarrin/discordgo"
	"os"
	"streemly-backend/service"
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
			messageUpdates <- service.MessageUpdate{
				Source: service.Discord,
				Update: service.New,
				Message: service.Message{
					Id: m.ID,
					Author: service.Author{
						Username: m.Author.Username,
						IsAdmin:  false,
						IsBot:    false,
						Color:    "",
					},
					Content: discordMsgParser.ParseMessage(m.Message),
				},
			}
		}
	})

	// Open a websocket connection to Discord and begin listening.
	err = client.Open()
	if err != nil {
		panic(err)
	}

	// Cleanly close down the Discord session.
	//_ = client.Close()

	return &DiscordEmitter{
		DiscordClient: client,
		updateEmitter: messageUpdates,
	}
}
