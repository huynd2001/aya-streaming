package main

import (
	"aya-backend/server/service"
	discordsource "aya-backend/server/service/discord"
	"aya-backend/server/service/test_source"
	youtubesource "aya-backend/server/service/youtube"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"os"
)

const (
	YOUTUBE_API_KEY_ENV           = "YOUTUBE_API_KEY"
	YOUTUBE_CLIENT_ID_ENV         = "YOUTUBE_CLIENT_ID"
	YOUTUBE_CLIENT_SECRET_ENV     = "YOUTUBE_CLIENT_SECRET"
	YOUTUBE_FLOW_ENV              = "YOUTUBE_FLOW"
	YOUTUBE_BOT_ACCOUNT_EMAIL_ENV = "YOUTUBE_BOT_ACCOUNT_EMAIL"

	DISCORD_TOKEN_ENV = "DISCORD_TOKEN"
)

type MessagesChannel struct {
	service.ChatEmitter
	discordEmitter *discordsource.DiscordEmitter
	testEmitter    *test_source.TestEmitter
	youtubeEmitter *youtubesource.YoutubeEmitter

	updateEmitter chan service.MessageUpdate
}

type MessageChannelConfig struct {
	Discord bool
	Test    bool
	Youtube bool
	BaseURL string
	Router  *mux.Router
}

func (messageChannel MessagesChannel) UpdateEmitter() chan service.MessageUpdate {
	return messageChannel.updateEmitter
}

func (messageChannel MessagesChannel) CloseEmitter() error {

	close(messageChannel.updateEmitter)

	var testError error = nil
	var discordError error = nil
	var youtubeError error = nil

	if messageChannel.testEmitter != nil {
		testError = messageChannel.testEmitter.CloseEmitter()
	}

	if messageChannel.discordEmitter != nil {
		discordError = messageChannel.discordEmitter.CloseEmitter()
	}

	if messageChannel.youtubeEmitter != nil {
		testError = messageChannel.youtubeEmitter.CloseEmitter()
	}

	err := errors.Join(testError, discordError, youtubeError)

	if err != nil {
		return fmt.Errorf("error encounter during closing: %w", err)
	} else {
		return nil
	}
}

func NewMessageChannel(messageChannelConfig *MessageChannelConfig) *MessagesChannel {

	messageChannel := MessagesChannel{
		testEmitter:    nil,
		discordEmitter: nil,
		youtubeEmitter: nil,
	}

	if messageChannelConfig.Discord {
		discordToken := os.Getenv(DISCORD_TOKEN_ENV)
		discordEmitter, err := discordsource.NewEmitter(discordToken)

		if err != nil {
			fmt.Printf("Error during creating a discord emitter: %s\n", err.Error())
		}

		messageChannel.discordEmitter = discordEmitter
	}

	if messageChannelConfig.Test {
		testEmitter := test_source.NewEmitter()
		messageChannel.testEmitter = testEmitter
	}

	if messageChannelConfig.Youtube {
		ytApiKey := os.Getenv(YOUTUBE_API_KEY_ENV)
		ytClientId := os.Getenv(YOUTUBE_CLIENT_ID_ENV)
		ytClientSecret := os.Getenv(YOUTUBE_CLIENT_SECRET_ENV)
		ytFlow := os.Getenv(YOUTUBE_FLOW_ENV)

		ytEmitterConfig := &youtubesource.YoutubeEmitterConfig{}

		switch ytFlow {
		case "api":
			ytEmitterConfig.UseApiKey = true
			ytEmitterConfig.ApiKey = ytApiKey
		case "oauth":
			ytEmitterConfig.UseOAuth = true
			ytEmitterConfig.ClientID = ytClientId
			ytEmitterConfig.ClientSecret = ytClientSecret
			ytEmitterConfig.Router = messageChannelConfig.Router.PathPrefix("/auth").Subrouter()
			ytEmitterConfig.RedirectBasedUrl = fmt.Sprintf("%s/auth", messageChannelConfig.BaseURL)
		}

		youtubeEmitter, err := youtubesource.NewEmitter(ytEmitterConfig)
		if err != nil {
			fmt.Printf("Error during creating a youtube emitter: %s\n", err.Error())
		} else {
			messageChannel.youtubeEmitter = youtubeEmitter
		}
	}

	msgC := make(chan service.MessageUpdate)

	if messageChannel.testEmitter != nil {
		go func() {
			for {
				testMsg := <-messageChannel.testEmitter.UpdateEmitter()
				fmt.Println("Message from test source!")
				msgC <- testMsg
			}
		}()
	}

	if messageChannel.discordEmitter != nil {
		go func() {
			// TODO: get from db server
			guildID1 := os.Getenv("TEST_GUILD_ID_1")
			guildID2 := os.Getenv("TEST_GUILD_ID_2")
			channelID1 := os.Getenv("TEST_CHANNEL_ID_1")
			channelID2 := os.Getenv("TEST_CHANNEL_ID_2")
			messageChannel.discordEmitter.RegisterGuildChannel(guildID1, channelID1)
			messageChannel.discordEmitter.RegisterGuildChannel(guildID2, channelID2)
			for {
				discordMsg := <-messageChannel.discordEmitter.UpdateEmitter()
				fmt.Println("Message from discord!")
				msgC <- discordMsg
			}
		}()
	}

	if messageChannel.youtubeEmitter != nil {
		go func() {
			// TODO: get from db server
			ytChannelID1 := os.Getenv("TEST_YT_CHANNEL_ID_1")
			ytChannelID2 := os.Getenv("TEST_YT_CHANNEL_ID_2")
			messageChannel.youtubeEmitter.RegisterChannel(ytChannelID1)
			messageChannel.youtubeEmitter.RegisterChannel(ytChannelID2)
			for {
				select {
				case ytMsg := <-messageChannel.youtubeEmitter.UpdateEmitter():
					fmt.Println("Message from youtube!")
					msgC <- ytMsg
				case err := <-messageChannel.youtubeEmitter.ErrorEmitter():
					fmt.Printf("Error from youtube:%s\n", err.Error())
					break
				}

			}
		}()
	}
	messageChannel.updateEmitter = msgC

	return &messageChannel
}
