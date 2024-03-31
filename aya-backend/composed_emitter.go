package main

import (
	. "aya-backend/service"
	discordsource "aya-backend/service/discord"
	"aya-backend/service/test_source"
	youtube_source "aya-backend/service/youtube"
	"errors"
	"fmt"
	"os"
	"strings"
)

type MessagesChannel struct {
	ChatEmitter
	discordEmitter *discordsource.DiscordEmitter
	testEmitter    *test_source.TestEmitter
	youtubeEmitter *youtube_source.YoutubeEmitter
}

type MessageChannelConfig struct {
	Discord bool
	Test    bool
	Youtube bool
}

func (messageChannel MessagesChannel) UpdateEmitter() chan MessageUpdate {
	msgC := make(chan MessageUpdate)

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
		ytEmitterConfig := youtube_source.YoutubeEmitterConfig{
			ApiKey: ytApiKey,
		}
		youtubeEmitter, err := youtube_source.NewEmitter(&ytEmitterConfig)
		if err != nil {
			fmt.Printf("Error during creating a youtube emitter: %s\n", err.Error())
		}
		msgChannel.youtubeEmitter = youtubeEmitter
	}

	return &msgChannel
}

func parseConfig(msgSettingStr string) *MessageChannelConfig {

	config := MessageChannelConfig{
		Test:    false,
		Discord: false,
		Youtube: false,
	}
	enabledSources := strings.Split(msgSettingStr, " ")
	for _, enabledSource := range enabledSources {
		switch enabledSource {
		case "test_source":
			config.Test = true
		case "discord":
			config.Discord = true
		case "youtube":
			config.Youtube = true
		}

	}
	return &config
}
