package main

import (
	. "aya-backend/service"
	discordsource "aya-backend/service/discord"
	"aya-backend/service/test_source"
	"errors"
	"fmt"
	"os"
	"strings"
)

type MessagesChannel struct {
	ChatEmitter
	discordEmitter *discordsource.DiscordEmitter
	testEmitter    *test_source.TestEmitter
}

type MessageChannelConfig struct {
	Discord bool
	Test    bool
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
	return msgC
}

func (messageChannel MessagesChannel) CloseEmitter() error {

	var testError error = nil
	var discordError error = nil

	if messageChannel.testEmitter != nil {
		testError = messageChannel.testEmitter.CloseEmitter()
	}

	if messageChannel.discordEmitter != nil {
		discordError = messageChannel.discordEmitter.CloseEmitter()
	}

	err := errors.Join(testError, discordError)

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
	}

	if settings.Discord {
		discordToken := os.Getenv(DISCORD_TOKEN_ENV)
		discordEmitter, err := discordsource.NewEmitter(discordToken)

		if err != nil {
			fmt.Printf("Error during creating a discord emitter:%s\n", err.Error())
		}

		msgChannel.discordEmitter = discordEmitter
	}

	if settings.Test {
		testEmitter := test_source.NewEmitter()
		msgChannel.testEmitter = testEmitter
	}
	return &msgChannel
}

func parseConfig(msgSettingStr string) *MessageChannelConfig {

	config := MessageChannelConfig{
		Test:    false,
		Discord: false,
	}
	enabledSources := strings.Split(msgSettingStr, " ")
	for _, enabledSource := range enabledSources {
		switch enabledSource {
		case "test_source":
			config.Test = true
			break
		case "discord":
			config.Discord = true
			break
		}
	}
	return &config
}
