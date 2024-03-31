package main

import (
	. "aya-backend/service"
	"aya-backend/socket"
	"errors"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

const (
	YOUTUBE_API_KEY_ENV = "YOUTUBE_API_KEY"
	DISCORD_TOKEN_ENV   = "DISCORD_TOKEN"
	SOURCES_ENV         = "SOURCES"
)

func sendMessage(chanMap map[string]*chan MessageUpdate, msg MessageUpdate) {
	for key, value := range chanMap {
		fmt.Printf("Sent message to channel %s\n", key)
		*value <- msg
	}
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
