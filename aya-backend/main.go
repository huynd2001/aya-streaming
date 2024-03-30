package main

import (
	. "aya-backend/service"
	discordsource "aya-backend/service/discord"
	"aya-backend/service/test_source"
	"aya-backend/socket"
	"errors"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

const (
	DISCORD_TOKEN_ENV = "DISCORD_TOKEN"
	SOURCES_ENV       = "SOURCES"
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

func sendMessage(chanMap map[string]*chan MessageUpdate, msg MessageUpdate) {
	for key, value := range chanMap {
		fmt.Printf("Sent message to channel %s\n", key)
		*value <- msg
	}
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

func main() {

	server := &http.Server{
		Addr: ":8000",
	}

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	fmt.Println("Hello, world!")

	enabledSourceStr := os.Getenv(SOURCES_ENV)
	msgChanConfig := parseConfig(enabledSourceStr)

	msgChanEmitter := NewMessageChannel(msgChanConfig)

	msgC := msgChanEmitter.UpdateEmitter()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	wsServer, err := socket.NewWSServer()
	if err != nil {
		fmt.Printf("Error during create a web socker: %s\n", err.Error())
	}

	fmt.Println("Ready to send messages through web sockets!")

	go func() {
		for {
			select {
			case msg := <-msgC:
				fmt.Printf("%#v\n", msg)
				sendMessage(wsServer.ChanMap, msg)
			case _ = <-sc:
				fmt.Println("End Server!")
				if err := server.Close(); err != nil {
					fmt.Printf("Error when closing server: %s\n", err.Error())
				}
				if err := msgChanEmitter.CloseEmitter(); err != nil {
					fmt.Printf("%s\n", err.Error())
				}
				return
			}
		}

	}()

	if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("HTTP server error: %s\n", err.Error())
	}
}
