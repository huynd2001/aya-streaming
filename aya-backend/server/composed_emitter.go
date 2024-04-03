package main

import (
	"aya-backend/server/service"
	discordsource "aya-backend/server/service/discord"
	"aya-backend/server/service/test_source"
	youtubesource "aya-backend/server/service/youtube"
	"errors"
	"fmt"
	"os"
)

const (
	YOUTUBE_API_KEY_ENV = "YOUTUBE_API_KEY"
	DISCORD_TOKEN_ENV   = "DISCORD_TOKEN"
)

type MessagesChannel struct {
	service.ChatEmitter
	discordEmitter *discordsource.DiscordEmitter
	testEmitter    *test_source.TestEmitter
	youtubeEmitter *youtubesource.YoutubeEmitter
}

type MessageChannelConfig struct {
	Discord bool
	Test    bool
	Youtube bool
}

func (messageChannel MessagesChannel) UpdateEmitter() chan service.MessageUpdate {
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
			for {
				discordMsg := <-messageChannel.discordEmitter.UpdateEmitter()
				fmt.Println("Message from discord!")
				msgC <- discordMsg
			}
		}()
	}

	if messageChannel.youtubeEmitter != nil {
		go func() {
			for {
				ytMsg := <-messageChannel.youtubeEmitter.UpdateEmitter()
				fmt.Println("Message from youtube!")
				msgC <- ytMsg
			}
		}()
	}
	return msgC
}

func (messageChannel MessagesChannel) CloseEmitter() error {

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

func NewMessageChannel(settings *MessageChannelConfig) *MessagesChannel {

	msgChannel := MessagesChannel{
		testEmitter:    nil,
		discordEmitter: nil,
		youtubeEmitter: nil,
	}

	if settings.Discord {
		discordToken := os.Getenv(DISCORD_TOKEN_ENV)
		discordEmitter, err := discordsource.NewEmitter(discordToken)

		if err != nil {
			fmt.Printf("Error during creating a discord emitter: %s\n", err.Error())
		}

		msgChannel.discordEmitter = discordEmitter
	}

	if settings.Test {
		testEmitter := test_source.NewEmitter()
		msgChannel.testEmitter = testEmitter
	}

	if settings.Youtube {
		ytApiKey := os.Getenv(YOUTUBE_API_KEY_ENV)
		ytEmitterConfig := youtubesource.YoutubeEmitterConfig{
			ApiKey: ytApiKey,
		}
		youtubeEmitter, err := youtubesource.NewEmitter(&ytEmitterConfig)
		if err != nil {
			fmt.Printf("Error during creating a youtube emitter: %s\n", err.Error())
		}
		msgChannel.youtubeEmitter = youtubeEmitter
	}

	return &msgChannel
}
