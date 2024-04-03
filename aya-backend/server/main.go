package main

import (
	"aya-backend/server/api"
	. "aya-backend/server/service"
	"aya-backend/server/socket"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

const (
	SOURCES_ENV = "SOURCES"
)

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

	r := mux.NewRouter()
	streamRouter := r.PathPrefix("/stream").Subrouter()

	wsServer, err := socket.NewWSServer(streamRouter)
	if err != nil {
		fmt.Printf("Error during create the websocket server: %s\n", err.Error())
	}

	apiRouter := r.PathPrefix("/api").Subrouter()
	api.NewApiServer(apiRouter)

	http.Handle("/", r)
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
