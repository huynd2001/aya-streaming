package main

import (
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
	"syscall"
)

func main() {

	server := &http.Server{
		Addr: ":8000",
	}

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	fmt.Println("Hello, world!")

	testEmitter := test_source.NewEmitter()

	discordToken := os.Getenv("DISCORD_TOKEN")
	discordEmitter, err := discordsource.NewEmitter(discordToken)

	if err != nil {
		fmt.Printf("Error during creating a discord emitter:%s\n", err.Error())
		return
	}

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	wsServer, err := socket.NewWSServer()
	if err != nil {
		fmt.Printf("Errpr during create a web socker: %s\n", err.Error())
	}

	fmt.Println("Ready to send messages through web sockets!")

	go func() {
		for {
			select {
			case testMsg := <-testEmitter.UpdateEmitter():
				fmt.Println("Message from test source!")
				fmt.Printf("%#v\n", testMsg)
				for key, value := range wsServer.ChanMap {
					fmt.Printf("Sent message to channel %s\n", key)
					*value <- testMsg
				}
			case discordMsg := <-discordEmitter.UpdateEmitter():
				fmt.Println("Message from discord!")
				fmt.Printf("%#v\n", discordMsg)
			case _ = <-sc:
				fmt.Println("End Server!")
				if err := server.Close(); err != nil {
					fmt.Printf("Error when closing server: %s\n", err.Error())
				}
				if err := discordEmitter.DiscordClient.Close(); err != nil {
					fmt.Printf("Error when closing discord: %s\n", err.Error())
				}
				return
			}
		}

	}()

	if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("HTTP server error: %s\n", err.Error())
	}
}
